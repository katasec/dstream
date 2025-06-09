package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

var (
	// Environment variable names
	envLogLevel = "DSTREAM_LOG_LEVEL" // debug, info, warn, error
	envLogJSON  = "DSTREAM_LOG_JSON"  // true/false
)

// GetLogger returns an hclog-compatible logger for plugins to use
// This logger does NOT add timestamps as plugins should let the host
// handle timestamp formatting to avoid duplication
func GetLogger(name string) hclog.Logger {
	// Get log level from environment
	logLevel := hclog.Info
	if level := os.Getenv(envLogLevel); level != "" {
		logLevel = hclog.LevelFromString(level)
	}

	// Check if JSON logging is enabled
	jsonFormat := false
	if jsonEnv := os.Getenv(envLogJSON); jsonEnv == "true" || jsonEnv == "1" {
		jsonFormat = true
	}

	// Create a logger with no timestamps - the host will add these
	return hclog.New(&hclog.LoggerOptions{
		Name:              name,
		Level:             logLevel,
		Output:            os.Stderr, // Important: use stderr for plugin logs
		JSONFormat:        jsonFormat,
		DisableTime:       true, // Don't add timestamps - host will handle this
		IncludeLocation:   false,
		IndependentLevels: true,
		ColorHeaderOnly:   true,
	})
}

// SetupPluginLogger is a convenience function for plugin main functions
// It creates a logger and returns it, using the plugin name as the logger name
func SetupPluginLogger(pluginName string) hclog.Logger {
	return GetLogger(pluginName)
}

// SetupBareLogger creates a logger with absolutely no formatting
// This is useful when you want completely clean output with no prefixes or timestamps
func SetupBareLogger() hclog.Logger {
	// Get log level from environment
	logLevel := hclog.Info
	if level := os.Getenv(envLogLevel); level != "" {
		logLevel = hclog.LevelFromString(level)
	}

	// Create the underlying logger for compatibility
	baseLogger := hclog.New(&hclog.LoggerOptions{
		Name:              "",
		Level:             logLevel,
		Output:            os.Stderr,
		DisableTime:       true,
		IncludeLocation:   false,
		IndependentLevels: true,
	})
	
	return &bareLogger{
		logger: baseLogger,
		output: os.Stderr,
		level:  logLevel,
	}
}

// SetupCleanLogger creates a logger without the standard prefix format
// This is useful when you want cleaner log output without the [LEVEL] name: prefix
func SetupCleanLogger() hclog.Logger {
	// Get log level from environment
	logLevel := hclog.Info
	if level := os.Getenv(envLogLevel); level != "" {
		logLevel = hclog.LevelFromString(level)
	}

	// Check if JSON logging is enabled
	jsonFormat := false
	if jsonEnv := os.Getenv(envLogJSON); jsonEnv == "true" || jsonEnv == "1" {
		jsonFormat = true
	}

	// Create the underlying logger for compatibility
	baseLogger := hclog.New(&hclog.LoggerOptions{
		Name:              "",
		Level:             logLevel,
		Output:            os.Stderr,
		JSONFormat:        jsonFormat,
		DisableTime:       true,
		IncludeLocation:   false,
		IndependentLevels: true,
	})

	// Create a clean logger with no prefixes
	return &cleanLogger{
		logger: baseLogger,
		output: os.Stderr,
		level:  logLevel,
	}
}

// cleanLogger is a wrapper around hclog.Logger that removes prefixes
type cleanLogger struct {
	logger hclog.Logger
	output io.Writer
	level  hclog.Level
}

// bareLogger is a wrapper around hclog.Logger that provides completely bare output
type bareLogger struct {
	logger hclog.Logger
	output io.Writer
	level  hclog.Level
}

// Implement the hclog.Logger interface for bareLogger
func (l *bareLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	if level >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) Trace(msg string, args ...interface{}) {
	if hclog.Trace >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) Debug(msg string, args ...interface{}) {
	if hclog.Debug >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) Info(msg string, args ...interface{}) {
	if hclog.Info >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) Warn(msg string, args ...interface{}) {
	if hclog.Warn >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) Error(msg string, args ...interface{}) {
	if hclog.Error >= l.level {
		fmt.Fprintf(l.output, "%s%s\n", msg, formatArgs(args...))
	}
}

func (l *bareLogger) IsTrace() bool {
	return l.level <= hclog.Trace
}

func (l *bareLogger) IsDebug() bool {
	return l.level <= hclog.Debug
}

func (l *bareLogger) IsInfo() bool {
	return l.level <= hclog.Info
}

func (l *bareLogger) IsWarn() bool {
	return l.level <= hclog.Warn
}

func (l *bareLogger) IsError() bool {
	return l.level <= hclog.Error
}

func (l *bareLogger) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *bareLogger) With(args ...interface{}) hclog.Logger {
	// We don't support implied args in the bare logger
	return l
}

func (l *bareLogger) Name() string {
	return ""
}

