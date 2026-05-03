package main

import (
	"log"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/config"
)

const appID = "gophermart"

func main() {
	cfg, err := config.LoadConfig(appID)
	if err != nil {
		log.Fatalf("error loading config: %s", err.Error())
	}

	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("error initializing logger: %s", err.Error())
	}

	logger.WithFields(logrus.Fields{
		"RunAddress":           cfg.RunAddress,
		"AccrualSystemAddress": cfg.AccrualSystemAddress,
		"LogLevel":             cfg.LogLevel,
	}).Infoln("Config")
}
