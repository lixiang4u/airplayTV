package service

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

var (
	mydHost         = "https://myd04.com/"
	mydImageHost    = "https://www.mdzypic.com/"
	mydTagUrl       = "https://myd04.com/vodshow/1--------%d---.html"
	mydSearchUrl    = "https://myd04.com/vodsearch/%s----------%d---.html"
	mydDetailUrl    = "https://myd04.com/voddetail/%s.html"
	mydM3u8Url      = "https://nnyy.in/url.php"
	mydPlayUrl      = "https://myd04.com/vodplay/%s.html"
	mydPlayFrameUrl = "https://myd04.com/player/?type=%d&url=%s"
	//mydPlayFrameUrl = "https://myd04.com/player/?type=1&url=https://v.cdnlz12.com/20240923/17088_8a1a7530/index.m3u8&token=23bae8bed3694acc42860719a84db8ef"
)

type MYDMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
}

func (x *MYDMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}

	x.httpWrapper.SetHeader(headers.Origin, xkHost)
	x.httpWrapper.SetHeader(headers.Referer, xkHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
}

func (x *MYDMovie) ListByTag(tagName, page string) model.Pager {
	return x._ListByTag("", page)
}

func (x *MYDMovie) Search(search, page string) model.Pager {
	return x._ListBySearch(search, page)
}

func (x *MYDMovie) Detail(id string) model.MovieInfo {
	return x._VideoDetail(id)
}

func (x *MYDMovie) Source(sid, vid string) model.Video {
	return x._VideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *MYDMovie) _ListByTag(tagName, page string) model.Pager {
	var pager = model.Pager{Limit: 72, Current: util.HandlePageNumber(page)}

	b, err := x.httpWrapper.Get(fmt.Sprintf(mydTagUrl, pager.Current))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".module .module-item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".module-poster-item-title").Text()
		tmpUrl, _ := selection.Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".module-item-note").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:       name,
			Thumb:      util.FillUrlHost(thumb, mydImageHost),
			Url:        util.FillUrlHost(tmpUrl, mydImageHost),
			Tag:        tag,
			Resolution: tag,
		})
	})

	var maxPage = 0
	doc.Find("#page .page-link").Each(func(i int, selection *goquery.Selection) {
		tmpUrl, _ := selection.Attr("href")
		tmpNumber := util.StringToInt(util.SimpleRegEx(tmpUrl, `--------(\d+)---.html`))
		if tmpNumber >= maxPage {
			maxPage = tmpNumber
		}
	})
	pager.Total = pager.Limit*maxPage + 1

	return pager
}

func (x *MYDMovie) _ListBySearch(search, page string) model.Pager {
	var pager = model.Pager{Limit: 15, Current: util.HandlePageNumber(page)}

	b, err := x.httpWrapper.Get(fmt.Sprintf(mydSearchUrl, search, pager.Current))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".module .module-item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".module-card-item-title").Text()
		tmpUrl, _ := selection.Find(".module-card-item-title a").Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".module-item-note").Text()
		//intro := selection.Find(".module-info-item-content").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:       strings.TrimSpace(name),
			Thumb:      util.FillUrlHost(thumb, mydImageHost),
			Url:        util.FillUrlHost(tmpUrl, mydImageHost),
			Tag:        tag,
			Resolution: tag,
			//Intro:      strings.TrimSpace(intro),
		})
	})

	var maxPage = 0
	doc.Find("#page .page-link").Each(func(i int, selection *goquery.Selection) {
		tmpUrl, _ := selection.Attr("href")
		tmpNumber := util.StringToInt(util.SimpleRegEx(tmpUrl, `--------(\d+)---.html`))
		if tmpNumber >= maxPage {
			maxPage = tmpNumber
		}
	})
	pager.Total = pager.Limit*maxPage + 1

	return pager
}

