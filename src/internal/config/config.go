package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds everything the process needs to boot. Values are resolved from,
// in order of precedence:
//
//  1. <KEY>_FILE  — path to a file holding the value (Docker secret style)
//  2. <KEY>       — a plain environment variable
//  3. /run/secrets/<key lowercased> — the default Swarm secret mount
//
// The file-based paths matter because the production image is FROM scratch:
// there is no shell to expand secrets into the environment for us.
type Config struct {
	TelegramToken string
	DatabaseURL   string
	HealthAddr    string
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
