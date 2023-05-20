package logger

import (
	"github.com/fatih/color"
)

func (logger *Logger) Warn(message string) {
	color.Yellow(message)
}

func (logger *Logger) Info(message string) {
	color.Cyan(message)
}

func (logger *Logger) Error(message string) {
	color.Red(message)
}

func (logger *Logger) Debug(message string) {
	if logger.enableDebug {
		color.Blue(message)
	}
}

func (logger *Logger) EnableDebug() {
	logger.enableDebug = true
}

type Logger struct {
	enableDebug bool
}

func NewLogger(debug bool) *Logger {
	logger := Logger{enableDebug: debug}
	return &logger
}
