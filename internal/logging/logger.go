package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
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
	colorReset = "\033[0m"
	colorDebug = "\033[38;5;215m" // Debug - soft orange
	colorInfo  = "\033[38;5;78m"  // Info - brighter leafy green
	colorWarn  = "\033[38;5;220m" // kind of an amber/golden sand
	colorError = "\033[38;5;167m" // Error - toned-down crimson
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
	// Standard log package compatible methods
	Printf(format string, v ...any)
	Println(v ...any)

	// Structured logging methods
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
		// Use os.Stderr consistently for all log levels to match plugin logging
		// We use empty prefixes here because we'll add the level indicators in the logging methods
		debugLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)
		infoLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)
		warnLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)
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
	if globalLogger == nil {
		SetupLogging()
	}
	return globalLogger
}

// GetHCLogger returns an hclog-compatible logger that wraps the standard logger
func GetHCLogger() hclog.Logger {
	// Get the standard logger
	stdLogger := GetLogger()

	// Wrap it in the HcLogAdapter
	return NewHcLogAdapter(stdLogger)
}

// Helper function to get log level from environment
func getLogLevel() LogLevel {
	level := strings.ToLower(os.Getenv(envLogLevel))
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo // Default to INFO level
	}
}

// Printf implements the standard log.Printf method
func (l *stdLogger) Printf(format string, v ...any) {
	l.info.Printf(format, v...)
}

// Println implements the standard log.Println method
func (l *stdLogger) Println(v ...any) {
	l.info.Println(v...)
}

// Debug logs a debug message
func (l *stdLogger) Debug(msg string, args ...any) {
	if l.logLevel <= LevelDebug {
		if len(args) > 0 {
			l.debug.Printf("%s [DEBUG-HOST] %s%s %v", colorDebug, colorReset, msg, formatArgs(args))
		} else {
			l.debug.Printf("%s [DEBUG-HOST] %s%s", colorDebug, colorReset, msg)
		}
	}
}

// Info logs an info message
func (l *stdLogger) Info(msg string, args ...any) {
	if l.logLevel <= LevelInfo {
		if len(args) > 0 {
			l.info.Printf("%s[INFO-HOST] %s%s %v", colorInfo, colorReset, msg, formatArgs(args))
		} else {
			l.info.Printf("%s[INFO-HOST] %s%s", colorInfo, colorReset, msg)
		}
	}
}

// Warn logs a warning message
func (l *stdLogger) Warn(msg string, args ...any) {
	if l.logLevel <= LevelWarn {
		if len(args) > 0 {
			l.warn.Printf("%s[WARN-HOST] %s%s %v", colorWarn, colorReset, msg, formatArgs(args))
		} else {
			l.warn.Printf("%s[WARN-HOST] %s%s", colorWarn, colorReset, msg)
		}
	}
}

// Error logs an error message
func (l *stdLogger) Error(msg string, args ...any) {
	if l.logLevel <= LevelError {
		if len(args) > 0 {
			l.error.Printf("%s[ERROR-HOST] %s%s %v", colorError, colorReset, msg, formatArgs(args))
		} else {
			l.error.Printf("%s[ERROR-HOST] %s%s", colorError, colorReset, msg)
		}
	}
}

// Helper function to format key-value pairs for structured logging
func formatArgs(args []any) string {
	if len(args)%2 != 0 {
		args = append(args, "MISSING_VALUE")
	}

	var result strings.Builder
	for i := 0; i < len(args); i += 2 {
		if i > 0 {
			result.WriteString(", ")
		}
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", args[i])
		}
		value := args[i+1]
		result.WriteString(fmt.Sprintf("%s=%v", key, value))
	}
	return result.String()
}

// SetLogLevel sets the log level for the global logger
func SetLogLevel(level string) {
	if globalLogger == nil {
		SetupLogging()
	}

	var lvl LogLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = LevelDebug
	case "warn":
		lvl = LevelWarn
	case "error":
		lvl = LevelError
	default:
		lvl = LevelInfo
	}

	if logger, ok := globalLogger.(*stdLogger); ok {
		logger.logLevel = lvl
	}
}
