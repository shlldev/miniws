package ipc

import "log"

func LogIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func LogIfErrorFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
