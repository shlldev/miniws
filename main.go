package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
)

const (
	PATH_ACCESSLOG        string      = "access.log"
	PATH_ERRORLOG         string      = "error.log"
	PATH_IPFILTER         string      = "ipfilter.conf"
	PATH_USERAGENTFILTER  string      = "useragentfilter.conf"
	FLAGS_LOG_OPEN        int         = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	FLAGS_CONFIG_OPEN     int         = os.O_RDONLY | os.O_CREATE
	PERMS_LOG_OPEN        os.FileMode = os.ModeType | os.ModePerm
	PERMS_CONFIG_OPEN     os.FileMode = os.ModeType | os.ModePerm
	PERMS_MKDIR           os.FileMode = os.ModeDir | os.ModePerm
	FILTER_MODE_WHITELIST             = 0
	FILTER_MODE_BLACKLIST             = 1
)

var (
	logFolder           string   = ""
	configFolder        string   = ""
	ipFilter            []string = make([]string, 0)
	ipFilterMode                 = FILTER_MODE_WHITELIST
	userAgentFilter     []string = make([]string, 0)
	userAgentFilterMode          = FILTER_MODE_WHITELIST
)

func main() {

	parser := argparse.NewParser("miniws", "")
	port := parser.String("p", "port", &argparse.Options{Default: "8040"})
	_logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs"})
	_configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config"})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	logFolder = *_logFolder
	configFolder = *_configFolder

	parseIpFilter()
	parseUserAgentFilter()
	http.HandleFunc("/", get)
	log.Println("Server started on port " + *port)
	http.ListenAndServe(":"+*port, nil)
}

func get(writer http.ResponseWriter, req *http.Request) {

	respStatusCode := http.StatusOK

	if !isIpValid(req.RemoteAddr) || !isUserAgentValid(req.UserAgent()) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	fetchedData, fetchErr := fetchFileContents(req.URL.Path)

	sentBytes := 0

	if logIfError(fetchErr) {
		respStatusCode = http.StatusNotFound
		writer.WriteHeader(respStatusCode)
	} else {
		sentBytesCount, _ := writer.Write(fetchedData)
		sentBytes = sentBytesCount
	}

	logAccess(
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

func fetchFileContents(filepath string) ([]byte, error) {
	if filepath == "/" {
		filepath = "."
	} else {
		filepath_relative, _ := strings.CutPrefix(filepath, "/")
		filepath = filepath_relative
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

// returns error != nil
func logIfError(err error) bool {
	if err != nil {
		logError(err.Error())
		return true
	}
	return false
}

func logAccess(
	remoteAddr, identifier, authuser, timestamp, request,
	status, bytesSent, referer, user_agent string,
) {
	out := fmt.Sprintf("%v %v %v [%v] \"%v\" %v %v \"%v\" \"%v\"\n",
		remoteAddr, identifier, authuser, timestamp, request, status, bytesSent, referer, user_agent,
	)
	os.Mkdir(logFolder, os.ModeDir|os.ModePerm)
	file, err := os.OpenFile(ensureSlashSuffix(logFolder)+PATH_ACCESSLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log access file at", ensureSlashSuffix(logFolder)+PATH_ACCESSLOG)
	}
	defer file.Close()
	file.WriteString(out)
}

func logError(str string) {
	os.Mkdir(logFolder, PERMS_MKDIR)
	file, err := os.OpenFile(ensureSlashSuffix(logFolder)+PATH_ERRORLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log error file at", ensureSlashSuffix(logFolder)+PATH_ERRORLOG)
	}
	defer file.Close()
	file.WriteString(str + "\n")
}

func getHttpVersionString(major, minor int) string {
	return "HTTP/" + strconv.Itoa(major) + "." + strconv.Itoa(minor)
}

func getOrDash(str string) string {
	if str == "" {
		return "-"
	}
	return str
}

func ensureSlashSuffix(str string) string {
	return strings.TrimSuffix(str, "/") + "/"
}

func isIpValid(ip string) bool {
	ip = strings.Split(ip, ":")[0] // remove port
	switch ipFilterMode {
	case FILTER_MODE_WHITELIST:
		return slices.Contains(ipFilter, ip)
	case FILTER_MODE_BLACKLIST:
		return !slices.Contains(ipFilter, ip)
	default:
		return false
	}
}

func isUserAgentValid(userAgent string) bool {
	switch userAgentFilterMode {
	case FILTER_MODE_WHITELIST:
		return slices.Contains(userAgentFilter, userAgent)
	case FILTER_MODE_BLACKLIST:
		return !slices.Contains(userAgentFilter, userAgent)
	default:
		return false
	}
}

func parseIpFilter() {
	os.Mkdir(configFolder, PERMS_MKDIR)
	fileinfo, err := os.Stat(ensureSlashSuffix(configFolder) + PATH_IPFILTER)

	if errors.Is(err, os.ErrNotExist) {
		os.Create(ensureSlashSuffix(configFolder) + PATH_IPFILTER)
	}
	if fileinfo.Size() == 0 { // empty config
		ipFilterMode = FILTER_MODE_BLACKLIST
		ipFilter = make([]string, 0)
		return
	}
	ipFilterContent, err := os.ReadFile(ensureSlashSuffix(configFolder) + PATH_IPFILTER)

	if logIfError(err) {
		return
	}

	ipFilterLines := strings.Split(string(ipFilterContent), "\n")

	_ipFilterMode := ipFilterLines[0]

	switch _ipFilterMode {
	case "allow":
		ipFilterMode = FILTER_MODE_WHITELIST
	case "deny":
		ipFilterMode = FILTER_MODE_BLACKLIST
	default:
		logError("invalid ip filter mode, use allow|deny")
	}

	ipFilter = ipFilterLines[1:]

}

func parseUserAgentFilter() {
	os.Mkdir(configFolder, PERMS_MKDIR)
	fileinfo, err := os.Stat(ensureSlashSuffix(configFolder) + PATH_USERAGENTFILTER)
	if errors.Is(err, os.ErrNotExist) {
		os.Create(ensureSlashSuffix(configFolder) + PATH_USERAGENTFILTER)
	}
	if fileinfo.Size() == 0 { // empty config
		userAgentFilterMode = FILTER_MODE_BLACKLIST
		userAgentFilter = make([]string, 0)
		return
	}

	userAgentFilterContent, err := os.ReadFile(ensureSlashSuffix(configFolder) + PATH_USERAGENTFILTER)

	if logIfError(err) {
		return
	}

	userAgentFilterLines := strings.Split(string(userAgentFilterContent), "\n")
	_userAgentFilterMode := userAgentFilterLines[0]

	println(len(userAgentFilterLines))
	switch _userAgentFilterMode {
	case "allow":
		userAgentFilterMode = FILTER_MODE_WHITELIST
	case "deny":
		userAgentFilterMode = FILTER_MODE_BLACKLIST
	default:
		logError("invalid userAgent filter mode, use allow|deny")
	}
	userAgentFilter = userAgentFilterLines[1:]
}
