package config

import (
	"log"
	"os"
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

// getEnvBool reads a boolean environment variable, falling back to defaultVal.
//
// Unrecognized values fall back to defaultVal rather than false. This matters
// for PROXMOX_READ_ONLY: treating garbage as false would let a typo such as
// "tru" silently enable write operations on a cluster.
func getEnvBool(key string, defaultVal bool) bool {
	raw := os.Getenv(key)
	val := strings.ToLower(strings.TrimSpace(raw))
	if val == "" {
		return defaultVal
	}
	switch val {
	case "1", "true", "yes", "on", "enabled", "y", "t":
		return true
	case "0", "false", "no", "off", "disabled", "n", "f":
		return false
	}
	log.Printf("WARNING: %s has unrecognized value %q; using default %v", key, raw, defaultVal)
	return defaultVal
}
