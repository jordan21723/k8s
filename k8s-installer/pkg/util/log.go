package util

import (
	"fmt"
	"time"
)

func LogStyleMessage(logLevel, message string) string {
	template := `{"level":"%s","msg":"%s","time":"%s"}`
	return fmt.Sprintf(template, logLevel, message, time.Now().Format("2006-01-02T15:04:05Z07:00"))
}
