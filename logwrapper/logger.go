package logwrapper

import "github.com/sirupsen/logrus"

var logger = logrus.New()

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func setLogLevel(level logrus.Level) {
	logger.SetLevel(level)
}
