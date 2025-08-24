package miniws

import (
	"bytes"
	"errors"
	"ipc"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wneessen/go-fileperm"
)

const (
	FILTER_MODE_INVALID   FilterMode = -1
	FILTER_MODE_WHITELIST FilterMode = 0
	FILTER_MODE_BLACKLIST FilterMode = 1

	SOCKET_PATH string = "/tmp/miniws_commands_server"
)

type FilterMode int

type WebServer struct {
	logger              *Logger
	port                int
	configFolder        string
	wwwFolder           string
	ipFilter            []string
	userAgentFilter     []string
	ipFilterMode        FilterMode
	userAgentFilterMode FilterMode
}

func NewWebServer(port_ int, logFolder_, configFolder_, wwwFolder_ string, maxLogBytes_ int64) *WebServer {

	return &WebServer{
		logger:              NewLogger(logFolder_, maxLogBytes_),
		port:                port_,
		configFolder:        configFolder_,
		wwwFolder:           wwwFolder_,
		ipFilter:            make([]string, 0),
		userAgentFilter:     make([]string, 0),
		ipFilterMode:        FILTER_MODE_BLACKLIST,
		userAgentFilterMode: FILTER_MODE_BLACKLIST,
	}
}

func (ws *WebServer) Run() {

	_, err := os.Lstat(ws.wwwFolder)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalln("Fatal: www folder " + ws.wwwFolder + " does not exist")
	} else if err != nil {
		log.Fatalln("Fatal: " + err.Error())
	}
	perms, err := fileperm.New(ws.wwwFolder)
	if err != nil {
		log.Fatalln("Fatal: " + err.Error())
	}
	if !perms.UserReadable() {
		log.Fatalln("Fatal: missing permissions to read www folder")
	}

	ipFilterMode, ipFilter, err := ws.parseFilter(FILENAME_IPFILTER)
	if err != nil {
		log.Fatalln("Fatal: IP filter invalid:", err)
	}
	ws.ipFilterMode = ipFilterMode
	ws.ipFilter = ipFilter
	userAgentFilterMode, userAgentFilter, err := ws.parseFilter(FILENAME_USERAGENTFILTER)
	if err != nil {
		log.Fatalln("Fatal: UserAgent filter invalid:", err)
	}
	ws.userAgentFilter = userAgentFilter
	ws.userAgentFilterMode = userAgentFilterMode

	// create and start a unix socket server (to accept signal from another process using -s <cmd>)
	ipcServer := ipc.Server{}
	go ipcServer.Start(ws.onRecieveSignal, "unix", SOCKET_PATH)

	http.HandleFunc("/", ws.get)
	log.Println("Server starting on port " + strconv.Itoa(ws.port) + "...")
	httpErr := http.ListenAndServe(":"+strconv.Itoa(ws.port), nil)
	if httpErr != nil {
		log.Fatalln(httpErr)
	}
}

func (ws *WebServer) onRecieveSignal(command string, arguments []string) bool {
	command = string(bytes.Trim([]byte(command), "\x00"))
	switch command {
	case "reload":
		ws.parseFilter(FILENAME_IPFILTER)
		ws.parseFilter(FILENAME_USERAGENTFILTER)
		return true
	default:
		log.Println("Error: unknown command", command, arguments)
		return false
	}
}

func (ws *WebServer) parseFilter(fileName string) (FilterMode, []string, error) {

	log.Println("loaded filter: ", fileName)

	filterMode := FILTER_MODE_BLACKLIST
	filter := make([]string, 0)

	os.Mkdir(ws.configFolder, PERMS_MKDIR)
	fileinfo, err := os.Stat(filepath.Join(ws.configFolder, fileName))

	fullPath := filepath.Join(ws.configFolder, fileName)
	if errors.Is(err, os.ErrNotExist) {
		os.Create(fullPath)
		fileinfo, err = os.Stat(fullPath)
	}

	if err != nil {
		return FILTER_MODE_INVALID, nil,
			errors.New("Error opening " + fileName + ": " + err.Error())
	}

	if fileinfo.Size() == 0 { // empty config
		return filterMode, filter, nil
	}

	filterContent, err := os.ReadFile(fullPath)

	if ws.logger.logIfError(err, fullPath) {
		return FILTER_MODE_INVALID, nil,
			errors.New("Error reading " + fileName + ": " + err.Error())
	}

	lines := strings.Split(string(filterContent), "\n")
	var linesNoComments []string = make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(strings.Split(line, "#")[0]) // only take portion before comments
		if line == "" {
			continue
		}
		linesNoComments = append(linesNoComments, line)
	}

	readFilterMode := linesNoComments[0]
	filter = linesNoComments[1:]

	switch readFilterMode {
	case "allow":
		filterMode = FILTER_MODE_WHITELIST
	case "deny":
		filterMode = FILTER_MODE_BLACKLIST
	default:
		return FILTER_MODE_INVALID, nil,
			errors.New("invalid filter mode for " + fileName + ": use allow|deny")
	}

	return filterMode, filter, nil

}

