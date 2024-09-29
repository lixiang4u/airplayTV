package service

import (
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"log"
	"strings"
)

var (
	aspUrl = "XXX"
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
