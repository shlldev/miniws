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

func NewWebServer(port_ int, logFolder_, configFolder_, wwwFolder_ string) *WebServer {

	return &WebServer{
		logger:              NewLogger(logFolder_),
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

func (ws *WebServer) parseFilterPanics(filename string) (FilterMode, []string) {

	filterMode := FILTER_MODE_BLACKLIST
	filter := make([]string, 0)

	os.Mkdir(ws.configFolder, PERMS_MKDIR)
	fileinfo, err := os.Stat(ensureSlashSuffix(ws.configFolder) + filename)

	if errors.Is(err, os.ErrNotExist) {
		os.Create(ensureSlashSuffix(ws.configFolder) + filename)
		fileinfo, err = os.Stat(ensureSlashSuffix(ws.configFolder) + filename)
	}

	if err != nil {
		panic("Error opening " + filename + ": " + err.Error())
	}

	if fileinfo.Size() == 0 { // empty config
		return filterMode, filter
	}

	filterContent, err := os.ReadFile(ensureSlashSuffix(ws.configFolder) + filename)

	if ws.logger.logIfError(err) {
		panic("Error reading " + filename + ": " + err.Error())
	}

	filterLines := strings.Split(string(filterContent), "\n")
	readFilterMode := filterLines[0]

	switch readFilterMode {
	case "allow":
		filterMode = FILTER_MODE_WHITELIST
	case "deny":
		filterMode = FILTER_MODE_BLACKLIST
	default:
		panic("invalid filter mode for " + filename + ": use allow|deny")
	}

	filter = filterLines[1:]

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

func (ws *WebServer) fetchFileContents(filepath string) ([]byte, error) {
	if filepath == "/" {
		filepath = "."
	}
	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	if fileinfo.IsDir() {
		filepath += "/index.html"
	}
	return os.ReadFile(filepath)

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

	fetchedData, fetchErr := ws.fetchFileContents(ensureSlashSuffix(ws.wwwFolder) + strings.TrimPrefix(req.URL.Path, "/"))

	sentBytes := 0

	if ws.logger.logIfError(fetchErr) {
		respStatusCode = http.StatusNotFound
		writer.WriteHeader(respStatusCode)
	} else {
		writer.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(req.URL.Path)))
		sentBytesCount, _ := writer.Write(fetchedData)
		sentBytes = sentBytesCount
	}

	ws.logger.logAccess(
		strings.Split(req.RemoteAddr, ":")[0], //remote address
		"-",                                   //identifier (can't get)
		getOrDash(req.URL.User.Username()),    //username
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),                                      //timestamp
		req.Method+" "+req.URL.Path+" "+getHttpVersionString(req.ProtoMajor, req.ProtoMinor), //HTTP version
		strconv.Itoa(respStatusCode),                                                         //response code
		strconv.Itoa(sentBytes),                                                              //# of sent bytes
		req.Referer(),                                                                        //Referer
		req.UserAgent(),                                                                      //User Agent
	)
}
