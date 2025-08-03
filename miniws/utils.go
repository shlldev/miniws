package miniws

import (
	"strconv"
	"strings"
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

func ensureSlashSuffix(str string) string {
	return strings.TrimSuffix(str, "/") + "/"
}
