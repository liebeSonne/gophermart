package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	iocloser "github.com/liebeSonne/gophermart/internal/io/closer"
)

const appID = "gophermart"

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	closer := iocloser.MultiCloser{}
	defer func() {
		closeErr := closer.Close()
		if closeErr != nil {
			log.Fatalf("error closing closer: %v", closeErr)
		}
	}()

	cfg := initConfig()
	logger := initLogger(cfg)

	logger.Infow("Config", map[string]any{
		"RunAddress":           cfg.RunAddress,
		"AccrualSystemAddress": cfg.AccrualSystemAddress,
		"LogLevel":             cfg.LogLevel,
		"AuthCookieTokenKey":   cfg.AuthCookieTokenKey,
		"AuthTokenExpires":     cfg.AuthTokenExpires,
	})

	err := runMigrator(cfg)
	if err != nil {
		logger.Errorf("Failed to run migrator: %s", err.Error())
		return
	}

	err = runApp(ctx, cfg, &closer, logger)
	if err != nil {
		logger.Errorf("error running app: %s", err.Error())
		return
	}
}
