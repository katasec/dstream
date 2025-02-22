package logging

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// Log levels
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	// Global logger instance
	globalLogger Logger
	once         sync.Once

	// Environment variable names
	envLogLevel    = "DSTREAM_LOG_LEVEL"    // debug, info, warn, error
	envLogHandler  = "DSTREAM_LOG_HANDLER"   // text, json
	envLogWithTime = "DSTREAM_LOG_WITH_TIME" // true, false
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// slogLogger implements Logger using slog
type slogLogger struct {
	logger *slog.Logger
}

// Initialize the global logger
func init() {
	SetupLogging()
}

// SetupLogging configures the global logger based on environment variables
func SetupLogging() {
	once.Do(func() {
		// Get log level from environment
		level := getLogLevel()

		// Get handler type from environment
		handlerType := strings.ToLower(os.Getenv(envLogHandler))
		withTime := strings.ToLower(os.Getenv(envLogWithTime)) == "true"

		// Configure handler options
		opts := &slog.HandlerOptions{
			Level: slog.Level(level),
			AddSource: withTime,
		}

		// Create handler based on type
		var handler slog.Handler
		switch handlerType {
		case "json":
			handler = slog.NewJSONHandler(os.Stdout, opts)
		default:
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		// Create logger
		logger := slog.New(handler)
		globalLogger = &slogLogger{logger: logger}
	})
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	return globalLogger
}

// Helper function to get log level from environment
func getLogLevel() LogLevel {
	level := strings.ToLower(os.Getenv(envLogLevel))
	switch level {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Implementation of Logger interface for slog
func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
