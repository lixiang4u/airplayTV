package service

import (
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/util"
)

type Movie struct {
	IsCache bool
}

func (x Movie) NewColly() *colly.Collector {
	if x.IsCache {
		return colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))
	}
	return colly.NewCollector()
}
