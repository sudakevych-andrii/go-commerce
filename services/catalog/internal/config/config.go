package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds runtime settings for the catalog service, read from env.
type Config struct {
	Env             string
	HTTPAddr        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	shutdown, err := durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	if shutdown <= 0 {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be positive, got %s", shutdown)
	}

	return Config{
		Env:             stringEnv("APP_ENV", "dev"),
		HTTPAddr:        stringEnv("HTTP_ADDR", ":8081"),
		ShutdownTimeout: shutdown,
	}, nil
}

func stringEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func durationEnv(key string, def time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q: %w", key, v, err)
	}
	return d, nil
}
