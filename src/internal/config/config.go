package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken   string
	DatabaseURL     string
	HealthAddr      string
	AdminTelegramID int64
	LogLevel        string
	LogFormat       string
}

func Load() (*Config, error) {
	// Local development convenience. Absent in a container, and that is fine.
	_ = godotenv.Load(".env.dev")

	cfg := &Config{HealthAddr: lookupOr("HEALTH_ADDR", ":8080")}

	var err error
	if cfg.TelegramToken, err = require("TG_TOKEN"); err != nil {
		return nil, err
	}
	if cfg.DatabaseURL, err = require("DATABASE_URL"); err != nil {
		return nil, err
	}
	if cfg.LogLevel, err = require("LOG_LEVEL"); err != nil {
		return nil, err
	}
	if cfg.LogFormat, err = require("LOG_FORMAT"); err != nil {
		return nil, err
	}

	adminID, err := require("ADMIN_TELEGRAM_ID")
	if err != nil {
		return nil, err
	}
	if cfg.AdminTelegramID, err = strconv.ParseInt(adminID, 10, 64); err != nil {
		return nil, fmt.Errorf("config; ADMIN_TELEGRAM_ID must be an integer: %w", err)
	}
	os.Setenv("ADMIN_TELEGRAM_ID", strconv.FormatInt(cfg.AdminTelegramID, 10))

	return cfg, nil
}

func require(key string) (string, error) {
	v, err := lookup(key)
	if err != nil {
		return "", err
	}
	if v == "" {
		return "", fmt.Errorf(
			"config: %s is not set (tried %s_FILE, %s and /run/secrets/%s)",
			key, key, key, strings.ToLower(key),
		)
	}

	return v, nil
}

func lookupOr(key, def string) string {
	v, err := lookup(key)
	if err != nil || v == "" {
		return def
	}

	return v
}

func lookup(key string) (string, error) {
	if path := os.Getenv(key + "_FILE"); path != "" {
		return readTrimmed(path)
	}

	if v := os.Getenv(key); v != "" {
		return v, nil
	}

	path := "/run/secrets/" + strings.ToLower(key)
	if _, err := os.Stat(path); err == nil {
		return readTrimmed(path)
	}

	return "", nil
}

func readTrimmed(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("config: read %s: %w", path, err)
	}

	return strings.TrimSpace(string(b)), nil
}
