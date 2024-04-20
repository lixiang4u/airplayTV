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

// 判断字符串是否在某个列表中
func StringInList(str string, strList []string) bool {
	for _, el := range strList {
		// 不区分大小写到字符比较
		if strings.EqualFold(el, str) {
			return true
		}
	}
	return false
}

// TV: Mozilla/5.0 (Linux; Android 10; BRAVIA VH1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.96 Mobile Safari/537.36
func IsTv(ua string) bool {
	return strings.Contains(ua, "BRAVIA")
}

func SimpleRegEx(plainText, regex string) string {
	//regEx := regexp.MustCompile(`(\d+)`)
	regEx := regexp.MustCompile(regex)
	tmpList := regEx.FindStringSubmatch(plainText)
	if len(tmpList) < 2 {
		return ""
	}
	return tmpList[1]
}
