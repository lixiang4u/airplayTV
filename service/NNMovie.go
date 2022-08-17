package service

import "github.com/lixiang4u/ShotTv-api/model"

type NNMovie struct{}

func (x NNMovie) ListByTag(tagName, page string) model.Pager {
	return model.Pager{}
}

func (x NNMovie) Search(search, page string) model.Pager {
	var data = model.Pager{}
	data.Total = 888
	return data
}

func (x NNMovie) Detail(id string) model.MovieInfo {
	return model.MovieInfo{}
}

func (x NNMovie) Source(id string) model.Video {
	return model.Video{}
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================
