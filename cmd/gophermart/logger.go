package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/config"
)

var configToLoggerLogLevelMap = map[config.LogLevel]logrus.Level{
	config.LogLevelTrace: logrus.TraceLevel,
	config.LogLevelDebug: logrus.DebugLevel,
	config.LogLevelInfo:  logrus.InfoLevel,
	config.LogLevelWarn:  logrus.WarnLevel,
	config.LogLevelError: logrus.ErrorLevel,
	config.LogLevelFatal: logrus.FatalLevel,
	config.LogLevelPanic: logrus.PanicLevel,
}

func initLogger(cfg config.Config) (*logrus.Logger, error) {
	loggerLevel, ok := configToLoggerLogLevelMap[cfg.LogLevel]
	if !ok {
		return nil, fmt.Errorf("unknown log level: %s", cfg.LogLevel)
	}

	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(loggerLevel)

	return logger, nil
}
