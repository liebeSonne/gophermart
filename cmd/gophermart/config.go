package main

import (
	"log"

	"github.com/liebeSonne/gophermart/internal/config"
)

func initConfig() config.Config {
	cfg, err := config.LoadConfig(appID)
	if err != nil {
		log.Fatalf("error loading config: %s", err.Error())
	}
	return cfg
}
