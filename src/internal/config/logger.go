package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	if entry.Data["source"] == "gorm" {
		message := fmt.Sprintf("[%s][%sQuery%s] %s\n", timestamp, colorBlue, colorReset, entry.Message)
		return []byte(message), nil
	}

	level := strings.ToUpper(entry.Level.String())
	color := levelColor(entry.Level)
	message := fmt.Sprintf("[%s] [%s%s%s] - %s\n", timestamp, color, level, colorReset, entry.Message)
	return []byte(message), nil
}

const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
)

func levelColor(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return colorRed
	case logrus.WarnLevel:
		return colorYellow
	case logrus.InfoLevel:
		return colorGreen
	case logrus.DebugLevel:
		return colorBlue
	case logrus.TraceLevel:
		return colorBlue
	default:
		return colorReset
	}
}

var Logger = logrus.New()

var once sync.Once

func StartLogger() {
	once.Do(func() {
		Logger.SetFormatter(&CustomFormatter{})
		Logger.SetLevel(logrus.InfoLevel)
	})
}
