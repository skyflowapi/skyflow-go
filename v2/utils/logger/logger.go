package logger

import "github.com/sirupsen/logrus"

type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	WARN
	OFF
)

var log = logrus.New()

// TO DO
