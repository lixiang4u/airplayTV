package util

import (
	"regexp"
	"strings"
)

func CZHandleUrlToId(url string) string {
	regex := regexp.MustCompile(`(\d{1,9})`)
	return regex.FindString(url)
}

func CZHandleUrlToId2(url string) string {
	tmpList := strings.Split(url, "/")
	return strings.Trim(tmpList[len(tmpList)-1], ".html")
}
