package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/shlldev/miniws/miniws"
)

const (
	HELP_PORT         string = "what port miniws will run on"
	HELP_LOGFOLDER    string = "the logs folder"
	HELP_CONFIGFOLDER string = "the configurations folder"
	HELP_WWWFOLDER    string = "the www folder where miniws will look for files to serve"
)

func main() {
	parser := argparse.NewParser("miniws", "")

	port := parser.Int("p", "port", &argparse.Options{Default: 8040, Help: HELP_PORT})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs", Help: HELP_LOGFOLDER})
	configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config", Help: HELP_CONFIGFOLDER})
	wwwFolder := parser.String("w", "www-folder", &argparse.Options{Default: ".", Help: HELP_WWWFOLDER})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	webserver := miniws.NewWebServer(*port, *logFolder, *configFolder, *wwwFolder)
	webserver.Run()
}
