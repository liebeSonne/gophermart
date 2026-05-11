package logger

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/liebeSonne/gophermart/internal/config"
)

var slogLogLevelMap = map[config.LogLevel]slog.Level{
	config.LogLevelDebug: slog.LevelDebug,
	config.LogLevelInfo:  slog.LevelInfo,
	config.LogLevelWarn:  slog.LevelWarn,
	config.LogLevelError: slog.LevelError,
}

func NewSlogLogger(level config.LogLevel, format config.LogFormat, w io.Writer) (Logger, error) {
	slogLevel, ok := slogLogLevelMap[level]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownLogLevel, level)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var logger *slog.Logger

	switch format {
	case config.LogFormatText:
		logger = slog.New(slog.NewTextHandler(w, opts))
	case config.LogFormatJSON:
		logger = slog.New(slog.NewJSONHandler(w, opts))
	default:
		return nil, fmt.Errorf("%v: %s", ErrUnknownLogFormat, format)
	}

	return &slogLogger{
		logger: logger,
	}, nil
}

type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Print(v ...interface{}) {
	l.Infow("", v)
}

func (l *slogLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

func (l *slogLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *slogLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, keysAndValues...)
}

func (l *slogLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}
