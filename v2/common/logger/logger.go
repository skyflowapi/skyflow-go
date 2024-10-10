package logger

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
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
		log.SetOutput(ioutil.Discard)
	default:
		log.SetLevel(logrus.ErrorLevel)
	}
}
