package main

import (
	"os"
	"strings"
)

func fetchFileContents(filepath string) ([]byte, error) {
	if filepath == "/" {
		filepath = "."
	} else {
		filepath_relative, _ := strings.CutPrefix(filepath, "/")
		filepath = filepath_relative
	}
	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	if fileinfo.IsDir() {
		filepath += "/index.html"
	}
	return os.ReadFile(filepath)

}
