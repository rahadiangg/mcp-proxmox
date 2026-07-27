package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ApiURL      string
	Username    string
	Password    string
	TokenID     string
	TokenSecret string
	ReadOnly    bool // defaults to true (secure by default)
	TLSInsecure bool // defaults to false (verify certificates)
	CAFile      string
	Timeout     time.Duration
	TaskTimeout int
}

func Load() *Config {
	return &Config{
		ApiURL:      getEnv("PROXMOX_API_URL", "https://localhost:8006/api2/json"),
		Username:    os.Getenv("PROXMOX_USERNAME"),
		Password:    os.Getenv("PROXMOX_PASSWORD"),
		TokenID:     os.Getenv("PROXMOX_TOKEN_ID"),
		TokenSecret: os.Getenv("PROXMOX_TOKEN_SECRET"),
		ReadOnly:    getEnvBool("PROXMOX_READ_ONLY", true),     // defaults to TRUE
		TLSInsecure: getEnvBool("PROXMOX_TLS_INSECURE", false), // defaults to FALSE
		CAFile:      os.Getenv("PROXMOX_CA_FILE"),
		Timeout:     getEnvDuration("PROXMOX_HTTP_TIMEOUT", 30*time.Second),
		TaskTimeout: getEnvInt("PROXMOX_TASK_TIMEOUT", 300),
	}
}

// HasCredentials reports whether a usable authentication method is configured.
func (c *Config) HasCredentials() bool {
	return (c.TokenID != "" && c.TokenSecret != "") || (c.Username != "" && c.Password != "")
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	if d, err := time.ParseDuration(val); err == nil && d > 0 {
		return d
	}
	// Bare numbers are a natural thing to write; read them as seconds.
	if secs, err := strconv.Atoi(val); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	log.Printf("WARNING: %s has unrecognized duration %q; using default %v", key, val, defaultVal)
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	if n, err := strconv.Atoi(val); err == nil && n > 0 {
		return n
	}
	log.Printf("WARNING: %s has unrecognized integer %q; using default %d", key, val, defaultVal)
	return defaultVal
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
