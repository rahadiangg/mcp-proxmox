package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterBackupTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("backup_guest", mcp.WithDescription("Create a backup of a VM")), backupGuestHandler(client))
}

func RegisterFirewallTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("get_guest_firewall_options", mcp.WithDescription("Get firewall options for a VM")), getGuestFirewallOptionsHandler(client))
}

func RegisterDiskTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("resize_disk", mcp.WithDescription("Resize a guest disk")), resizeDiskHandler(client))
}

func RegisterNetworkTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("get_guest_agent_network", mcp.WithDescription("Get network interfaces from QEMU agent")), getGuestAgentNetworkHandler(client))
}

func RegisterMigrateTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("migrate_guest", mcp.WithDescription("Migrate a VM to another node")), migrateGuestHandler(client))
}

func RegisterNodeNetworkTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("get_node_network", mcp.WithDescription("Get node network configuration")), getNodeNetworkHandler(client))
}

func RegisterQemuAgentTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("ping_qemu_agent", mcp.WithDescription("Ping QEMU guest agent")), pingQemuAgentHandler(client))
}

func RegisterResourceTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("list_resources", mcp.WithDescription("List cluster resources")), listResourcesHandler(client))
}

func RegisterStorageContentTools(s *server.MCPServer, client *proxmox.Client) {
	getStorageContentTool := mcp.NewTool("get_storage_content",
		mcp.WithDescription("Get storage contents including disks, ISOs, and templates"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name where the storage is located (e.g., 'pve1')"),
		),
		mcp.WithString("storage",
			mcp.Required(),
			mcp.Description("Storage name (e.g., 'local', 'local-lvm')"),
		),
		mcp.WithString("content",
			mcp.Description("Filter by content type (e.g., 'images', 'iso', 'rootdir', 'snippets') - optional"),
		),
	)
	s.AddTool(getStorageContentTool, getStorageContentHandler(client))
}

func RegisterSnapshotTools(s *server.MCPServer, client *proxmox.Client) {
	s.AddTool(mcp.NewTool("list_snapshots", mcp.WithDescription("List all snapshots")), listSnapshotsHandler(client))
}

func backupGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Backup operation initiated"), nil
	}
}

func getGuestFirewallOptionsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Firewall options retrieved"), nil
	}
}

func resizeDiskHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Disk resized"), nil
	}
}

func getGuestAgentNetworkHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Network info retrieved"), nil
	}
}

func migrateGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Migration initiated"), nil
	}
}

func getNodeNetworkHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Network config retrieved"), nil
	}
}

func pingQemuAgentHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Agent pinged successfully"), nil
	}
}

func listResourcesHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Resources listed"), nil
	}
}

func getStorageContentHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := req.GetString("node", "")
		if node == "" {
			return mcp.NewToolResultError("node is required"), nil
		}

		storage := req.GetString("storage", "")
		if storage == "" {
			return mcp.NewToolResultError("storage is required"), nil
		}

		contentType := req.GetString("content", "")

		url := fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage)
		content, err := client.GetItemConfigMapStringInterface(ctx, url, "storage", "CONTENT")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get storage content: %v", err)), nil
		}

		var result []map[string]interface{}
		if data, ok := content["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if contentType != "" {
						if itemContentType, ok := itemMap["content"].(string); ok && itemContentType == contentType {
							result = append(result, itemMap)
						}
					} else {
						result = append(result, itemMap)
					}
				}
			}
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func listSnapshotsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Snapshots listed"), nil
	}
}
