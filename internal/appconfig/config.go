package appconfig

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AdminKey    string
	DatabaseURL string
	ListenAddr  string
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	return configFromEnv(os.Getenv), nil
}

func configFromEnv(getenv func(string) string) Config {
	cfg := Config{
		AdminKey:    strings.TrimSpace(getenv("ADMIN_KEY")),
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
		ListenAddr:  strings.TrimSpace(getenv("LISTEN_ADDR")),
	}

	if cfg.AdminKey == "" {
		cfg.AdminKey = "dev"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8787"
	}

	return cfg
}
