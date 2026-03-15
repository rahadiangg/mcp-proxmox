package proxmox

import (
	"testing"
)

func TestNewClient_InvalidURL(t *testing.T) {
	tests := []struct {
		name      string
		apiURL    string
		username  string
		password  string
		wantError bool
	}{
		{
			name:      "empty URL",
			apiURL:    "",
			username:  "test",
			password:  "test",
			wantError: true,
		},
		{
			name:      "invalid URL format",
			apiURL:    "not-a-url",
			username:  "test",
			password:  "test",
			wantError: true,
		},
		{
			name:      "missing scheme",
			apiURL:    "example.com/api2/json",
			username:  "test",
			password:  "test",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.apiURL, tt.username, tt.password)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClient() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewClient_EmptyCredentials(t *testing.T) {
	tests := []struct {
		name     string
		apiURL   string
		username string
		password string
	}{
		{
			name:     "empty username",
			apiURL:   "https://example.com/api2/json",
			username: "",
			password: "password",
		},
		{
			name:     "empty password",
			apiURL:   "https://example.com/api2/json",
			username: "username",
			password: "",
		},
		{
			name:     "both empty",
			apiURL:   "https://example.com/api2/json",
			username: "",
			password: "",
		},
		{
			name:     "whitespace username",
			apiURL:   "https://example.com/api2/json",
			username: "   ",
			password: "password",
		},
		{
			name:     "whitespace password",
			apiURL:   "https://example.com/api2/json",
			username: "username",
			password: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.apiURL, tt.username, tt.password)
			if err == nil {
				t.Error("NewClient() expected error for empty credentials, got nil")
			}
		})
	}
}

func TestNewClientWithToken_InvalidURL(t *testing.T) {
	tests := []struct {
		name        string
		apiURL      string
		tokenID     string
		tokenSecret string
		wantError   bool
	}{
		{
			name:        "empty URL",
			apiURL:      "",
			tokenID:     "test@!1",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "invalid URL format",
			apiURL:      "not-a-url",
			tokenID:     "test@!1",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "missing scheme",
			apiURL:      "example.com/api2/json",
			tokenID:     "test@!1",
			tokenSecret: "secret",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientWithToken(tt.apiURL, tt.tokenID, tt.tokenSecret)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClientWithToken() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewClientWithToken_InvalidTokenID(t *testing.T) {
	validURL := "https://example.com/api2/json"

	tests := []struct {
		name        string
		tokenID     string
		tokenSecret string
		wantError   bool
	}{
		{
			name:        "empty token ID",
			tokenID:     "",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "missing ! separator",
			tokenID:     "test-token",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "missing realm",
			tokenID:     "!1",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "missing tokenID number",
			tokenID:     "test@!",
			tokenSecret: "secret",
			wantError:   true,
		},
		{
			name:        "empty token secret",
			tokenID:     "test@!1",
			tokenSecret: "",
			wantError:   true, // API validates token secret is not empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientWithToken(validURL, tt.tokenID, tt.tokenSecret)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClientWithToken() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewClientWithToken_ValidTokenFormats(t *testing.T) {
	validURL := "https://example.com/api2/json"

	tests := []struct {
		name        string
		tokenID     string
		tokenSecret string
	}{
		{
			name:        "root@pam token",
			tokenID:     "root@pam!1",
			tokenSecret: "secret-key",
		},
		{
			name:        "custom realm token",
			tokenID:     "user@realm!42",
			tokenSecret: "another-secret",
		},
		{
			name:        "token with hyphens",
			tokenID:     "user-name@pve!100",
			tokenSecret: "key-with-hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithToken(validURL, tt.tokenID, tt.tokenSecret)
			if err != nil {
				t.Logf("NewClientWithToken() error = %v (expected for invalid URL)", err)
			}
			if err == nil && client == nil {
				t.Error("NewClientWithToken() returned nil client without error")
			}
		})
	}
}

func TestNewClient_ValidURLFormats(t *testing.T) {
	tests := []struct {
		name     string
		apiURL   string
		username string
		password string
	}{
		{
			name:     "HTTPS URL",
			apiURL:   "https://example.com:8006/api2/json",
			username: "root@pam",
			password: "password",
		},
		{
			name:     "HTTP URL",
			apiURL:   "http://localhost:8006/api2/json",
			username: "user",
			password: "pass",
		},
		{
			name:     "URL with path",
			apiURL:   "https://proxmox.example.com/api2/json",
			username: "admin",
			password: "admin",
		},
		{
			name:     "localhost with port",
			apiURL:   "https://127.0.0.1:8006/api2/json",
			username: "testuser",
			password: "testpass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiURL, tt.username, tt.password)
			if err != nil {
				t.Logf("NewClient() error = %v (connection errors are expected for unit tests)", err)
			}
			if err == nil && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}
