package logging

import (
	"fmt"
	"log"
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

	// ANSI color codes
	colorReset  = "\033[0m"
	colorGray   = "\033[38;5;242m" // Debug
	colorGreen  = "\033[38;5;46m"  // Info
	colorYellow = "\033[38;5;220m" // Warn
	colorRed    = "\033[38;5;196m" // Error
)

var (
	// Global logger instance
	globalLogger Logger
	once         sync.Once

	// Environment variable names
	envLogLevel = "DSTREAM_LOG_LEVEL" // debug, info, warn, error
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// stdLogger implements Logger using the standard log package
type stdLogger struct {
	logLevel LogLevel
	debug    *log.Logger
	info     *log.Logger
	warn     *log.Logger
	error    *log.Logger
}

// Initialize the global logger
func init() {
	SetupLogging()
}

// SetupLogging configures the global logger based on environment variables
func SetupLogging() {
	once.Do(func() {
		// Get log level from environment
		logLevel := getLogLevel()

			// Create loggers for each level with appropriate prefixes and colors
		debugLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
		infoLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
		warnLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
		errorLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)

		// Create logger
		globalLogger = &stdLogger{
			logLevel: logLevel,
			debug:    debugLogger,
			info:     infoLogger,
			warn:     warnLogger,
			error:    errorLogger,
		}
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

// formatMessage formats the message with any additional arguments
func formatMessage(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}

	// Convert args to key-value pairs
	pairs := make([]string, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			pairs = append(pairs, fmt.Sprintf("%v=%v", args[i], args[i+1]))
		}
	}

	// If there are key-value pairs, append them to the message
	if len(pairs) > 0 {
		return fmt.Sprintf("%s [%s]", msg, strings.Join(pairs, " "))
	}
	return msg
}

// Debug logs a debug message
func (l *stdLogger) Debug(msg string, args ...any) {
	if l.logLevel <= LevelDebug {
		l.debug.Printf("%s[DEBUG]%s %s", colorGray, colorReset, formatMessage(msg, args...))
	}
}

// Info logs an info message
func (l *stdLogger) Info(msg string, args ...any) {
	if l.logLevel <= LevelInfo {
		l.info.Printf("%s[INFO]%s %s", colorGreen, colorReset, formatMessage(msg, args...))
	}
}

// Warn logs a warning message
func (l *stdLogger) Warn(msg string, args ...any) {
	if l.logLevel <= LevelWarn {
		l.warn.Printf("%s[WARN]%s %s", colorYellow, colorReset, formatMessage(msg, args...))
	}
}

// Error logs an error message
func (l *stdLogger) Error(msg string, args ...any) {
	if l.logLevel <= LevelError {
		l.error.Printf("%s[ERROR]%s %s", colorRed, colorReset, formatMessage(msg, args...))
	}
}
