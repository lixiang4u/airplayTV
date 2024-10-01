package service

import (
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"log"
	"strings"
)

var (
	m3u8pUrl          = "http://106.15.56.178:38386/api/m3u8p?q=%s"
	cloudflareUrl     = "http://106.15.56.178:38386/api/cloudflare?q=%s&wait=%s"
	cloudflarePostUrl = "http://106.15.56.178:38386/api/cloudflare?q=%s&wait=%s&post=%s&headers=%s"
)

var (
	HostWithAds = []string{
		"c1.rrcdnbf3.com",
	}
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
