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
	if strings.Contains(tmpUrl, "m3u8") {
		t = "hls" // https://t02.cz01.org/play/db64Kjc_bEupHVpe6kYgihH8QSZ8hARNT9aZ34FXoyoZFuoYqU6-DPachsx5lpmd4Uvlz8FUTy02tUCP6fvXK1QKEj4huO3TcGFUry4BBZkYm_tQjQ/m3u8
	}

	return t
}
