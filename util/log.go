package util

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"time"
)

func InitLog() {
	logf, err := rotatelogs.New(
		"./app/log/%Y-%m-%d.log",
		rotatelogs.WithLinkName("./app/log/app.log"),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
		return
	}
	log.SetOutput(logf)
}
