package proxmox

import (
	"context"
	"crypto/tls"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

// Client wraps the Proxmox SDK client
type Client struct {
	*proxmox.Client
}

// NewClient creates a new Proxmox client using username/password authentication
func NewClient(apiURL, username, password string) (*Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	client, err := proxmox.NewClient(apiURL, nil, "", tlsConfig, "", 60, false)
	if err != nil {
		return nil, err
	}

	if err := client.Login(context.Background(), username, password, ""); err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

// NewClientWithToken creates a new Proxmox client using API token authentication
func NewClientWithToken(apiURL, tokenID, tokenSecret string) (*Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	client, err := proxmox.NewClient(apiURL, nil, "", tlsConfig, "", 60, false)
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
