package model

type Pager struct {
	Total   int         `json:"total"`
	Current int         `json:"current"`
	Limit   int         `json:"limit"`
	List    []MovieInfo `json:"list"`
}
