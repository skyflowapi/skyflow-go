/*
	Copyright (c) 2022 Skyflow, Inc. 
*/
package logger

import "github.com/sirupsen/logrus"

// This is the description for LogLevel enum
type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	WARN
)

var log = logrus.New()

func init() {
	var formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(logrus.ErrorLevel)
}

// Internal
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Internal
func Info(args ...interface{}) {
	log.Info(args...)
}

// Internal
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Internal
func Error(args ...interface{}) {
	log.Error(args...)
}

// This is the description for SetLogLevel function
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
	default:
		log.SetLevel(logrus.ErrorLevel)
	}
}
