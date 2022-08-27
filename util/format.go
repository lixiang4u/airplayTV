package util

import "encoding/json"

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
