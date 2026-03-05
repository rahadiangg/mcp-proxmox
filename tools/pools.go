package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterPoolTools(s *server.MCPServer, client *proxmox.Client) {
	listPoolsTool := mcp.NewTool("list_pools", mcp.WithDescription("List all resource pools in the cluster"))
	s.AddTool(listPoolsTool, listPoolsHandler(client))
}

func listPoolsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pools, err := client.GetPoolList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list pools: %v", err)), nil
		}
		result, _ := json.MarshalIndent(pools, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}
