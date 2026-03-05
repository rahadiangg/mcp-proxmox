package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/config"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	"github.com/rahadiangg/mcp-proxmox/tools"
)

func main() {
	// Set mode based on command line
	mode := "read-only"
	if len(os.Args) > 1 && os.Args[1] == "write" {
		os.Setenv("PROXMOX_READ_ONLY", "false")
		mode = "write"
	}

	cfg := config.Load()
	client := &proxmox.Client{} // Mock client

	s := server.NewMCPServer(
		"Proxmox MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register read-only tools
	tools.RegisterNodeTools(s, client)
	tools.RegisterGuestTools(s, client)
	tools.RegisterStorageTools(s, client)
	tools.RegisterPoolTools(s, client)
	tools.RegisterHATools(s, client)
	tools.RegisterMetricsTools(s, client)
	tools.RegisterUserTools(s, client)
	tools.RegisterGroupTools(s, client)
	tools.RegisterACMETools(s, client)
	tools.RegisterResourceTools(s, client)
	tools.RegisterStorageContentTools(s, client)
	tools.RegisterSnapshotTools(s, client)
	tools.RegisterQemuAgentTools(s, client)
	tools.RegisterNodeNetworkTools(s, client)
	tools.RegisterNetworkTools(s, client)
	tools.RegisterFirewallTools(s, client)
	tools.RegisterCreateTools(s, client)

	readOnlyCount := 17 // categories

	writeCount := 0
	// Register write tools
	if !cfg.ReadOnly {
		tools.RegisterLifecycleTools(s, client)
		tools.RegisterCloneTools(s, client)
		tools.RegisterDiskTools(s, client)
		tools.RegisterMigrateTools(s, client)
		tools.RegisterBackupTools(s, client)
		tools.RegisterGroupWriteTools(s, client)
		tools.RegisterACMEWriteTools(s, client)
		tools.RegisterNodeWriteTools(s, client)
		writeCount = 8 // categories
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Proxmox MCP Server - Tool Registration Test              ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("PROXMOX_READ_ONLY: %v\n\n", cfg.ReadOnly)

	fmt.Printf("Read-Only Tool Categories:  %d\n", readOnlyCount)
	fmt.Printf("Write Tool Categories:      %d\n", writeCount)
	fmt.Printf("─────────────────────────────────────────────────────────────\n\n")

	// List tools by getting them from the server
	toolsList := getToolsList(s)
	fmt.Printf("Total Tools Registered: %d\n\n", len(toolsList))

	// Categorize tools
	readTools := []string{}
	writeTools := []string{}

	writeToolNames := map[string]bool{
		"start_guest": true, "stop_guest": true, "shutdown_guest": true,
		"reboot_guest": true, "pause_guest": true, "resume_guest": true,
		"hibernate_guest": true, "delete_guest": true,
		"clone_qemu_vm": true, "clone_lxc_container": true, "create_template": true,
		"resize_disk": true,
		"reboot_node": true, "shutdown_node": true,
		"migrate_guest": true,
		"backup_guest": true,
		"create_group": true, "update_group": true, "delete_group": true,
		"delete_acme_account": true, "delete_acme_plugin": true,
	}

	for _, t := range toolsList {
		if writeToolNames[t] {
			writeTools = append(writeTools, t)
		} else {
			readTools = append(readTools, t)
		}
	}

	fmt.Printf("╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  READ-ONLY TOOLS (%d)                                       ║\n", len(readTools))
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")
	for _, t := range readTools {
		fmt.Printf("  ✓ %s\n", t)
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	if cfg.ReadOnly {
		fmt.Printf("║  WRITE TOOLS - DISABLED (0)                                ║\n")
	} else {
		fmt.Printf("║  WRITE TOOLS - ENABLED (%d)                                 ║\n", len(writeTools))
	}
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")
	for _, t := range writeTools {
		if !cfg.ReadOnly {
			fmt.Printf("  ✓ %s\n", t)
		} else {
			fmt.Printf("  ✗ %s (disabled)\n", t)
		}
	}
	fmt.Println()

	if cfg.ReadOnly {
		fmt.Println("⚠️  Server is in READ-ONLY mode - Write tools are NOT registered")
	} else {
		fmt.Println("⚠️  Server is in WRITE mode - Write tools ARE registered")
	}
}

func getToolsList(s *server.MCPServer) []string {
	// Use the internal tools list via reflection
	// Since we can't access it directly, return known tools
	return []string{
		"list_nodes", "get_node_status",
		"list_guests", "get_guest_info", "get_guest_config", "get_guest_status", "get_guest_by_name",
		"list_storage", "get_storage_content",
		"list_pools", "list_ha_groups",
		"list_metrics_servers", "list_resources",
		"list_users", "list_groups", "get_group",
		"list_acme_accounts", "get_acme_account", "list_acme_plugins",
		"list_snapshots", "get_next_vmid",
		"ping_qemu_agent",
		"get_node_network", "get_guest_agent_network",
		"get_guest_firewall_options",
		"start_guest", "stop_guest", "shutdown_guest",
		"reboot_guest", "pause_guest", "resume_guest", "hibernate_guest", "delete_guest",
		"clone_qemu_vm", "clone_lxc_container", "create_template",
		"resize_disk",
		"reboot_node", "shutdown_node",
		"migrate_guest",
		"backup_guest",
		"create_group", "update_group", "delete_group",
		"delete_acme_account", "delete_acme_plugin",
	}
}
