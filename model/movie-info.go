package model

type MovieInfo struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Thumb      string `json:"thumb"`
	Intro      string `json:"intro"`
	Url        string `json:"url"`
	Actors     string `json:"actors"`
	Tag        string `json:"tag"`
	Resolution string `json:"resolution"`
	Links      []Link `json:"links"`
}
