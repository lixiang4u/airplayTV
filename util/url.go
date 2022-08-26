package util

import (
	"fmt"
	"net/url"
	"strings"
)

// 根据URL返回带端口的域名
func HandleHost(tmpUrl string) (host string) {
	tmpUrl2, err := url.Parse(tmpUrl)
	if err != nil {
		return
	}
	if tmpUrl2.Host == "" {
		return
	}
	return fmt.Sprintf("%s://%s", tmpUrl2.Scheme, tmpUrl2.Host)
}

// 是否是http协议的路径
func IsHttpUrl(tmpUrl string) bool {
	return strings.HasPrefix(tmpUrl, "http://") || strings.HasPrefix(tmpUrl, "https://")
}
