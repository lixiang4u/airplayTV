package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"time"
)

func InitLog() {
	logf, err := rotatelogs.New(
		"./app/log/%Y-%m-%d.log",
		rotatelogs.WithLinkName("./app/log/app.log"),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
		return
	}
	log.SetOutput(logf)
}

func SetGINLoggerFormat() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		// your custom format
		logString := fmt.Sprintf("%s | %d | %s | %s | %s | \"%s\" | %s | %s\n",
			params.TimeStamp.Format(time.RFC3339),
			params.StatusCode,
			params.Latency,
			params.ClientIP,
			params.Method,
			params.Path,
			params.ErrorMessage,
			params.Request.UserAgent(),
		)
		log.Println("[GIN]", logString)
		return ""
	})
}
