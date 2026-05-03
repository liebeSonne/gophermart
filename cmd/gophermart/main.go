package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

const appID = "gophermart"

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := initConfig()
	logger := initLogger(cfg)

	logger.WithFields(logrus.Fields{
		"RunAddress":           cfg.RunAddress,
		"AccrualSystemAddress": cfg.AccrualSystemAddress,
		"LogLevel":             cfg.LogLevel,
	}).Infoln("Config")

	err := runApp(ctx, cfg, logger)
	if err != nil {
		logger.Fatalf("error running app: %s", err.Error())
	}
}