// 根据id获取视频播放列表信息
func (x *MYDMovie) _VideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}
	//var sourceMap = make(map[string]string, 0)

	b, err := x.httpWrapper.Get(fmt.Sprintf(mydDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	var tmpSelection = doc.Find(".module .module-main")
	{
		info.Id = id
		info.Name = tmpSelection.Find(".module-info-heading h1").Text()
		info.Intro = tmpSelection.Find(".module-info-introduction-content").Text()
		info.Thumb, _ = tmpSelection.Find(".module-item-cover .lazyload").Attr("data-original")
		//info.Actors = tmpSelection.Find(".module-info-item-title").Text()
		info.Url = fmt.Sprintf(mydDetailUrl, id)

		info.Intro = strings.TrimSpace(info.Intro)
	}

	var groupList = make([]string, 0)
	doc.Find("#y-playList .tab-item").Each(func(i int, selection *goquery.Selection) {
		tmpGroupName, _ := selection.Attr("data-dropdown-value")
		groupList = append(groupList, tmpGroupName)
	})
	doc.Find(".module .module-list.his-tab-list").Each(func(i int, selection *goquery.Selection) {
		var tmpGroup = groupList[i]
		selection.Find(".module-play-list .module-play-list-link").Each(func(j int, selection *goquery.Selection) {
			tmpUrl, _ := selection.Attr("href")
			info.Links = append(info.Links, model.Link{
				Id:    util.SimpleRegEx(tmpUrl, `(\d+-\d+-\d+)`),
				Name:  selection.Find("span").Text(),
				Url:   tmpUrl,
				Group: tmpGroup,
			})
		})
	})

	return info
}

func (x *MYDMovie) _VideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Source: sid, Vid: vid}

	h, b, err := x.httpWrapper.GetResponse(fmt.Sprintf(mydPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	//doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	//if err != nil {
	//	log.Println("[文档解析失败]", err.Error())
	//	return info
	//}

	var findJson = util.SimpleRegEx(string(b), `player_aaaa=(\{[\s\S]*?\})</script>`)
	log.Println("[player_aaaa]", findJson)
	var result = gjson.Parse(findJson)
	video.Url = result.Get("url").String()

	var _type = result.Get("encrypt").Int()
	switch _type {
	case 1:
		video.Url = url.QueryEscape(video.Url)
		break
	case 2:
		tmpBuff, _ := base64.StdEncoding.DecodeString(video.Url)
		video.Url = url.QueryEscape(string(tmpBuff))
		break
	default:
		break
	}

	video.Source = result.Get("url").String()
	video.Url = x.handleEncryptUrl(fmt.Sprintf(mydPlayFrameUrl, _type, video.Url), result, h)
	video.Type = util.GuessVideoType(video.Url)

	return video
}

func (x *MYDMovie) handleEncryptUrl(playFrameUrl string, playerAAA gjson.Result, header http.Header) string {
	log.Println("[playFrameUrl]", playFrameUrl)

	var parse = ""
	var playServer = playerAAA.Get("server").String()
	var playFrom = playerAAA.Get("from").String()
	var playUrl = playerAAA.Get("url").String()
	if playServer == "no" {
		playServer = ""
	}

	// 获取配置
	b, err := x.httpWrapper.Get("https://myd04.com/static/js/playerconfig.js?t=20240923")
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

	// MacPlayer.Parse + MacPlayer.PlayUrl
	var reqUrl = fmt.Sprintf("%s/%s%s", strings.TrimRight(mydHost, "/"), strings.TrimLeft(parse, "/"), playUrl)
	log.Println("[player.request.url]", reqUrl)

	// 需要带Cookie
	x.httpWrapper.SetHeader(headers.Cookie, header.Get("Set-Cookie"))

	// 获取配置
	b, err = x.httpWrapper.Get(reqUrl)
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return ""
	}
	//var findConfig = util.SimpleRegEx(string(b), `var config = ([\S\s]+)YKQ.start();`)
	var findConfig = util.SimpleRegEx(string(b), `var config = (\{[\s\S]*?\})`)
	var configJson = gjson.Parse(findConfig)
	if !configJson.Get("url").Exists() {
		log.Println("[config.parse.error]")
		return ""
	}

	log.Println("[config.url]", configJson.Get("url"), configJson.Get("id"))

	// key来源：https://myd04.com/player/js/setting.js?v=4
	//// https://myd04.com/static/js/playerconfig.js?t=20240923
	return x.fuckRc4(configJson.Get("url").String(), "202205051426239465", 1)
}

func (x *MYDMovie) getPlayerConfig(playFrameUrl string) {
	// 数据来源：https://myd04.com/static/js/playerconfig.js?t=20240923
	b, err := x.httpWrapper.Get("https://myd04.com/static/js/playerconfig.js?t=20240923")
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return
	}
	var jsText = string(b)
	var findPlayerList = util.SimpleRegEx(jsText, `MacPlayerConfig.player_list=(\S+),MacPlayerConfig.downer_list`)
	var findDownerList = util.SimpleRegEx(jsText, `MacPlayerConfig.downer_list=(\S+),MacPlayerConfig.server_list`)
	var findServerList = util.SimpleRegEx(jsText, `MacPlayerConfig.server_list=(\S+);`)

	var playerListJson = gjson.Parse(findPlayerList)
	var downerListJson = gjson.Parse(findDownerList)
	var serverListJson = gjson.Parse(findServerList)

	log.Println(playerListJson)
	log.Println(downerListJson)
	log.Println(serverListJson)
}

func (x *MYDMovie) fuckRc4(data, key string, t int) string {
	var scriptBuff = append(
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/base64-polyfill.js")),
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/fuck-crypto-bridge-myd.js"))...,
	)
	vm := goja.New()
	_, err := vm.RunString(string(scriptBuff))

	if err != nil {
		log.Println("[LoadGojaError]", err.Error())
		return ""
	}

	var rc4Decode func(string, string, int) string
	err = vm.ExportTo(vm.Get("rc4Decode"), &rc4Decode)
	if err != nil {
		log.Println("[ExportGojaFnError]", err.Error())
		return ""
	}

	return rc4Decode(data, key, t)
}
