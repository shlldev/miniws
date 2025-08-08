package miniws

import (
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	FILTER_MODE_WHITELIST FilterMode = 0
	FILTER_MODE_BLACKLIST FilterMode = 1
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

	_, err := os.Stat(ws.wwwFolder)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalln("Fatal: www folder " + ws.wwwFolder + " does not exist")
	}

	ws.ipFilterMode, ws.ipFilter = ws.parseFilterPanics(FILENAME_IPFILTER)
	ws.userAgentFilterMode, ws.userAgentFilter = ws.parseFilterPanics(FILENAME_USERAGENTFILTER)

	http.HandleFunc("/", ws.get)
	log.Println("Server started on port " + strconv.Itoa(ws.port))
	http.ListenAndServe(":"+strconv.Itoa(ws.port), nil)
}

func (ws *WebServer) parseFilterPanics(fileName string) (FilterMode, []string) {

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
		panic("Error opening " + fileName + ": " + err.Error())
	}

	if fileinfo.Size() == 0 { // empty config
		return filterMode, filter
	}

	filterContent, err := os.ReadFile(fullPath)

	if ws.logger.logIfError(err, fullPath) {
		panic("Error reading " + fileName + ": " + err.Error())
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
		panic("invalid filter mode for " + fileName + ": use allow|deny")
	}

	return filterMode, filter

}

func (ws *WebServer) isIpValid(ip string) bool {
	ip = strings.Split(ip, ":")[0] // remove port
	switch ws.ipFilterMode {
	case FILTER_MODE_WHITELIST:
		return slices.Contains(ws.ipFilter, ip)
	case FILTER_MODE_BLACKLIST:
		return !slices.Contains(ws.ipFilter, ip)
	default:
		return false
	}
}

func (ws *WebServer) isUserAgentValid(userAgent string) bool {
	switch ws.userAgentFilterMode {
	case FILTER_MODE_WHITELIST:
		return slices.Contains(ws.userAgentFilter, userAgent)
	case FILTER_MODE_BLACKLIST:
		return !slices.Contains(ws.userAgentFilter, userAgent)
	default:
		return false
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
