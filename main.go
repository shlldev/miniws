package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

const (
	FILENAME_ACCESSLOG       string = "access.log"
	FILENAME_ERRORLOG        string = "error.log"
	FILENAME_IPFILTER        string = "ipfilter.conf"
	FILENAME_USERAGENTFILTER string = "useragentfilter.conf"
)

func main() {

	parser := argparse.NewParser("miniws", "")

	port := parser.Int("p", "port", &argparse.Options{Default: 8040})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs"})
	configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config"})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	webserver := NewWebServer(*port, *logFolder, *configFolder)
	webserver.Run()

}
