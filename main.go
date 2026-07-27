package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	"github.com/rahadiangg/mcp-proxmox/config"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	"github.com/rahadiangg/mcp-proxmox/tools"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := godotenv.Load(); err != nil {
		// .env resolves against the working directory of whatever launched
		// this server, which for an MCP host is often not the project dir.
		// Say so rather than silently falling back to the process env.
		log.Printf("No .env file loaded (%v); using process environment", err)
	}

	cfg := config.Load()

	if !cfg.HasCredentials() {
		log.Fatalf("No Proxmox credentials configured. Set PROXMOX_TOKEN_ID and " +
			"PROXMOX_TOKEN_SECRET (recommended), or PROXMOX_USERNAME and PROXMOX_PASSWORD.")
	}

	if cfg.TLSInsecure {
		log.Printf("WARNING: TLS certificate verification is DISABLED (PROXMOX_TLS_INSECURE=true)")
	}

	opts := proxmox.Options{
		TLSInsecure: cfg.TLSInsecure,
		CAFile:      cfg.CAFile,
		Timeout:     cfg.Timeout,
		TaskTimeout: cfg.TaskTimeout,
	}

	var client *proxmox.Client
	var err error

	// Try API token auth first, fall back to password auth
	if cfg.TokenID != "" && cfg.TokenSecret != "" {
		client, err = proxmox.NewClientWithToken(cfg.ApiURL, cfg.TokenID, cfg.TokenSecret, opts)
	} else {
		client, err = proxmox.NewClient(cfg.ApiURL, cfg.Username, cfg.Password, opts)
	}
	if err != nil {
		log.Fatalf("Failed to create Proxmox client: %v", err)
	}

	s := server.NewMCPServer(
		"Proxmox MCP Server",
		version,
		server.WithToolCapabilities(true),
		// A panic in one handler must not take the whole server down with it.
		// The Proxmox SDK does unchecked type assertions on optional API
		// fields, so this is a real risk rather than a theoretical one.
		server.WithRecovery(),
	)

	// Log read-only mode status
	if !cfg.ReadOnly {
		log.Printf("WARNING: Write operations ENABLED - use with caution")
	} else {
		log.Printf("Starting in READ-ONLY mode (default) - write operations disabled")
	}

	// ALWAYS REGISTERED (read-only tools)
	tools.RegisterNodeTools(s, client)
	tools.RegisterGuestTools(s, client)
	tools.RegisterStorageTools(s, client)
	tools.RegisterPoolTools(s, client)
	tools.RegisterHATools(s, client)
	tools.RegisterMetricsTools(s, client)
	tools.RegisterUserTools(s, client)
	tools.RegisterGroupTools(s, client)         // read-only only
	tools.RegisterACMETools(s, client)          // read-only only
	tools.RegisterResourceTools(s, client)
	tools.RegisterStorageContentTools(s, client)
	tools.RegisterSnapshotTools(s, client)
	tools.RegisterQemuAgentTools(s, client)
	tools.RegisterNodeNetworkTools(s, client)
	tools.RegisterNetworkTools(s, client)
	tools.RegisterFirewallTools(s, client)
	tools.RegisterCreateTools(s, client)
	tools.RegisterDiskBandwidthTools(s, client)

	// CONDITIONALLY REGISTERED (write tools)
	if !cfg.ReadOnly {
		tools.RegisterLifecycleTools(s, client)
		tools.RegisterCloneTools(s, client)
		tools.RegisterDiskTools(s, client)
		tools.RegisterDiskBandwidthWriteTools(s, client)
		tools.RegisterMigrateTools(s, client)
		tools.RegisterBackupTools(s, client)
		tools.RegisterGroupWriteTools(s, client)
		tools.RegisterACMEWriteTools(s, client)
		tools.RegisterNodeWriteTools(s, client)
		log.Printf("Write operations enabled - 9 write tool categories registered")
	} else {
		log.Printf("Write operations disabled - 9 write tool categories skipped")
	}

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
