package utils

import (
	"fmt"
	"time"
)

// Logger provides logging functionality
type Logger struct {
	prefix string
}

// NewLogger creates a new logger
func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
	}
}

// Log logs a message with timestamp
func (l *Logger) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s: %s\n", timestamp, l.prefix, message)
}

// LogError logs an error with timestamp
func (l *Logger) LogError(err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s ERROR: %v\n", timestamp, l.prefix, err)
}
