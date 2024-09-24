package model

type Pager struct {
	Total   int         `json:"total"` // 总记录数，前端限制分页用
	Current int         `json:"current"`
	Limit   int         `json:"limit"`
	List    []MovieInfo `json:"list"`
}
