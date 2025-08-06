package miniws

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
)

const (
	FLAGS_LOG_OPEN    int         = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	FLAGS_CONFIG_OPEN int         = os.O_RDONLY | os.O_CREATE
	PERMS_LOG_OPEN    os.FileMode = os.ModeType | os.ModePerm
	PERMS_CONFIG_OPEN os.FileMode = os.ModeType | os.ModePerm
	PERMS_MKDIR       os.FileMode = os.ModeDir | os.ModePerm
)

type Logger struct {
	logFolder   string
	maxLogBytes int64
}

func NewLogger(logFolder_ string, maxLogBytes_ int64) *Logger {
	return &Logger{
		logFolder:   logFolder_,
		maxLogBytes: maxLogBytes_,
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
	l.writeToLogFileAndRenameIfBig(FILENAME_ACCESSLOG, out)
}

func (l *Logger) logError(str string) {
	os.Mkdir(l.logFolder, PERMS_MKDIR)
	l.writeToLogFileAndRenameIfBig(FILENAME_ERRORLOG, str+"\n")
}

func (l *Logger) writeToLogFileAndRenameIfBig(filename, content string) {
	file, err := os.OpenFile(ensureSlashSuffix(l.logFolder)+filename, FLAGS_LOG_OPEN, PERMS_LOG_OPEN)

	if err != nil {
		log.Println("couldn't open log access file at", ensureSlashSuffix(l.logFolder)+filename)
	}

	defer file.Close()
	file.WriteString(content)

	fileinfo, err := file.Stat()

	if err != nil {
		log.Println("Error reading access file at", ensureSlashSuffix(l.logFolder)+filename)
	}

	if fileinfo.Size() > l.maxLogBytes {
		defer os.Rename(ensureSlashSuffix(
			l.logFolder)+fileinfo.Name(),
			ensureSlashSuffix(l.logFolder)+fileinfo.Name()+"."+uuid.NewString(),
		)
	}
}
