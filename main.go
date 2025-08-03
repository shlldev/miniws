package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/shlldev/miniws/miniws"
)

func main() {
	parser := argparse.NewParser("miniws", "")

	port := parser.Int("p", "port", &argparse.Options{Default: 8040})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs"})
	configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config"})
	wwwFolder := parser.String("w", "www-folder", &argparse.Options{Default: "www"})

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
