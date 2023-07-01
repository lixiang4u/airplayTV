package util

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
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

// 合并URL
func ChangeUrlPath(tmpUrl, tmpPath string) string {
	if tmpPath == "" {
		return tmpUrl
	}
	// 如果是 / 开头，直接域名+路径
	if strings.HasPrefix(tmpPath, "/") {
		return fmt.Sprintf("%s/%s", HandleHost(tmpUrl), strings.TrimLeft(tmpPath, "/"))
	}
	parsedUrl, err := url.Parse(tmpUrl)
	if err != nil {
		return tmpPath
	}
	// 防止空域导致地址不对
	if parsedUrl.Host == "" {
		return tmpPath
	}
	return fmt.Sprintf(
		"%s://%s/%s/%s",
		parsedUrl.Scheme,
		parsedUrl.Host,
		strings.TrimLeft(path.Dir(parsedUrl.Path), "/"),
		tmpPath,
	)
}

func CheckVideoUrl(url string) bool {

	log.Println("[checkUrl]", url)
	var httpW = HttpWrapper{}
	headers, err := httpW.Head(url)
	if err != nil {
		log.Println("[CheckVideoUrl.Error]", err.Error())
		return false
	}
	v, ok := headers["Content-Type"]
	if !ok {
		return false
	}
	for _, s := range v {
		if s == "video/mp4" {
			return true
		}
	}
	return false
}