func (l *bareLogger) Named(name string) hclog.Logger {
	// We don't support naming in the bare logger
	return l
}

func (l *bareLogger) ResetNamed(name string) hclog.Logger {
	// We don't support naming in the bare logger
	return l
}

func (l *bareLogger) SetLevel(level hclog.Level) {
	l.level = level
}

func (l *bareLogger) GetLevel() hclog.Level {
	return l.level
}

func (l *bareLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	// Create a standard logger that writes directly to our output
	return log.New(l.output, "", 0)
}

func (l *bareLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.output
}

// ANSI color codes for log levels
const (
	colorReset = "\033[0m"
	colorTrace = "\033[38;5;246m" // Light gray
	colorDebug = "\033[38;5;215m" // Soft orange
	colorInfo  = "\033[38;5;78m"  // Leafy green
	colorWarn  = "\033[38;5;220m" // Amber/golden
	colorError = "\033[38;5;167m" // Crimson
)

// formatArgs formats key-value pairs for structured logging
func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	if len(args)%2 != 0 {
		args = append(args, "MISSING_VALUE")
	}

	var result strings.Builder
	result.WriteString(" ")

	for i := 0; i < len(args); i += 2 {
		if i > 0 {
			result.WriteString(" ")
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

// getTimestamp returns the current timestamp formatted as YYYY/MM/DD HH:MM:SS
func getTimestamp() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

// getLevelColor returns the ANSI color code for the given log level
func getLevelColor(level hclog.Level) string {
	switch level {
	case hclog.Trace:
		return colorTrace
	case hclog.Debug:
		return colorDebug
	case hclog.Info:
		return colorInfo
	case hclog.Warn:
		return colorWarn
	case hclog.Error:
		return colorError
	default:
		return colorReset
	}
}

// getLevelString returns the string representation of the log level
func getLevelString(level hclog.Level) string {
	switch level {
	case hclog.Trace:
		return "TRACE"
	case hclog.Debug:
		return "DEBUG"
	case hclog.Info:
		return "INFO"
	case hclog.Warn:
		return "WARN"
	case hclog.Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Implement the hclog.Logger interface
func (l *cleanLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	if level >= l.level {
		timestamp := getTimestamp()
		levelColor := getLevelColor(level)
		levelStr := getLevelString(level)
		fmt.Fprintf(l.output, "%s %s[%s]%s %s%s\n", timestamp, levelColor, levelStr, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) Trace(msg string, args ...interface{}) {
	if hclog.Trace >= l.level {
		timestamp := getTimestamp()
		fmt.Fprintf(l.output, "%s %s[TRACE]%s %s%s\n", timestamp, colorTrace, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) Debug(msg string, args ...interface{}) {
	if hclog.Debug >= l.level {
		timestamp := getTimestamp()
		fmt.Fprintf(l.output, "%s %s[DEBUG]%s  %s%s\n", timestamp, colorDebug, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) Info(msg string, args ...interface{}) {
	if hclog.Info >= l.level {
		timestamp := getTimestamp()
		fmt.Fprintf(l.output, "%s %s[INFO]%s  %s%s\n", timestamp, colorInfo, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) Warn(msg string, args ...interface{}) {
	if hclog.Warn >= l.level {
		timestamp := getTimestamp()
		fmt.Fprintf(l.output, "%s %s[WARN]%s  %s%s\n", timestamp, colorWarn, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) Error(msg string, args ...interface{}) {
	if hclog.Error >= l.level {
		timestamp := getTimestamp()
		fmt.Fprintf(l.output, "%s %s[ERROR]%s  %s%s\n", timestamp, colorError, colorReset, msg, formatArgs(args...))
	}
}

func (l *cleanLogger) IsTrace() bool {
	return l.level <= hclog.Trace
}

func (l *cleanLogger) IsDebug() bool {
	return l.level <= hclog.Debug
}

func (l *cleanLogger) IsInfo() bool {
	return l.level <= hclog.Info
}

func (l *cleanLogger) IsWarn() bool {
	return l.level <= hclog.Warn
}

func (l *cleanLogger) IsError() bool {
	return l.level <= hclog.Error
}

func (l *cleanLogger) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *cleanLogger) With(args ...interface{}) hclog.Logger {
	// We don't support implied args in the clean logger
	return l
}

func (l *cleanLogger) Name() string {
	return ""
}

func (l *cleanLogger) Named(name string) hclog.Logger {
	// We don't support naming in the clean logger
	return l
}

func (l *cleanLogger) ResetNamed(name string) hclog.Logger {
	// We don't support naming in the clean logger
	return l
}

func (l *cleanLogger) SetLevel(level hclog.Level) {
	l.level = level
}

func (l *cleanLogger) GetLevel() hclog.Level {
	return l.level
}

func (l *cleanLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	// Create a standard logger that writes directly to our output
	return log.New(l.output, "", 0)
}

func (l *cleanLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.output
}
