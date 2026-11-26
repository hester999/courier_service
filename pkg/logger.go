package pkg

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Level int

const (
	INFO Level = iota
	WARN
	ERROR
	MONITORING
)

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Monitoring(msg string, args ...any)
}

type LoggerImpl struct {
	mu     sync.Mutex
	level  Level
	writer *log.Logger
}

func NewLogger(level Level) *LoggerImpl {
	return &LoggerImpl{
		level:  level,
		writer: log.New(os.Stdout, "", 0),
	}
}

func (l *LoggerImpl) log(level Level, msg string, args ...any) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var levelStr, color string
	switch level {
	case INFO:
		levelStr, color = "INFO", colorBlue
	case WARN:
		levelStr, color = "WARN", colorYellow
	case ERROR:
		levelStr, color = "ERROR", colorRed
	case MONITORING:
		levelStr, color = "MONITORING", colorGreen
	}

	message := fmt.Sprintf(msg, args...)
	l.writer.Printf("%s[%s] [%s] %s%s",
		colorGray, timestamp, color+levelStr+colorReset, message, colorReset)
}

func (l *LoggerImpl) Info(msg string, args ...any)       { l.log(INFO, msg, args...) }
func (l *LoggerImpl) Warn(msg string, args ...any)       { l.log(WARN, msg, args...) }
func (l *LoggerImpl) Error(msg string, args ...any)      { l.log(ERROR, msg, args...) }
func (l *LoggerImpl) Monitoring(msg string, args ...any) { l.log(MONITORING, msg, args...) }
