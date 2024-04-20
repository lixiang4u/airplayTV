package util

import (
	"encoding/json"
	"strings"
)

func ToJSON(data interface{}, pretty bool) string {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(data, "", "\t")
	} else {
		b, err = json.Marshal(data)
	}
	if err != nil {
		return ""
	}
	return string(b)
}

func GuessVideoType(tmpUrl string) string {
	var t = "auto"
	if strings.HasSuffix(tmpUrl, ".mp4") {
		t = "auto"
	}
	if strings.HasSuffix(tmpUrl, ".m3u8") {
		t = "hls" //m3u8 都是hls ???
	}
	return t
}
