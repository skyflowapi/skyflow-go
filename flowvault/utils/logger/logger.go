package logger

import (
	"io"

	v2logger "github.com/skyflowapi/skyflow-go/common/logger"
)

// Shared logger from skyflow-go/common/logger, re-exported under flowvault's path. LogLevel is an
// alias, so it is the same type.
type LogLevel = v2logger.LogLevel

const (
	ERROR LogLevel = v2logger.ERROR
	INFO  LogLevel = v2logger.INFO
	DEBUG LogLevel = v2logger.DEBUG
	WARN  LogLevel = v2logger.WARN
	OFF   LogLevel = v2logger.OFF
)

func Debug(args ...interface{}) { v2logger.Debug(args...) }
func Info(args ...interface{})  { v2logger.Info(args...) }
func Warn(args ...interface{})  { v2logger.Warn(args...) }
func Error(args ...interface{}) { v2logger.Error(args...) }

func SetOutput(w io.Writer)      { v2logger.SetOutput(w) }
func SetLogLevel(level LogLevel) { v2logger.SetLogLevel(level) }
