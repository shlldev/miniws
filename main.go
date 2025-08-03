package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
)

const (
	PATH_ACCESSLOG string      = "access.log"
	PATH_ERRORLOG  string      = "error.log"
	FLAGS_LOG_OPEN int         = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	PERMS_LOG_OPEN os.FileMode = os.ModeType | os.ModePerm
)

var _logFolder string = ""

func main() {

	parser := argparse.NewParser("miniws", "")
	port := parser.String("p", "port", &argparse.Options{Default: "8040"})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs"})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	_logFolder = *logFolder

	http.HandleFunc("/{resource...}", get)
	log.Println("Server started on port " + *port)
	http.ListenAndServe(":"+*port, nil)
}

func get(writer http.ResponseWriter, req *http.Request) {
	fetchedData, fetchErr := fetchFileContents(req.URL.Path)

	sentBytes := 0
	respStatusCode := http.StatusOK

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
	os.Mkdir(_logFolder, os.ModeDir|os.ModePerm)
	file, err := os.OpenFile(ensureSlashSuffix(_logFolder)+PATH_ACCESSLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log access file at", ensureSlashSuffix(_logFolder)+PATH_ACCESSLOG)
	}
	defer file.Close()
	file.WriteString(out)
}

func logError(str string) {
	os.Mkdir(_logFolder, os.ModeDir|os.ModePerm)
	file, err := os.OpenFile(ensureSlashSuffix(_logFolder)+PATH_ERRORLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log error file at", ensureSlashSuffix(_logFolder)+PATH_ERRORLOG)
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
