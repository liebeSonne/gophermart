package config

import (
	"errors"
	"fmt"
)

var ErrInvalidLogFormat = errors.New("invalid log format")

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

func (lf *LogFormat) String() string {
	return string(*lf)
}

func (lf *LogFormat) Validate() error {
	switch *lf {
	case LogFormatJSON, LogFormatText:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLogFormat, lf)
	}
}

func (lf *LogFormat) Set(val string) error {
	r := LogFormat(val)
	err := r.Validate()
	if err != nil {
		return err
	}
	*lf = r
	return nil
}

func (lf *LogFormat) SetValue(val string) error {
	return lf.Set(val)
}
