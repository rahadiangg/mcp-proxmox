package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// RegisterNodeTools registers read-only node-related tools
func RegisterNodeTools(s *server.MCPServer, client *proxmox.Client) {
	// List nodes
	listNodesTool := mcp.NewTool("list_nodes",
		mcp.WithDescription("List all nodes in the Proxmox cluster"),
	)
	s.AddTool(listNodesTool, listNodesHandler(client))

	// Get node status
	getNodeStatusTool := mcp.NewTool("get_node_status",
		mcp.WithDescription("Get detailed status and health information for a specific node"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name (e.g., 'pve1')"),
		),
	)
	s.AddTool(getNodeStatusTool, getNodeStatusHandler(client))
}

// RegisterNodeWriteTools registers node write tools (requires PROXMOX_READ_ONLY=false)
func RegisterNodeWriteTools(s *server.MCPServer, client *proxmox.Client) {
	// Reboot node - note: may not be available in all SDK versions
	rebootNodeTool := mcp.NewTool("reboot_node",
		mcp.WithDescription("Reboot a specific node in the cluster - requires direct API access"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name to reboot"),
		),
	)
	s.AddTool(rebootNodeTool, rebootNodeHandler(client))

	// Shutdown node - note: may not be available in all SDK versions
	shutdownNodeTool := mcp.NewTool("shutdown_node",
		mcp.WithDescription("Shutdown a specific node in the cluster - requires direct API access"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name to shutdown"),
		),
	)
	s.AddTool(shutdownNodeTool, shutdownNodeHandler(client))
}

func listNodesHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodes, err := client.GetNodeList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list nodes: %v", err)), nil
		}

		result, _ := json.MarshalIndent(nodes, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func getNodeStatusHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := req.GetString("node", "")
		if node == "" {
			return mcp.NewToolResultError("node is required"), nil
		}

		// Get node status via GetItemConfigMapStringInterface
		url := fmt.Sprintf("/nodes/%s/status", node)
		status, err := client.GetItemConfigMapStringInterface(ctx, url, "node", "STATUS")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get node status: %v", err)), nil
		}

		result, _ := json.MarshalIndent(status, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func rebootNodeHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := req.GetString("node", "")
		if node == "" {
			return mcp.NewToolResultError("node is required"), nil
		}

		// Node reboot requires direct API call - not directly exposed in SDK
		url := fmt.Sprintf("/nodes/%s/status", node)
		params := map[string]interface{}{"command": "reboot"}
		upid, err := client.PostWithTask(ctx, params, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reboot node: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Node %s reboot initiated. UPID: %s", node, upid)), nil
	}
}

func shutdownNodeHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := req.GetString("node", "")
		if node == "" {
			return mcp.NewToolResultError("node is required"), nil
		}

		// Node shutdown requires direct API call - not directly exposed in SDK
		url := fmt.Sprintf("/nodes/%s/status", node)
		params := map[string]interface{}{"command": "shutdown"}
		upid, err := client.PostWithTask(ctx, params, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to shutdown node: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Node %s shutdown initiated. UPID: %s", node, upid)), nil
	}
}
