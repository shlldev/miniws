package main

import (
	"fmt"
	"log"
	"os"
)

func logIfError(err error) bool {
	if err != nil {
		logError(err.Error())
		return true
	}
	return false
}

// func logAccess(
// 	remoteAddr, remoteUser, timeLocal, request, status,
// 	bodyBytesSent, httpReferrer, httpUseragent string,
// ) {
// 	out := fmt.Sprintf("%v - %v - [%v] \"%v\" %v %v \"%v\" \"%v\"\n",
// 		remoteAddr, remoteUser, timeLocal, request, status, bodyBytesSent,
// 		httpReferrer, httpUseragent,
// 	)
// 	file, err := os.OpenFile(PATH_ACCESSLOG, os.O_APPEND|os.O_WRONLY, os.ModeType)
// 	if err != nil {
// 		log.Println("couldn't open log access file at", PATH_ACCESSLOG)
// 	}
// 	defer file.Close()
// 	file.WriteString(out)
// }

func logAccess(
	remoteAddr, identifier, authuser, timestamp, request,
	status, bytesSent string,
) {
	out := fmt.Sprintf("%v %v %v [%v] \"%v\" %v %v\n",
		remoteAddr, identifier, authuser, timestamp, request, status, bytesSent,
	)
	file, err := os.OpenFile(PATH_ACCESSLOG, os.O_APPEND|os.O_WRONLY, os.ModeType)
	if err != nil {
		log.Println("couldn't open log access file at", PATH_ACCESSLOG)
	}
	defer file.Close()
	file.WriteString(out)
}

func logError(str string) {
	file, err := os.OpenFile(PATH_ERRORLOG, os.O_APPEND|os.O_WRONLY, os.ModeType)
	if err != nil {
		log.Println("couldn't open log error file at", PATH_ACCESSLOG)
	}
	defer file.Close()
	file.WriteString(str + "\n")
}

func getOrDash(str string) string {
	if str == "" {
		return "-"
	}
	return str
}
