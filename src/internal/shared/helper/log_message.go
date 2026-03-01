package helper

import (
	"log"
	"time"
)

func LogMessage(level string, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.SetFlags(0)
	log.Printf("[%s] [%s] - %s", timestamp, level, message)
}
