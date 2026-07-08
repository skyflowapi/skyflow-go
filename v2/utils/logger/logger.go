package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	WARN
	OFF
)

var (
	writer   io.Writer = os.Stderr
	levelVar           = new(slog.LevelVar)
	log      *slog.Logger
)

func init() {
	levelVar.Set(slog.LevelError)
	rebuild()
}

// rebuild recreates the underlying logger. slog handlers are immutable and
// bound to a writer, so the logger must be rebuilt when the output changes.
// The level is held in a *slog.LevelVar, so level changes do not require it.
func rebuild() {
	log = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: levelVar}))
}

func Debug(args ...interface{}) {
	log.Debug(fmt.Sprint(args...))
}

func Info(args ...interface{}) {
	log.Info(fmt.Sprint(args...))
}

func Warn(args ...interface{}) {
	log.Warn(fmt.Sprint(args...))
}

func Error(args ...interface{}) {
	log.Error(fmt.Sprint(args...))
}

func SetOutput(w io.Writer) {
	writer = w
	rebuild()
}

func SetLogLevel(level LogLevel) {
	switch level {
	case INFO:
		levelVar.Set(slog.LevelInfo)
	case DEBUG:
		levelVar.Set(slog.LevelDebug)
	case WARN:
		levelVar.Set(slog.LevelWarn)
	case ERROR:
		levelVar.Set(slog.LevelError)
	case OFF:
		SetOutput(io.Discard)
	default:
		levelVar.Set(slog.LevelError)
	}
}
