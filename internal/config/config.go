package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultRunAddress           = ":8080"
	DefaultDatabaseURI          = ""
	DefaultAccrualSystemAddress = ":8081"
)

const (
	RunAddressEnv           = "RUN_ADDRESS"
	DatabaseURIEnv          = "DATABASE_URI"
	AccrualSystemAddressEnv = "ACCRUAL_SYSTEM_ADDRESS"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" env-default:":8080" env-description:"run address: host and port"`
	DatabaseURI          string `env:"DATABASE_URI" env-default:"" env-description:"database URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:":8081" env-description:"accrual system address"`
}

func LoadConfig(appID string) (Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	fs := flag.NewFlagSet(appID, flag.ContinueOnError)

	fs.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "run address: host and port")
	fs.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database URI")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address: host and port")

	fs.Usage = cleanenv.Usage(&cfg, nil, fs.Usage)

	err = fs.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		return Config{}, fmt.Errorf("failed parsing flags: %v", err)
	}

	return cfg, nil
}
