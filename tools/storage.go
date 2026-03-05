package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterStorageTools(s *server.MCPServer, client *proxmox.Client) {
	// List storage
	listStorageTool := mcp.NewTool("list_storage", mcp.WithDescription("List all storage in the cluster"))
	s.AddTool(listStorageTool, listStorageHandler(client))

	// Get storage status
	getStorageStatusTool := mcp.NewTool("get_storage_status",
		mcp.WithDescription("Get storage usage statistics including available space, used space, and usage percentage"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name where the storage is located (e.g., 'pve1')"),
		),
		mcp.WithString("storage",
			mcp.Required(),
			mcp.Description("Storage name (e.g., 'local', 'local-lvm')"),
		),
	)
	s.AddTool(getStorageStatusTool, getStorageStatusHandler(client))

	// Get storage config
	getStorageConfigTool := mcp.NewTool("get_storage_config",
		mcp.WithDescription("Get storage configuration details including type, content types, and settings"),
		mcp.WithString("storage",
			mcp.Required(),
			mcp.Description("Storage name (e.g., 'local', 'local-lvm')"),
		),
	)
	s.AddTool(getStorageConfigTool, getStorageConfigHandler(client))
}

func listStorageHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storageList, err := client.GetStorageList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list storage: %v", err)), nil
		}
		result, _ := json.MarshalIndent(storageList, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func getStorageStatusHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := req.GetString("node", "")
		if node == "" {
			return mcp.NewToolResultError("node is required"), nil
		}

		storage := req.GetString("storage", "")
		if storage == "" {
			return mcp.NewToolResultError("storage is required"), nil
		}

		url := fmt.Sprintf("/nodes/%s/storage/%s/status", node, storage)
		status, err := client.GetItemConfigMapStringInterface(ctx, url, "storage", "STATUS")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get storage status: %v", err)), nil
		}

		result, _ := json.MarshalIndent(status, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func getStorageConfigHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		storage := req.GetString("storage", "")
		if storage == "" {
			return mcp.NewToolResultError("storage is required"), nil
		}

		url := fmt.Sprintf("/storage/%s", storage)
		config, err := client.GetItemConfigMapStringInterface(ctx, url, "storage", "CONFIG")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get storage config: %v", err)), nil
		}

		result, _ := json.MarshalIndent(config, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}
