package main

import (
	"fmt"
	"ipc"
	"os"

	"miniws"

	"github.com/akamensky/argparse"
)

const (
	HELP_LONGDESCRIPTION string = "minimal web server - lightweight and easy to configure web " +
		"server.\n\nYou can specify the following options via command line arguments: the server port, " +
		"where to put the configuration files (they are auto-generated on first run), the folder with your " +
		"web content, the folder to put the access and error logs inside, and more (see below)."
	HELP_PORT         string = "what port miniws will run on"
	HELP_LOGFOLDER    string = "the logs folder"
	HELP_CONFIGFOLDER string = "the configurations folder"
	HELP_WWWFOLDER    string = "the www folder where miniws will look for files to serve"
	HELP_MAXLOGBYTES  string = "the maximum bytes after which the log files get split"
	HELP_SIGNAL       string = "runs the executable in command mode, meaning it will just " +
		"send a command to an already running miniws server process, then terminate"
)

func main() {
	parser := argparse.NewParser("miniws", HELP_LONGDESCRIPTION)

	signal := parser.String("s", "signal", &argparse.Options{Help: HELP_SIGNAL})
	port := parser.Int("p", "port", &argparse.Options{Default: 8040, Help: HELP_PORT})
	logFolder := parser.String("l", "logs-folder", &argparse.Options{Default: "logs", Help: HELP_LOGFOLDER})
	configFolder := parser.String("c", "config-folder", &argparse.Options{Default: "config", Help: HELP_CONFIGFOLDER})
	wwwFolder := parser.String("w", "www-folder", &argparse.Options{Default: ".", Help: HELP_WWWFOLDER})
	maxLogBytes := parser.Int("b", "max-log-bytes", &argparse.Options{Default: 1048576, Help: HELP_MAXLOGBYTES})

	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	// signal mode
	if *signal != "" {
		client := ipc.Client{}
		client.OneShotWrite("unix", miniws.SOCKET_PATH, []byte(*signal))
		return
	}

	// webserver mode
	webserver := miniws.NewWebServer(*port, *logFolder, *configFolder, *wwwFolder, int64(*maxLogBytes))
	webserver.Run()

}
