package service

import (
	"fmt"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"net/http"
	"strings"
)

var (
	m3u8pUrl          = "http://106.15.56.178:38386/api/m3u8p?q=%s"
	cloudflareUrl     = "http://106.15.56.178:38386/api/cloudflare?q=%s&wait=%s"
	cloudflarePostUrl = "http://106.15.56.178:38386/api/cloudflare?q=%s&wait=%s&post=%s&headers=%s"
	m3u8pAirplayUrl   = "https://air.artools.cc/api/m3u8p?q=%s"
)

func fuckCloudflare(tmpHtml, cloudflareUrl string) string {
	if strings.Contains(tmpHtml, "window._cf_chl_opt") {
		log.Println("[cloudflare] waf", cloudflareUrl)
		var httpWrapper = util.HttpWrapper{}
		buf, err := httpWrapper.Get(cloudflareUrl)
		if err != nil {
			log.Println("[cloudflare] challenge", err.Error())
			return ""
		}
		var result = gjson.ParseBytes(buf)
		tmpHtml = result.Get("html").String()
	}
	return tmpHtml
}

func HandleUrlCorsProxy(m3u8Url string) string {
	resp, err := http.Head(m3u8Url)
	if err != nil {
		log.Println("[Head Url Error]", m3u8Url, err.Error())
		return m3u8Url
	}
	if len(resp.Header.Get("x-amz-request-id")) > 0 {
		// 破逼ctyun为什么返回amz头
		return m3u8Url
	}
	log.Println("[AccessControlAllowOrigin]", resp.Header.Get(headers.AccessControlAllowOrigin))
	if resp.Header.Get(headers.AccessControlAllowOrigin) == "*" {
		return m3u8Url
	}
	return fmt.Sprintf(m3u8pAirplayUrl, m3u8Url)
}
