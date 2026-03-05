package config

import (
	"os"
	"testing"
)

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name      string
		setEnv    map[string]string
		key       string
		defaultVal bool
		want      bool
	}{
		// Truthy values
		{"true lowercase", map[string]string{"TEST": "true"}, "TEST", false, true},
		{"true uppercase", map[string]string{"TEST": "TRUE"}, "TEST", false, true},
		{"true mixed case", map[string]string{"TEST": "TrUe"}, "TEST", false, true},
		{"1", map[string]string{"TEST": "1"}, "TEST", false, true},
		{"yes lowercase", map[string]string{"TEST": "yes"}, "TEST", false, true},
		{"yes uppercase", map[string]string{"TEST": "YES"}, "TEST", false, true},
		{"on lowercase", map[string]string{"TEST": "on"}, "TEST", false, true},
		{"on uppercase", map[string]string{"TEST": "ON"}, "TEST", false, true},
		{"enabled", map[string]string{"TEST": "enabled"}, "TEST", false, true},
		{"y", map[string]string{"TEST": "y"}, "TEST", false, true},
		{"t", map[string]string{"TEST": "t"}, "TEST", false, true},

		// Falsy values
		{"false lowercase", map[string]string{"TEST": "false"}, "TEST", true, false},
		{"false uppercase", map[string]string{"TEST": "FALSE"}, "TEST", true, false},
		{"0", map[string]string{"TEST": "0"}, "TEST", true, false},
		{"no", map[string]string{"TEST": "no"}, "TEST", true, false},
		{"off", map[string]string{"TEST": "off"}, "TEST", true, false},

		// Whitespace handling
		{"true with spaces", map[string]string{"TEST": " true "}, "TEST", false, true},
		{"false with tabs", map[string]string{"TEST": "\tfalse\t"}, "TEST", true, false},

		// Empty/unset - returns default
		{"empty string uses default", map[string]string{"TEST": ""}, "TEST", true, true},
		{"unset env uses default", nil, "UNSET_VAR", true, true},
		{"unset env default false", nil, "UNSET_VAR", false, false},

		// Invalid values - default to false (safe behavior)
		{"invalid value", map[string]string{"TEST": "invalid"}, "TEST", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env before test
			for k := range tt.setEnv {
				os.Unsetenv(k)
			}

			// Set env vars
			for k, v := range tt.setEnv {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.setEnv {
					os.Unsetenv(k)
				}
			}()

			got := getEnvBool(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getEnvBool(%q, %v) = %v, want %v", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Clean up env before test
	os.Unsetenv("PROXMOX_READ_ONLY")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	t.Run("default ReadOnly to true when env not set", func(t *testing.T) {
		cfg := Load()
		if cfg.ReadOnly != true {
			t.Errorf("Load() ReadOnly = %v, want true (default)", cfg.ReadOnly)
		}
	})

	t.Run("ReadOnly from env true", func(t *testing.T) {
		os.Setenv("PROXMOX_READ_ONLY", "true")
		defer os.Unsetenv("PROXMOX_READ_ONLY")

		cfg := Load()
		if cfg.ReadOnly != true {
			t.Errorf("Load() ReadOnly = %v, want true", cfg.ReadOnly)
		}
	})

	t.Run("ReadOnly from env false", func(t *testing.T) {
		os.Setenv("PROXMOX_READ_ONLY", "false")
		defer os.Unsetenv("PROXMOX_READ_ONLY")

		cfg := Load()
		if cfg.ReadOnly != false {
			t.Errorf("Load() ReadOnly = %v, want false", cfg.ReadOnly)
		}
	})

	t.Run("ReadOnly from env 1", func(t *testing.T) {
		os.Setenv("PROXMOX_READ_ONLY", "1")
		defer os.Unsetenv("PROXMOX_READ_ONLY")

		cfg := Load()
		if cfg.ReadOnly != true {
			t.Errorf("Load() ReadOnly = %v, want true", cfg.ReadOnly)
		}
	})

	t.Run("ReadOnly from env 0", func(t *testing.T) {
		os.Setenv("PROXMOX_READ_ONLY", "0")
		defer os.Unsetenv("PROXMOX_READ_ONLY")

		cfg := Load()
		if cfg.ReadOnly != false {
			t.Errorf("Load() ReadOnly = %v, want false", cfg.ReadOnly)
		}
	})
}
