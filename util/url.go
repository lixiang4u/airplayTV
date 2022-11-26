package util

import (
	"fmt"
	"log"
	"net/http"
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

func HandleHostname(tmpUrl string) (host string) {
	tmpUrl2, err := url.Parse(tmpUrl)
	if err != nil {
		return
	}
	return tmpUrl2.Hostname()
}

// 是否是http协议的路径
func IsHttpUrl(tmpUrl string) bool {
	return strings.HasPrefix(tmpUrl, "http://") || strings.HasPrefix(tmpUrl, "https://")
}

// 获取重定向内容
func HandleRedirectUrl(requestUrl string) (redirectUrl string) {
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return nil
		},
	}
	_, err := httpClient.Head(requestUrl)
	if err != nil {
		log.Println("[http.Client.Do.Error]", err)
		return requestUrl
	}

	return redirectUrl
}

// 把url转为请求 /api/video/cors 接口的形式，方便后续获取重定向内容
func HandleUrlToCORS(tmpUrl string) string {
	return fmt.Sprintf(
		"%s/api/video/cors?src=%s",
		strings.TrimRight(ApiConfig.Server, "/"),
		url.QueryEscape(tmpUrl),
	)
}
