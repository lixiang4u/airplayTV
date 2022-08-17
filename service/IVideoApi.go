package service

import "github.com/lixiang4u/ShotTv-api/model"

// 定义视频源操作方法，支持多源站接入
type IVideoApi interface {
	ListByTag(tagName, page string) model.Pager
	Search(search, page string) model.Pager
	Detail(id string) model.MovieInfo
	Source(id string) model.Video
}
