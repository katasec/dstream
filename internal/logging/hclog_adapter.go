package logging

import (
	"io"
	"log"
	"os"

	hclog "github.com/hashicorp/go-hclog"
)

// HcLogAdapter bridges your custom Logger to the hclog.Logger interface.
type HcLogAdapter struct {
	delegate Logger
	name     string
}

// NewHcLogAdapter creates an hclog-compatible adapter from your Logger.
func NewHcLogAdapter(delegate Logger) hclog.Logger {
	return &HcLogAdapter{delegate: delegate}
}

func (l *HcLogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Trace, hclog.Debug:
		l.delegate.Debug(msg, args...)
	case hclog.Info:
		l.delegate.Info(msg, args...)
	case hclog.Warn:
		l.delegate.Warn(msg, args...)
	case hclog.Error:
		l.delegate.Error(msg, args...)
	default:
		l.delegate.Info(msg, args...)
	}
}

func (l *HcLogAdapter) Trace(msg string, args ...interface{}) { l.Log(hclog.Trace, msg, args...) }
func (l *HcLogAdapter) Debug(msg string, args ...interface{}) { l.Log(hclog.Debug, msg, args...) }
func (l *HcLogAdapter) Info(msg string, args ...interface{})  { l.Log(hclog.Info, msg, args...) }
func (l *HcLogAdapter) Warn(msg string, args ...interface{})  { l.Log(hclog.Warn, msg, args...) }
func (l *HcLogAdapter) Error(msg string, args ...interface{}) { l.Log(hclog.Error, msg, args...) }

func (l *HcLogAdapter) IsTrace() bool { return true }
func (l *HcLogAdapter) IsDebug() bool { return true }
func (l *HcLogAdapter) IsInfo() bool  { return true }
func (l *HcLogAdapter) IsWarn() bool  { return true }
func (l *HcLogAdapter) IsError() bool { return true }

func (l *HcLogAdapter) ImpliedArgs() []interface{} { return nil }

func (l *HcLogAdapter) With(args ...interface{}) hclog.Logger {
	return l
}

func (l *HcLogAdapter) Named(name string) hclog.Logger {
	return &HcLogAdapter{delegate: l.delegate, name: name}
}

func (l *HcLogAdapter) Name() string {
	return l.name
}

func (l *HcLogAdapter) ResetNamed(name string) hclog.Logger {
	return l.Named(name)
}

func (l *HcLogAdapter) SetLevel(hclog.Level) {}

func (l *HcLogAdapter) GetLevel() hclog.Level {
	return hclog.Info
}

// StandardLogger returns a standard logger that writes to stderr to prevent plugin handshake issues.
func (l *HcLogAdapter) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

// StandardWriter returns stderr to avoid writing to stdout before handshake.
func (l *HcLogAdapter) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}
