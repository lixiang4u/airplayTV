package util

import (
	"regexp"
	"strconv"
	"strings"
)

// 直接从url里解析数字串作为id返回
func CZHandleUrlToId(url string) string {
	regex := regexp.MustCompile(`(\d{1,9})`)
	return regex.FindString(url)
}

// 根据规则从url里解析出id
func CZHandleUrlToId2(url string) string {
	tmpList := strings.Split(url, "/")
	return strings.Trim(tmpList[len(tmpList)-1], ".html")
}

// 把字符串页数转为数字，处理空字符串问题
func HandlePageNumber(page string) int {
	n, err := strconv.Atoi(page)
	if err != nil {
		return 1
	}
	if n <= 0 {
		return 1
	}
	return n
}
