package model

type Video struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Url    string `json:"url"`
	Type   string `json:"type"`
	Thumb  string `json:"thumb"`
}
