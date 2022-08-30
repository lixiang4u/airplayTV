package service

import (
	"github.com/lixiang4u/airplayTV/model"
)

var (
//eightM3u8Url   = "https://www.88hd.com/vod-play-id-200881-src-4-num-12.html"
//eightPlayUrl   = "https://www.88hd.com/oumeiju/202204/200881.html"
//eightSearchUrl = "https://www.88hd.com/vod-search-pg-2-wd-天.html"
)

type EightMovie struct{ Movie }

func (x EightMovie) ListByTag(tagName, page string) model.Pager {
	return model.Pager{}
}

func (x EightMovie) Search(search, page string) model.Pager {
	return model.Pager{}
}

func (x EightMovie) Detail(id string) model.MovieInfo {
	return model.MovieInfo{}
}

func (x EightMovie) Source(sid, vid string) model.Video {
	return model.Video{}
}
