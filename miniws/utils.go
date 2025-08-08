package miniws

import (
	"strconv"
)

func getHttpVersionString(major, minor int) string {
	return "HTTP/" + strconv.Itoa(major) + "." + strconv.Itoa(minor)
}

func getOrDash(str string) string {
	if str == "" {
		return "-"
	}
	return str
}
