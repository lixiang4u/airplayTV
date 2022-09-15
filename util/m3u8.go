package util

import (
	"bytes"
	"fmt"
	"github.com/grafov/m3u8"
	"log"
	"strings"
)

// 处理m3u8内容（修正地址问题）
func HandleM3U8Contents(data []byte, host string) []byte {
	if host == "" {
		return data
	}
	playList, listType, err := m3u8.DecodeFrom(bytes.NewBuffer(data), true)
	if err != nil {
		log.Println("[m3u8.DecodeFrom.error]", err)
		return data
	}

	switch listType {
	case m3u8.MEDIA:
		mediapl := playList.(*m3u8.MediaPlaylist)
		for idx, val := range mediapl.Segments {
			if val == nil {
				continue
			}
			if IsHttpUrl(val.URI) == false {
				mediapl.Segments[idx].URI = fmt.Sprintf("%s/%s", host, strings.TrimLeft(val.URI, "/"))
			}
			if StringInList(HandleHost(mediapl.Segments[idx].URI), CORSConfig) {
				mediapl.Segments[idx].URI = HandleUrlToCORS(mediapl.Segments[idx].URI)
			}
			// log.Println("[fix]", mediapl.Segments[idx].URI)
		}
	case m3u8.MASTER:
		masterpl := playList.(*m3u8.MasterPlaylist)
		for idx, val := range masterpl.Variants {
			if val == nil {
				continue
			}
			if IsHttpUrl(val.URI) == false {
				masterpl.Variants[idx].URI = fmt.Sprintf("%s/%s", host, strings.TrimLeft(val.URI, "/"))
			}
		}
	}

	return playList.Encode().Bytes()
}
