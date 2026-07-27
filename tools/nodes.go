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
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
	s.AddTool(listNodesTool, listNodesHandler(client))

	// Get node status
	getNodeStatusTool := mcp.NewTool("get_node_status",
		mcp.WithDescription("Get detailed status and health information for a specific node"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name (e.g., 'pve1')"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
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
		mcp.WithDestructiveHintAnnotation(true),
	)
	s.AddTool(rebootNodeTool, rebootNodeHandler(client))

	// Shutdown node - note: may not be available in all SDK versions
	shutdownNodeTool := mcp.NewTool("shutdown_node",
		mcp.WithDescription("Shutdown a specific node in the cluster - requires direct API access"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name to shutdown"),
		),
		mcp.WithDestructiveHintAnnotation(true),
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
		node, err := getRequiredNameParam(req, "node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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
		node, err := getRequiredNameParam(req, "node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Node reboot requires a direct API call - not exposed in the SDK.
		// Post without waiting: PostWithTask blocks polling the task until it
		// completes, but a rebooting node stops answering, so waiting can only
		// end in a timeout.
		url := fmt.Sprintf("/nodes/%s/status", node)
		params := map[string]interface{}{"command": "reboot"}
		if err := client.Post(ctx, params, url); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reboot node %s: %v", node, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Reboot requested for node %s. The node will stop answering the API while it restarts, so this cannot be confirmed here.", node)), nil
	}
}

func shutdownNodeHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node, err := getRequiredNameParam(req, "node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Same reasoning as reboot: a node that is shutting down cannot report
		// the completion of its own shutdown task.
		url := fmt.Sprintf("/nodes/%s/status", node)
		params := map[string]interface{}{"command": "shutdown"}
		if err := client.Post(ctx, params, url); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to shut down node %s: %v", node, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Shutdown requested for node %s. The node stops answering the API as it powers off, so this cannot be confirmed here.", node)), nil
	}
}
