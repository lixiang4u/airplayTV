package service

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	hboHost            = "https://hbottv.com"
	hboTagUrl          = "https://hbottv.com/index.php/api/vod"
	hboSearchUrl       = "https://hbottv.com/vod/v1/search?wd=%s&limit=20&page=%d"
	hboDetailUrl       = "https://hbottv.com/detail/%s.html"
	hboPlayUrl         = "https://hbottv.com/play/%s.html"
	hboPlayerConfigUrl = "https://hbottv.com/static/js/playerconfig.js?t=20241012"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type HBOMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
	btVerifyUrl string
}

func (x *HBOMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader(headers.Origin, hboHost)
	x.httpWrapper.SetHeader(headers.Referer, hboHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
	x.httpWrapper.SetHeader(headers.ContentType, "application/x-www-form-urlencoded; charset=UTF-8")
	x.httpWrapper.SetHeader(headers.XRequestedWith, "XMLHttpRequest")
	x.httpWrapper.SetHeader(headers.Referer, "https://hbottv.com/vodshow/1-----------.html")
}

func (x *HBOMovie) ListByTag(tagName, page string) model.Pager {
	return x.hboListByTag(tagName, page)
}

func (x *HBOMovie) Search(search, page string) model.Pager {
	return x.hboListBySearch(search, page)
}

func (x *HBOMovie) Detail(id string) model.MovieInfo {
	return x.hboVideoDetail(id)
}

func (x *HBOMovie) Source(sid, vid string) model.Video {
	return x.hboVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *HBOMovie) hboListByTag(tagName, page string) model.Pager {
	var _page = util.HandlePageNumber(page)

	var pager = model.Pager{}

	// 1728443104
	var ts = fmt.Sprintf("%d", time.Now().Unix())
	log.Println("[TS]", ts)
	var v = url.Values{}
	v.Add("type", "1")
	v.Add("class", "")
	v.Add("area", "")
	v.Add("lang", "")
	v.Add("version", "")
	v.Add("state", "")
	v.Add("letter", "")
	v.Add("page", strconv.Itoa(_page))
	v.Add("time", ts)
	// md5(DS+10位时间戳+DCC147D11943AF75)
	v.Add("key", util.StringMd5(fmt.Sprintf("DS%sDCC147D11943AF75", ts)))

	b, err := x.httpWrapper.Post(hboTagUrl, v.Encode())
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	var result = gjson.ParseBytes(b)

	pager.Total = int(result.Get("total").Int())
	pager.Current = int(result.Get("page").Int())
	pager.Limit = int(result.Get("limit").Int())

	result.Get("list").ForEach(func(key, value gjson.Result) bool {
		pager.List = append(pager.List, model.MovieInfo{
			Id:    value.Get("vod_id").String(),
			Name:  value.Get("vod_name").String(),
			Thumb: value.Get("vod_pic").String(),
			Intro: value.Get("vod_blurb").String(),
			//Url:   fmt.Sprintf(hboDetailUrl, value.Get("vod_id").String(), value.Get("type_id").String()),
			//Actors:     "",
			Tag: value.Get("vod_class").String(),
			//Resolution: "",
			//Links:      nil,
		})
		return true
	})

	return pager
}

func (x *HBOMovie) hboListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	b, err := x.httpWrapper.Get(fmt.Sprintf(hboSearchUrl, query, util.HandlePageNumber(page)))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	var result = gjson.ParseBytes(b)

	pager.Total = int(result.Get("data").Get("Total").Int())
	pager.Current = int(result.Get("data").Get("Page").Int())

	result.Get("data").Get("List").ForEach(func(key, value gjson.Result) bool {
		pager.List = append(pager.List, model.MovieInfo{
			Id:    fmt.Sprintf("%s_%s", value.Get("vod_id").String(), value.Get("type_id").String()),
			Name:  value.Get("vod_name").String(),
			Thumb: value.Get("vod_pic").String(),
			Intro: value.Get("vod_blurb").String(),
			Url:   fmt.Sprintf(hboDetailUrl, value.Get("vod_id").String(), value.Get("type_id").String()),
			//Actors:     "",
			//Tag:        "",
			//Resolution: "",
			//Links:      nil,
		})
		return true
	})

	return pager
}

