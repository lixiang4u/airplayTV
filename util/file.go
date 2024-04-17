package util

import (
	"fmt"
	"io"
	url2 "net/url"
	"os"
	"path/filepath"
	"strings"
)

func AppPath() string {
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}

// 目录是否存在
func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// 递归创建目录，存在则直接返回
func MkdirAll(path string) error {
	if PathExist(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func handlePureUrl(rawUrl string) string {
	tmpUrl, err := url2.Parse(rawUrl)
	if err != nil {
		return rawUrl
	}
	return fmt.Sprintf("%s://%s%s", tmpUrl.Scheme, tmpUrl.Host, tmpUrl.Path)
}

// 根据视频id生成视频本地存储地址
func NewLocalVideoFileName(id, rawUrl string) string {
	//pureUrl := handlePureUrl(rawUrl)
	hash := StringMd5(fmt.Sprintf("%s,%s", id, rawUrl))
	path := fmt.Sprintf("%s/app/m3u8/%s", AppPath(), hash[0:2])
	file := fmt.Sprintf("%s/%s", path, hash)
	file = fmt.Sprintf("%s.%s", file, "m3u8")
	_ = MkdirAll(path)

	return file
}

func GetLocalVideoFileUrl(absLocalPath string) string {
	return strings.TrimPrefix(absLocalPath, fmt.Sprintf("%s/app", AppPath()))
}

func GetCollyCacheDir() string {
	return fmt.Sprintf("%s/app/cache/colly", AppPath())
}

func FileReadAll(filename string) ([]byte, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(fi)
}
