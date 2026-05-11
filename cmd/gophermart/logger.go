package main

import (
	"log"
	"os"

	"github.com/liebeSonne/gophermart/internal/config"
	ilogger "github.com/liebeSonne/gophermart/internal/logger"
)

func initLogger(cfg config.Config) ilogger.Logger {
	l, err := ilogger.NewSlogLogger(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return l
}
