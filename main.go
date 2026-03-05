package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	"github.com/rahadiangg/mcp-proxmox/config"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	"github.com/rahadiangg/mcp-proxmox/tools"
)

func main() {
	godotenv.Load()

	cfg := config.Load()
	var client *proxmox.Client
	var err error

	// Try API token auth first, fall back to password auth
	if cfg.TokenID != "" && cfg.TokenSecret != "" {
		client, err = proxmox.NewClientWithToken(cfg.ApiURL, cfg.TokenID, cfg.TokenSecret)
	} else {
		client, err = proxmox.NewClient(cfg.ApiURL, cfg.Username, cfg.Password)
	}
	if err != nil {
		log.Fatalf("Failed to create Proxmox client: %v", err)
	}

	s := server.NewMCPServer(
		"Proxmox MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
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
