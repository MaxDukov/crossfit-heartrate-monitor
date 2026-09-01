// Package config — конфигурация приложения из переменных окружения.
package config

import (
	"os"
	"strconv"
)

// Config — настройки бэкенда.
type Config struct {
	DBPath      string // путь к SQLite-файлу (CF_DB_PATH)
	FrontendDir string // каталог собранного SPA (CF_FRONTEND_DIR)
	DevMode     bool   // старт в mock-режиме (CF_DEV_MODE)
	Port        int    // порт HTTP-сервера (CF_PORT)
}

// Load читает конфигурацию из окружения.
func Load() Config {
	cfg := Config{
		DBPath:      envOr("CF_DB_PATH", "/tmp/cf_monitor.db"),
		FrontendDir: os.Getenv("CF_FRONTEND_DIR"),
		Port:        8001,
	}
	if p, err := strconv.Atoi(envOr("CF_PORT", "8001")); err == nil {
		cfg.Port = p
	}
	cfg.DevMode = os.Getenv("CF_DEV_MODE") == "1"
	return cfg
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
