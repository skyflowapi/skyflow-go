package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	WARN
	OFF
)

var log = logrus.New()

func init() {
	var formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(logrus.ErrorLevel)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func SetLogLevel(level LogLevel) {
	switch level {
	case INFO:
		log.SetLevel(logrus.InfoLevel)
	case DEBUG:
		log.SetLevel(logrus.DebugLevel)
	case WARN:
		log.SetLevel(logrus.WarnLevel)
	case ERROR:
		log.SetLevel(logrus.ErrorLevel)
	case OFF:
		log.SetOutput(io.Discard)
	default:
		log.SetLevel(logrus.ErrorLevel)
	}
}