func (x *HBOMovie) hboVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}

	var httpWrapper = util.HttpWrapper{}
	httpWrapper.SetHeader(headers.Origin, hboHost)
	httpWrapper.SetHeader(headers.Referer, hboHost)
	httpWrapper.SetHeader(headers.UserAgent, ua)
	httpWrapper.SetHeader(headers.Referer, "https://hbottv.com/vodshow/1-----------.html")

	b, err := httpWrapper.Get(fmt.Sprintf(hboDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	//log.Println("[HTML]", string(b))
	//util.FileWriteAllBuf("./aaaaaa.html", b)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	info.Name = doc.Find(".slide-info-title").Text()
	info.Thumb, _ = doc.Find(".vod-detail .detail-pic img").Attr("data-src")
	info.Intro = doc.Find(".vod-detail .switch-box #height_limit").Text()
	info.Url = fmt.Sprintf(hboDetailUrl, id)

	var groupList = make([]string, 0)
	doc.Find(".anthology .anthology-tab a").Each(func(i int, selection *goquery.Selection) {
		groupList = append(groupList, strings.TrimSpace(selection.Text()))
	})

	doc.Find(".anthology .anthology-list .anthology-list-box").Each(func(i int, selection *goquery.Selection) {
		var tmpGroup = groupList[i]
		selection.Find("li").Each(func(i int, selection *goquery.Selection) {
			tmpUrl, _ := selection.Find("a").Attr("href")
			tmpName := strings.TrimSpace(selection.Find("a").Text())
			info.Links = append(info.Links, model.Link{
				Id:    util.SimpleRegEx(tmpUrl, `(\d+-\d+-\d+)`),
				Name:  tmpName,
				Url:   tmpUrl,
				Group: tmpGroup,
			})
		})
	})

	return info
}

func (x *HBOMovie) hboVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Vid: vid}

	log.Println("[sid]", sid)
	log.Println("[vid]", vid)

	var httpWrapper = util.HttpWrapper{}
	httpWrapper.SetHeader(headers.Origin, hboHost)
	httpWrapper.SetHeader(headers.Referer, hboHost)
	httpWrapper.SetHeader(headers.UserAgent, ua)
	httpWrapper.SetHeader(headers.Referer, "https://hbottv.com/vodshow/1-----------.html")

	b, err := httpWrapper.Get(fmt.Sprintf(hboPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	var findJson = util.SimpleRegEx(string(b), `player_aaaa=(\S+)</script>`)
	var result = gjson.Parse(findJson)
	video.Source = result.Get("url").String()
	video.Url = result.Get("url").String()

	var _type = result.Get("encrypt").Int()
	switch _type {
	case 1:
		video.Url = url.QueryEscape(video.Url)
		break
	case 2:
		tmpBuff, _ := base64.StdEncoding.DecodeString(video.Url)
		video.Url, _ = url.QueryUnescape(string(tmpBuff))
		break
	default:
		break
	}

	video.Url = x.handleEncryptUrl(video.Url, result)
	video.Type = util.GuessVideoType(video.Url)

	return video
}

func (x *HBOMovie) handleEncryptUrl(playUrl string, playerAAA gjson.Result) string {

	var parse = ""
	var playServer = playerAAA.Get("server").String()
	var playFrom = playerAAA.Get("from").String()
	if playServer == "no" {
		playServer = ""
	}

	// 获取配置
	var httpWrapper = util.HttpWrapper{}
	httpWrapper.SetHeader(headers.Origin, hboHost)
	httpWrapper.SetHeader(headers.Referer, hboHost)
	httpWrapper.SetHeader(headers.UserAgent, ua)

	b, err := httpWrapper.Get(hboPlayerConfigUrl)
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return ""
	}
	var jsText = string(b)
	var findPlayerConfig = util.SimpleRegEx(jsText, `MacPlayerConfig=(\S+);`)
	var findPlayerList = util.SimpleRegEx(jsText, `MacPlayerConfig.player_list=(\S+),MacPlayerConfig.downer_list`)
	var findServerList = util.SimpleRegEx(jsText, `MacPlayerConfig.server_list=(\S+);`)

	var playerConfigJson = gjson.Parse(findPlayerConfig)
	var playerListJson = gjson.Parse(findPlayerList)
	var serverListJson = gjson.Parse(findServerList)

	serverListJson.ForEach(func(key, value gjson.Result) bool {
		if playServer == key.String() {
			playServer = value.Get("des").String()
		}
		return true
	})

	playerListJson.ForEach(func(key, value gjson.Result) bool {
		if playFrom == key.String() {
			if value.Get("ps").String() == "1" {
				parse = value.Get("parse").String()
				if value.Get("parse").String() == "" {
					parse = playerConfigJson.Get("parse").String()
				}
				playFrom = "parse"
			}
		}
		return true
	})

	log.Println("[playServer]", playServer)
	log.Println("[parse]", parse)
	log.Println("[playFrom]", playFrom)
	log.Println("[playUrl]", playUrl)

	// 获取配置
	log.Println("[GET]", fmt.Sprintf("%s%s", parse, playUrl))
	b, err = x.getHttpWrapper().Get(fmt.Sprintf("%s%s", parse, playUrl))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return ""
	}

	//log.Println("[XCSDFDF]", string(b))

	//var findConfig = util.SimpleRegEx(string(b), `var config = ([\S\s]+)YKQ.start();`)
	var findConfig = util.SimpleRegEx(string(b), `let ConFig = (\{[\s\S]*?\}),box = `)
	//log.Println("[findConfig]", findConfig)
	var configJson = gjson.Parse(findConfig)
	if !configJson.Get("url").Exists() {
		log.Println("[config.parse.error]")
		return ""
	}

	log.Println("[config.url]", configJson.Get("url"))

	log.Println("【uid】", configJson.Get("config").Get("uid").String())

	// '2890' + ConFig['config']['uid'] + 'tB959C'
	bytes, err := util.AesDecrypt(
		[]byte(configJson.Get("url").String()),
		[]byte(fmt.Sprintf("2890%stB959C", configJson.Get("config").Get("uid").String())),
		[]byte("2F131BE91247866E"),
	)
	if err != nil {
		log.Println("[AesDecrypt.Error]", err.Error())
		return ""
	}

	log.Println("[AesDecrypt.Url]", string(bytes)[:50])

	return ""
}

func (x *HBOMovie) getHttpWrapper() *util.HttpWrapper {
	var httpWrapper = util.HttpWrapper{}
	httpWrapper.SetHeader(headers.Origin, hboHost)
	httpWrapper.SetHeader(headers.Referer, hboHost)
	httpWrapper.SetHeader(headers.UserAgent, ua)
	httpWrapper.SetHeader(headers.Referer, "https://hbottv.com/vodshow/1-----------.html")
	return &httpWrapper
}
