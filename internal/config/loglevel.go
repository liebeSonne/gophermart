package config

import (
	"errors"
	"fmt"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelPanic LogLevel = "panic"
	LogLevelFatal LogLevel = "fatal"
)

func (ll *LogLevel) String() string {
	return string(*ll)
}

func (ll *LogLevel) Validate() error {
	switch *ll {
	case LogLevelTrace, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLogLevel, *ll)
	}
}

func (ll *LogLevel) Set(val string) error {
	r := LogLevel(val)
	err := r.Validate()
	if err != nil {
		return err
	}
	*ll = r
	return nil
}

func (ll *LogLevel) SetValue(val string) error {
	return ll.Set(val)
}