func (ws *WebServer) isIpValid(ip string) bool {
	ip = strings.Split(ip, ":")[0] // remove port
	switch ws.ipFilterMode {
	case FILTER_MODE_WHITELIST:
		return slices.Contains(ws.ipFilter, ip)
	case FILTER_MODE_BLACKLIST:
		return !slices.Contains(ws.ipFilter, ip)
	default:
		return false //if something went wrong with conf parsing
	}
}

func (ws *WebServer) isUserAgentValid(userAgent string) bool {
	switch ws.userAgentFilterMode {
	case FILTER_MODE_WHITELIST:
		for _, userAgentFiltered := range ws.userAgentFilter {
			if strings.Contains(userAgent, userAgentFiltered) {
				return true
			}
		}
		return false
	case FILTER_MODE_BLACKLIST:
		for _, userAgentFiltered := range ws.userAgentFilter {
			if strings.Contains(userAgent, userAgentFiltered) {
				return false
			}
		}
		return true
	default:
		return false //if something went wrong with conf parsing
	}
}

// fetchFile is a safe wrapper for os.OpenFile, which sanitizes the
// provided filepath, and, if a folder is passed, it looks for an
// index.html to fetch.
//
// IMPORTANT: remember to close the file after use!!!! fetchFile doesn't
// do it for you for obvious reasons
func (ws *WebServer) fetchFile(filepath string) (*os.File, error) {
	return os.OpenFile(ws._addIndexIfDir(filepath), os.O_RDONLY, 0)
}

func (ws *WebServer) fetchStat(filepath string) (os.FileInfo, error) {
	return os.Stat(ws._addIndexIfDir(filepath))
}

func (ws *WebServer) _addIndexIfDir(filePath string) string {
	fileinfo, err := os.Stat(filePath)
	if ws.logger.logIfError(err, filePath) {
		return ""
	}
	if fileinfo.IsDir() {
		filePath += "/index.html"
	}
	return filePath
}

func (ws *WebServer) get(writer http.ResponseWriter, req *http.Request) {

	// handle OPTION preflight request
	if origin := req.Header.Get("Origin"); origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}

	respStatusCode := http.StatusOK

	if !ws.isIpValid(req.RemoteAddr) || !ws.isUserAgentValid(req.UserAgent()) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	fileToFetch := filepath.Join(ws.wwwFolder, req.URL.Path)
	fetchedFile, fetchErr := ws.fetchFile(fileToFetch)
	fetchedFileStat, _ := fetchedFile.Stat()
	fetchedStat, _ := ws.fetchStat(fileToFetch)

	sentBytes := int64(0)

	if ws.logger.logIfError(fetchErr, fileToFetch) {
		respStatusCode = http.StatusNotFound
		writer.WriteHeader(respStatusCode)
	} else {
		http.ServeContent(writer, req, fileToFetch, fetchedStat.ModTime(), fetchedFile)
		writer.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(req.URL.Path)))
		sentBytes = fetchedFileStat.Size()
		fetchedFile.Close()
	}

	// this thing writes to the log using the NCSA Combined Log Format
	ws.logger.logAccess(
		strings.Split(req.RemoteAddr, ":")[0], //remote address
		"-",                                   //identifier (can't get)
		getOrDash(req.URL.User.Username()),    //username
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),                                      //timestamp
		req.Method+" "+req.URL.Path+" "+getHttpVersionString(req.ProtoMajor, req.ProtoMinor), //HTTP version
		strconv.Itoa(respStatusCode),                                                         //response code
		strconv.Itoa(int(sentBytes)),                                                         //# of sent bytes
		req.Referer(),                                                                        //Referer
		req.UserAgent(),                                                                      //User Agent
	)
}
