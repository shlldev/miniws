package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/shlldev/miniws/miniws"
)

const (
	HELP_PORT          string = "what port miniws will run on"
	HELP_LOGFOLDER     string = "the logs folder"
	HELP_CONFIGFOLDER  string = "the configurations folder"
	HELP_WWWFOLDER     string = "the www folder where miniws will look for files to serve"
	HELP_MAXLOGBYTES   string = "the maximum bytes after which the log files get split"
	HELP_MAXCLIENTRATE string = "the maximum number of requests per minute that any particular " +
		"client can send. exceeding this rate will cause miniws to reply with HTTP error 429: " +
		"Too Many Requests."
)

func main() {
	parser := argparse.NewParser("miniws", "")

	port := parser.Int("p", "port", &argparse.Options{Default: 8040, Help: HELP_PORT})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs", Help: HELP_LOGFOLDER})
	configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config", Help: HELP_CONFIGFOLDER})
	wwwFolder := parser.String("w", "www-folder", &argparse.Options{Default: ".", Help: HELP_WWWFOLDER})
	maxLogBytes := parser.Int("b", "max-log-bytes", &argparse.Options{Default: 1048576, Help: HELP_MAXLOGBYTES})
	maxClientRatePerMin := parser.Int("r", "max-client-rate", &argparse.Options{Default: 600, Help: HELP_MAXCLIENTRATE})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	webserver := miniws.NewWebServer(miniws.WebServerConfig{
		LogFolder:               *logFolder,
		ConfigFolder:            *configFolder,
		WWWFolder:               *wwwFolder,
		Port:                    uint16(*port),
		MaxBytesPerLogFile:      uint64(*maxLogBytes),
		MaxConnectionsPerMinute: uint64(*maxClientRatePerMin),
	})
	webserver.Run()
}
