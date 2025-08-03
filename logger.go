package main

import (
	"fmt"
	"log"
	"os"
)

const (
	FLAGS_LOG_OPEN    int         = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	FLAGS_CONFIG_OPEN int         = os.O_RDONLY | os.O_CREATE
	PERMS_LOG_OPEN    os.FileMode = os.ModeType | os.ModePerm
	PERMS_CONFIG_OPEN os.FileMode = os.ModeType | os.ModePerm
	PERMS_MKDIR       os.FileMode = os.ModeDir | os.ModePerm
)

type Logger struct {
	logFolder string
}

func NewLogger(logFolder_ string) *Logger {
	return &Logger{
		logFolder: logFolder_,
	}
}

// returns error != nil
func (l *Logger) logIfError(err error) bool {
	if err != nil {
		l.logError(err.Error())
		return true
	}
	return false
}

func (l *Logger) logAccess(
	remoteAddr, identifier, authuser, timestamp, request,
	status, bytesSent, referer, user_agent string,
) {
	out := fmt.Sprintf("%v %v %v [%v] \"%v\" %v %v \"%v\" \"%v\"\n",
		remoteAddr, identifier, authuser, timestamp, request, status, bytesSent, referer, user_agent,
	)
	os.Mkdir(l.logFolder, os.ModeDir|os.ModePerm)
	file, err := os.OpenFile(ensureSlashSuffix(l.logFolder)+FILENAME_ACCESSLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log access file at", ensureSlashSuffix(l.logFolder)+FILENAME_ACCESSLOG)
	}
	defer file.Close()
	file.WriteString(out)
}

func (l *Logger) logError(str string) {
	os.Mkdir(l.logFolder, PERMS_MKDIR)
	file, err := os.OpenFile(ensureSlashSuffix(l.logFolder)+FILENAME_ERRORLOG, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log error file at", ensureSlashSuffix(l.logFolder)+FILENAME_ERRORLOG)
	}
	defer file.Close()
	file.WriteString(str + "\n")
}
