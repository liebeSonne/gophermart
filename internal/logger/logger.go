package logger

import (
	"errors"
)

var ErrUnknownLogLevel = errors.New("unknown log level")
var ErrUnknownLogFormat = errors.New("unknown log format")

type Logger interface {
	Print(v ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}
