package main

import (
	"fmt"
	"log"

	"github.com/liebeSonne/gophermart/internal/config"
)

const appID = "gophermart"

func main() {
	cfg, err := config.LoadConfig(appID)
	if err != nil {
		log.Fatalf("error loading config: %s", err.Error())
	}

	fmt.Printf("Config: %+v\n", cfg)
}
