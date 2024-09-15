package model

type Video struct {
	Id     string `json:"id"`  // 播放ID，vid下的播放源
	Vid    string `json:"vid"` // 电影/电视/综艺界面ID
	Name   string `json:"name"`
	Source string `json:"source"`
	Url    string `json:"url"`
	Type   string `json:"type"`
	Thumb  string `json:"thumb"`
}
