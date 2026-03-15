package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		defValue string
		expected string
	}{
		{
			name:     "env var set",
			envKey:   "TEST_VAR",
			envValue: "test_value",
			defValue: "default",
			expected: "test_value",
		},
		{
			name:     "env var not set",
			envKey:   "NON_EXISTENT_VAR",
			envValue: "",
			defValue: "default",
			expected: "default",
		},
		{
			name:     "env var empty string",
			envKey:   "EMPTY_VAR",
			envValue: "",
			defValue: "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q; want %q", tt.envKey, tt.defValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnv_Bool(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		defValue bool
		expected bool
	}{
		{
			name:     "env var set to true",
			envKey:   "TEST_BOOL",
			envValue: "true",
			defValue: false,
			expected: true,
		},
		{
			name:     "env var set to false",
			envKey:   "TEST_BOOL",
			envValue: "false",
			defValue: true,
			expected: false,
		},
		{
			name:     "env var not set",
			envKey:   "NON_EXISTENT_BOOL",
			envValue: "",
			defValue: true,
			expected: true,
		},
		{
			name:     "env var set to 1",
			envKey:   "TEST_BOOL",
			envValue: "1",
			defValue: false,
			expected: true,
		},
		{
			name:     "env var set to 0",
			envKey:   "TEST_BOOL",
			envValue: "0",
			defValue: true,
			expected: false,
		},
		{
			name:     "env var set to TRUE",
			envKey:   "TEST_BOOL",
			envValue: "TRUE",
			defValue: false,
			expected: true,
		},
		{
			name:     "env var set to FALSE",
			envKey:   "TEST_BOOL",
			envValue: "FALSE",
			defValue: true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvBool(tt.envKey, tt.defValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(%q, %v) = %v; want %v", tt.envKey, tt.defValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnv_NumericStrings(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		defValue string
		expected string
	}{
		{
			name:     "numeric string",
			envKey:   "TEST_NUM",
			envValue: "12345",
			defValue: "0",
			expected: "12345",
		},
		{
			name:     "float string",
			envKey:   "TEST_FLOAT",
			envValue: "12.34",
			defValue: "0",
			expected: "12.34",
		},
		{
			name:     "negative number",
			envKey:   "TEST_NEG",
			envValue: "-100",
			defValue: "0",
			expected: "-100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			result := getEnv(tt.envKey, tt.defValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q; want %q", tt.envKey, tt.defValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnv_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		defValue string
		expected string
	}{
		{
			name:     "url with special chars",
			envKey:   "TEST_URL",
			envValue: "https://example.com:8006/api2/json",
			defValue: "",
			expected: "https://example.com:8006/api2/json",
		},
		{
			name:     "string with spaces",
			envKey:   "TEST_SPACE",
			envValue: "hello world",
			defValue: "",
			expected: "hello world",
		},
		{
			name:     "path with slashes",
			envKey:   "TEST_PATH",
			envValue: "/path/to/file",
			defValue: "",
			expected: "/path/to/file",
		},
		{
			name:     "token with special chars",
			envKey:   "TEST_TOKEN",
			envValue: "user@realm!token",
			defValue: "",
			expected: "user@realm!token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			result := getEnv(tt.envKey, tt.defValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q; want %q", tt.envKey, tt.defValue, result, tt.expected)
			}
		})
	}
}
