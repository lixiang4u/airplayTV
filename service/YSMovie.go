package service

import (
	"fmt"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"strconv"
	"strings"
)

var (
	ysHost      = "https://yingshi.tv"
	ysTagUrl    = "https://api.yingshi.tv/vod/v1/vod/list?order=desc&limit=30&tid=2&by=time&page=%d"
	ysSearchUrl = "https://api.yingshi.tv/vod/v1/search?wd=%s&limit=20&page=%d"
	ysDetailUrl = "https://api.yingshi.tv/vod/v1/info?id=%s&tid=%s"
	ysPlayUrl   = "https://api.yingshi.tv/vod/v1/info?id=%s&tid=%s"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type YSMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
	btVerifyUrl string
}

func (x *YSMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader(headers.Origin, ysHost)
	x.httpWrapper.SetHeader(headers.Referer, ysHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
}

func (x *YSMovie) ListByTag(tagName, page string) model.Pager {
	return x.ysListByTag(tagName, page)
}

func (x *YSMovie) Search(search, page string) model.Pager {
	return x.ysListBySearch(search, page)
}

func (x *YSMovie) Detail(id string) model.MovieInfo {
	return x.ysVideoDetail(id)
}

func (x *YSMovie) Source(sid, vid string) model.Video {
	return x.ysVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *YSMovie) ysListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 30

	b, err := x.httpWrapper.Get(fmt.Sprintf(ysTagUrl, _page))
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
			Url:   fmt.Sprintf(ysDetailUrl, value.Get("vod_id").String(), value.Get("type_id").String()),
			//Actors:     "",
			//Tag:        "",
			//Resolution: "",
			//Links:      nil,
		})
		return true
	})

	return pager
}

func (x *YSMovie) ysListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	b, err := x.httpWrapper.Get(fmt.Sprintf(ysSearchUrl, query, util.HandlePageNumber(page)))
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
			Url:   fmt.Sprintf(ysDetailUrl, value.Get("vod_id").String(), value.Get("type_id").String()),
			//Actors:     "",
			//Tag:        "",
			//Resolution: "",
			//Links:      nil,
		})
		return true
	})

	return pager
}

func (x *YSMovie) ysVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	var tmpList = strings.Split(id, "_")
	if len(tmpList) != 2 {
		log.Println("[参数错误]", id)
		return info
	}

	b, err := x.httpWrapper.Get(fmt.Sprintf(ysDetailUrl, tmpList[0], tmpList[1]))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	var result = gjson.ParseBytes(b)

	info.Id = fmt.Sprintf("%s_%s", result.Get("data").Get("vod_id").String(), result.Get("data").Get("type_id").String())
	info.Name = result.Get("data").Get("vod_name").String()
	info.Thumb = result.Get("data").Get("vod_pic").String()
	info.Intro = result.Get("data").Get("vod_content").String()
	info.Url = fmt.Sprintf(ysDetailUrl, result.Get("data").Get("vod_id").String(), result.Get("data").Get("type_id").String())
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

func (x *YSMovie) ysVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: vid}

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

	b, err := x.httpWrapper.Get(fmt.Sprintf(ysDetailUrl, tmpVideoIdList[0], tmpVideoIdList[1]))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	var result = gjson.ParseBytes(b)

	video.Name = result.Get("data").Get("vod_name").String()
	video.Thumb = result.Get("data").Get("vod_pic").String()

	result.Get("data").Get("vod_sources").ForEach(func(key, value gjson.Result) bool {

		log.Println("[vod_sources.key]", key)

		if value.Get("source_id").String() == tmpSourceIdList[0] && value.Get("vod_play_list").Get("url_count").Int() > 0 {
			value.Get("vod_play_list").Get("urls").ForEach(func(key, value gjson.Result) bool {

				if tmpSourceIdList[1] == value.Get("nid").String() {
					video.Url = value.Get("url").String()
					video.Source = value.Get("url").String()
					video.Type = "hls"
				}

				return true
			})
		}

		return true
	})

	return video
}
