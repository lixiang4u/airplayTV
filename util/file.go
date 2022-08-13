package util

import (
	"fmt"
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

// 根据视频id生成视频本地存储地址
func NewLocalVideoFileName(id, url string) string {
	hash := StringMd5(id)
	path := fmt.Sprintf("%s/app/m3u8/%s", AppPath(), hash[0:2])
	file := fmt.Sprintf("%s/%s", path, hash)
	if filepath.Ext(url) != "" {
		file = fmt.Sprintf("%s.%s", file, strings.Trim(filepath.Ext(url), "."))
	}
	_ = MkdirAll(path)

	return file
}

func GetLocalVideoFileUrl(absLocalPath string) string {
	return strings.TrimPrefix(absLocalPath, fmt.Sprintf("%s/app", AppPath()))
}
