package proxmox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

// Options controls transport behavior for a Proxmox client.
type Options struct {
	// TLSInsecure disables certificate verification. Off by default: a
	// Proxmox API token travelling over an unverified connection can be
	// intercepted.
	TLSInsecure bool
	// CAFile is the PEM bundle to trust, typically a copy of the node's
	// /etc/pve/pve-root-ca.pem. This is the supported way to keep
	// verification on with Proxmox's default self-signed certificate.
	CAFile string
	// Timeout bounds a single HTTP request. Without it a hung node blocks a
	// tool worker forever.
	Timeout time.Duration
	// TaskTimeout bounds how long the SDK waits for an async task, in
	// seconds. Migrations and backups routinely exceed the SDK's own 60s.
	TaskTimeout int
}

// DefaultOptions returns transport defaults that fail closed.
func DefaultOptions() Options {
	return Options{
		TLSInsecure: false,
		Timeout:     30 * time.Second,
		TaskTimeout: 300,
	}
}

// Client wraps the Proxmox SDK client
type Client struct {
	*proxmox.Client
}

// newBase builds the underlying SDK client with our TLS and timeout policy.
func newBase(apiURL string, opts Options) (*proxmox.Client, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("API URL is required")
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: opts.TLSInsecure,
	}

	if opts.CAFile != "" {
		pem, err := os.ReadFile(opts.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA file %q: %w", opts.CAFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("CA file %q contains no usable certificates", opts.CAFile)
		}
		tlsConfig.RootCAs = pool
	}

	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.TaskTimeout <= 0 {
		opts.TaskTimeout = 300
	}

	httpClient := &http.Client{
		Timeout: opts.Timeout,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	return proxmox.NewClient(apiURL, httpClient, "", tlsConfig, "", opts.TaskTimeout, false)
}

// NewClient creates a new Proxmox client using username/password authentication
func NewClient(apiURL, username, password string, opts Options) (*Client, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required for password authentication")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required for password authentication")
	}
	// The SDK slices the username at '@' without checking, so a realm-less
	// username panics later inside Client.New(). Reject it up front.
	if !strings.Contains(username, "@") {
		return nil, fmt.Errorf("username %q must include a realm, e.g. %s@pam", username, username)
	}

	client, err := newBase(apiURL, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Login(context.Background(), username, password, ""); err != nil {
		return nil, WrapError("login", explainTLSError(apiURL, err))
	}

	return &Client{Client: client}, nil
}

// NewClientWithToken creates a new Proxmox client using API token authentication
func NewClientWithToken(apiURL, tokenID, tokenSecret string, opts Options) (*Client, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required for token authentication")
	}
	if tokenSecret == "" {
		return nil, fmt.Errorf("token secret is required for token authentication")
	}

	client, err := newBase(apiURL, opts)
	if err != nil {
		return nil, err
	}

	var tokenIDParsed proxmox.ApiTokenID
	if err := tokenIDParsed.Parse(tokenID); err != nil {
		return nil, WrapError("parse token ID", err)
	}

	client.SetAPIToken(tokenIDParsed, proxmox.ApiTokenSecret(tokenSecret))
	return &Client{Client: client}, nil
}

// explainTLSError turns an opaque x509 failure into an actionable message.
// Proxmox ships a self-signed certificate by default, so this is the error
// most users will hit first after verification became the default.
func explainTLSError(apiURL string, err error) error {
	if err == nil {
		return nil
	}
	var unknownAuthority x509.UnknownAuthorityError
	var hostnameErr x509.HostnameError
	var certInvalid x509.CertificateInvalidError

	isTLS := errors.As(err, &unknownAuthority) ||
		errors.As(err, &hostnameErr) ||
		errors.As(err, &certInvalid) ||
		strings.Contains(err.Error(), "x509")
	if !isTLS {
		return err
	}

	return fmt.Errorf("TLS verification failed for %s: %w\n"+
		"Proxmox VE uses a self-signed certificate by default. Either set "+
		"PROXMOX_CA_FILE to a copy of the node's /etc/pve/pve-root-ca.pem, or set "+
		"PROXMOX_TLS_INSECURE=true to skip verification (not recommended)", apiURL, err)
}
