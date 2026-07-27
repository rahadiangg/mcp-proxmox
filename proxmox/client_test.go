package proxmox

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ticketHandler serves the login endpoint the SDK calls for password auth.
func ticketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ticket":"PVE:fake","CSRFPreventionToken":"csrf"}}`))
	})
}

// These tests replace an earlier suite that dialed example.com and
// 127.0.0.1:8006 for real, costing roughly 150s of TCP timeouts per run
// while asserting almost nothing.

func TestNewClientWithToken_MakesNoNetworkCall(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()

	client, err := NewClientWithToken(srv.URL+"/api2/json", "root@pam!mcp", "secret", DefaultOptions())
	if err != nil {
		t.Fatalf("NewClientWithToken() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClientWithToken() returned a nil client")
	}
	if hits != 0 {
		t.Errorf("token auth should not contact the server; got %d requests", hits)
	}
}

func TestNewClientWithToken_Validation(t *testing.T) {
	tests := []struct {
		name        string
		apiURL      string
		tokenID     string
		tokenSecret string
		wantErr     string
	}{
		{"empty url", "", "root@pam!mcp", "s", "API URL is required"},
		{"empty token id", "https://pve:8006/api2/json", "", "s", "token ID is required"},
		{"empty token secret", "https://pve:8006/api2/json", "root@pam!mcp", "", "token secret is required"},
		{"token id without realm", "https://pve:8006/api2/json", "root!mcp", "s", "parse token ID"},
		{"token id without token name", "https://pve:8006/api2/json", "root@pam", "s", "parse token ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientWithToken(tt.apiURL, tt.tokenID, tt.tokenSecret, DefaultOptions())
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err)
			}
		})
	}
}

func TestNewClientWithToken_ValidTokenFormats(t *testing.T) {
	srv := httptest.NewServer(ticketHandler())
	defer srv.Close()

	tokens := []string{"root@pam!1", "user@realm!42", "user-name@pve!100"}
	for _, tokenID := range tokens {
		t.Run(tokenID, func(t *testing.T) {
			client, err := NewClientWithToken(srv.URL+"/api2/json", tokenID, "secret", DefaultOptions())
			if err != nil {
				t.Fatalf("NewClientWithToken(%q) error = %v", tokenID, err)
			}
			if client == nil {
				t.Fatalf("NewClientWithToken(%q) returned a nil client", tokenID)
			}
		})
	}
}

func TestNewClient_PasswordAuth(t *testing.T) {
	srv := httptest.NewServer(ticketHandler())
	defer srv.Close()

	client, err := NewClient(srv.URL+"/api2/json", "root@pam", "hunter2", DefaultOptions())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned a nil client")
	}
}

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  string
	}{
		{"empty username", "", "pw", "username is required"},
		{"empty password", "root@pam", "", "password is required"},
		// The SDK slices the username at '@' unchecked, so a realm-less
		// username would panic later inside Client.New().
		{"username without realm", "root", "pw", "must include a realm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient("https://pve:8006/api2/json", tt.username, tt.password, DefaultOptions())
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err)
			}
		})
	}
}

func TestNewClient_LoginFailureIsReported(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"data":null,"message":"authentication failure"}`))
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL+"/api2/json", "root@pam", "wrong", DefaultOptions()); err == nil {
		t.Fatal("expected a login failure to be reported")
	}
}

// TestTLSVerification_OnByDefault pins the security default: an untrusted
// certificate must be rejected, with an error that tells the user how to fix it.
func TestTLSVerification_OnByDefault(t *testing.T) {
	srv := httptest.NewTLSServer(ticketHandler())
	defer srv.Close()

	_, err := NewClient(srv.URL+"/api2/json", "root@pam", "pw", DefaultOptions())
	if err == nil {
		t.Fatal("expected certificate verification to reject a self-signed server")
	}
	for _, want := range []string{"PROXMOX_CA_FILE", "PROXMOX_TLS_INSECURE"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("TLS error should mention %s; got: %v", want, err)
		}
	}
}

func TestTLSVerification_CanBeDisabledExplicitly(t *testing.T) {
	srv := httptest.NewTLSServer(ticketHandler())
	defer srv.Close()

	opts := DefaultOptions()
	opts.TLSInsecure = true

	if _, err := NewClient(srv.URL+"/api2/json", "root@pam", "pw", opts); err != nil {
		t.Fatalf("with TLSInsecure the self-signed server should be accepted, got: %v", err)
	}
}

// TestTLSVerification_CAFileIsTrusted covers the supported way to keep
// verification enabled against Proxmox's own self-signed CA.
func TestTLSVerification_CAFileIsTrusted(t *testing.T) {
	srv := httptest.NewTLSServer(ticketHandler())
	defer srv.Close()

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caPath, encoded, 0o600); err != nil {
		t.Fatalf("writing CA file: %v", err)
	}

	opts := DefaultOptions()
	opts.CAFile = caPath

	if _, err := NewClient(srv.URL+"/api2/json", "root@pam", "pw", opts); err != nil {
		t.Fatalf("server signed by the supplied CA should be trusted, got: %v", err)
	}
}

func TestCAFile_Errors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		opts := DefaultOptions()
		opts.CAFile = filepath.Join(t.TempDir(), "absent.pem")
		_, err := NewClientWithToken("https://pve:8006/api2/json", "root@pam!m", "s", opts)
		if err == nil || !strings.Contains(err.Error(), "reading CA file") {
			t.Errorf("expected a readable error for a missing CA file, got: %v", err)
		}
	})

	t.Run("garbage file", func(t *testing.T) {
		caPath := filepath.Join(t.TempDir(), "junk.pem")
		if err := os.WriteFile(caPath, []byte("not a certificate"), 0o600); err != nil {
			t.Fatalf("writing file: %v", err)
		}
		opts := DefaultOptions()
		opts.CAFile = caPath
		_, err := NewClientWithToken("https://pve:8006/api2/json", "root@pam!m", "s", opts)
		if err == nil || !strings.Contains(err.Error(), "no usable certificates") {
			t.Errorf("expected a readable error for a non-certificate CA file, got: %v", err)
		}
	})
}

// TestHTTPTimeoutIsEnforced covers the indefinite hang a missing timeout
// used to allow when a Proxmox node stopped responding.
func TestHTTPTimeoutIsEnforced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = w.Write([]byte(`{"data":{"ticket":"PVE:fake","CSRFPreventionToken":"csrf"}}`))
	}))
	defer srv.Close()

	opts := DefaultOptions()
	opts.Timeout = 50 * time.Millisecond

	start := time.Now()
	_, err := NewClient(srv.URL+"/api2/json", "root@pam", "pw", opts)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected the request to time out")
	}
	if elapsed > 400*time.Millisecond {
		t.Errorf("timeout was not enforced; call took %v", elapsed)
	}
}
