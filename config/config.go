package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ApiURL      string
	Username    string
	Password    string
	TokenID     string
	TokenSecret string
	ReadOnly    bool // defaults to true (secure by default)
}

func Load() *Config {
	return &Config{
		ApiURL:      getEnv("PROXMOX_API_URL", "https://localhost:8006/api2/json"),
		Username:    os.Getenv("PROXMOX_USERNAME"),
		Password:    os.Getenv("PROXMOX_PASSWORD"),
		TokenID:     os.Getenv("PROXMOX_TOKEN_ID"),
		TokenSecret: os.Getenv("PROXMOX_TOKEN_SECRET"),
		ReadOnly:    getEnvBool("PROXMOX_READ_ONLY", true), // defaults to TRUE
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if val == "" {
		return defaultVal
	}
	// Truthy: 1, true, yes, on, enabled, y, t
	switch val {
	case "1", "true", "yes", "on", "enabled", "y", "t":
		return true
	}
	// Use strconv.ParseBool as fallback, default false for unknown
	if parsed, err := strconv.ParseBool(val); err == nil {
		return parsed
	}
	return false
}
