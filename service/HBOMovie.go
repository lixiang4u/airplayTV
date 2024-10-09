package service

import (
	"fmt"
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
	hboHost      = "https://hbottv.com"
	hboTagUrl    = "https://hbottv.com/index.php/api/vod"
	hboSearchUrl = "https://hbottv.com/vod/v1/search?wd=%s&limit=20&page=%d"
	hboDetailUrl = "https://hbottv.com/detail/%s.html"
	hboPlayUrl   = "https://hbottv.com/vod/v1/info?id=%s&tid=%s"
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
	var info = model.MovieInfo{}

	var tmpList = strings.Split(id, "_")
	if len(tmpList) != 2 {
		log.Println("[参数错误]", id)
		return info
	}

	b, err := x.httpWrapper.Get(fmt.Sprintf(hboDetailUrl, tmpList[0], tmpList[1]))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	var result = gjson.ParseBytes(b)

	info.Id = fmt.Sprintf("%s_%s", result.Get("data").Get("vod_id").String(), result.Get("data").Get("type_id").String())
	info.Name = result.Get("data").Get("vod_name").String()
	info.Thumb = result.Get("data").Get("vod_pic").String()
	info.Intro = result.Get("data").Get("vod_content").String()
	info.Url = fmt.Sprintf(hboDetailUrl, result.Get("data").Get("vod_id").String(), result.Get("data").Get("type_id").String())
	info.Actors = result.Get("data").Get("vod_content").String()

	result.Get("data").Get("vod_sources").ForEach(func(key, value gjson.Result) bool {

		var tmpSourceId = value.Get("source_id").String()
		var tmpGroup = value.Get("source_name").String()
		if value.Get("vod_play_list").Get("url_count").Int() > 0 {
			value.Get("vod_play_list").Get("urls").ForEach(func(key, value gjson.Result) bool {

				info.Links = append(info.Links, model.Link{
					Id:    fmt.Sprintf("%s_%s", tmpSourceId, value.Get("nid").String()),
					Name:  value.Get("name").String(),
					Url:   value.Get("url").String(),
					Group: tmpGroup,
				})
				return true
			})
		}

		return true
	})

	return info
}

func (x *HBOMovie) hboVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Vid: vid}

	log.Println("[sid]", sid)
	log.Println("[vid]", vid)

	var tmpSourceIdList = strings.Split(sid, "_")
	if len(tmpSourceIdList) != 2 {
		log.Println("[参数错误s]", sid)
		return video
	}

	var tmpVideoIdList = strings.Split(vid, "_")
	if len(tmpVideoIdList) != 2 {
		log.Println("[参数错误v]", vid)
		return video
	}

	b, err := x.httpWrapper.Get(fmt.Sprintf(hboDetailUrl, tmpVideoIdList[0], tmpVideoIdList[1]))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	var result = gjson.ParseBytes(b)

	video.Name = result.Get("data").Get("vod_name").String()
	video.Thumb = result.Get("data").Get("vod_pic").String()

	result.Get("data").Get("vod_sources").ForEach(func(key, value gjson.Result) bool {

		if value.Get("source_id").String() == tmpSourceIdList[0] && value.Get("vod_play_list").Get("url_count").Int() > 0 {
			value.Get("vod_play_list").Get("urls").ForEach(func(key, value gjson.Result) bool {

				if tmpSourceIdList[1] == value.Get("nid").String() {
					video.Url = value.Get("url").String()
					video.Source = value.Get("url").String()
					video.Type = util.GuessVideoType(video.Source)
				}

				return true
			})
		}

		return true
	})

	return video
}
