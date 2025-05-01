package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger interface {
	Debug(msg string, args ...any)	
	Info(msg string, args ...any)	
	Warn(msg string, args ...any)	
	Error(msg string, args ...any)	
}

// EmptyLogger impl Logger interface
type EmptyLogger struct{}

func (l *EmptyLogger) Debug(msg string, args ...any) {}
func (l *EmptyLogger) Info(msg string, args ...any)  {}
func (l *EmptyLogger) Warn(msg string, args ...any)  {}
func (l *EmptyLogger) Error(msg string, args ...any) {}

// The NewLogger initializes a structured logger with configurable levels, formats
// and option for disable logging (in this case, EmptyLogger is used).
func NewLogger(level, format string, enable bool) Logger {
	if !enable {
		return &EmptyLogger{}
	}

	l := validateSlogLevel(level)

	switch strings.ToUpper(format) {
	case "JSON":
		h := slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: l},
		)
		return slog.New(h)
	case "TXT":
		h := slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: l},
		)
		return slog.New(h)
	default:
		panic("undefined logger format " + format)
	}
}

func validateSlogLevel(l string) slog.Level {
	var slogLevel slog.Level
	switch strings.ToUpper(l) {
	case "DBG", "DEBUG":
		slogLevel = slog.LevelDebug
	case "INFO":
		slogLevel = slog.LevelInfo
	case "WARN", "WARNING":
		slogLevel = slog.LevelWarn
	case "ERR", "ERROR":
		slogLevel = slog.LevelError
	default:
		panic("undefined logger level " + l)
	}

	return slogLevel
}