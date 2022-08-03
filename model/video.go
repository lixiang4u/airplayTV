package model

type Video struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Url    string `json:"url"`
	Type   string `json:"type"`
	Thumb  string `json:"thumb"`
}
