package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	PATH_ACCESSLOG string = "access.log"
	PATH_ERRORLOG  string = "error.log"
)

func main() {
	http.HandleFunc("/{resource...}", get)
	log.Println("Server started")
	http.ListenAndServe(":8080", nil)
}

func get(writer http.ResponseWriter, req *http.Request) {
	fetchedData, fetchErr := fetchFileContents(req.URL.Path)
	respStatusCode := int(200)
	var sentBytes int = 0
	if logIfError(fetchErr) {
		respStatusCode = http.StatusNotFound
		writer.WriteHeader(respStatusCode)
	} else {
		sentBytesCount, _ := writer.Write(fetchedData)
		sentBytes = sentBytesCount
	}
	logAccess(strings.Split(req.RemoteAddr, ":")[0], "-", getOrDash(req.URL.User.Username()), time.Now().Format("02/Jan/2006:03:04:05 -0700"),
		req.Method+" "+req.URL.Path+" "+getHttpString(req.ProtoMajor, req.ProtoMinor), strconv.Itoa(respStatusCode),
		strconv.Itoa(sentBytes),
	)
}

func getHttpString(major, minor int) string {
	return "HTTP/" + strconv.Itoa(major) + "." + strconv.Itoa(minor)
}
