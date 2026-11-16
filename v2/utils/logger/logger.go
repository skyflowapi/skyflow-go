package logger

import (
	"io"

	cl "github.com/skyflowapi/skyflow-go/common/logger"
)

// The logger implementation lives in skyflow-go/common/logger and is shared by
// every SDK in this repository. This package keeps v2's import path and API:
// LogLevel is an alias, so it is the same type, and the functions forward.

type LogLevel = cl.LogLevel

const (
	ERROR LogLevel = cl.ERROR
	INFO  LogLevel = cl.INFO
	DEBUG LogLevel = cl.DEBUG
	WARN  LogLevel = cl.WARN
	OFF   LogLevel = cl.OFF
)

func Debug(args ...interface{}) { cl.Debug(args...) }
func Info(args ...interface{})  { cl.Info(args...) }
func Warn(args ...interface{})  { cl.Warn(args...) }
func Error(args ...interface{}) { cl.Error(args...) }

func SetOutput(w io.Writer)      { cl.SetOutput(w) }
func SetLogLevel(level LogLevel) { cl.SetLogLevel(level) }
