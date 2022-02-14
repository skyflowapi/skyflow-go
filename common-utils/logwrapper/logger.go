package logger

import "github.com/sirupsen/logrus"

var log = logrus.New()

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

func SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}
