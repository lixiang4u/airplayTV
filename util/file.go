package util

import (
	"os"
	"path/filepath"
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
