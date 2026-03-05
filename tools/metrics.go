package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterMetricsTools(s *server.MCPServer, client *proxmox.Client) {
	listMetricsServersTool := mcp.NewTool("list_metrics_servers", mcp.WithDescription("List all metrics servers"))
	s.AddTool(listMetricsServersTool, listMetricsServersHandler(client))
}

func listMetricsServersHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		servers, err := client.GetMetricsServerList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list metrics servers: %v", err)), nil
		}
		result, _ := json.MarshalIndent(servers, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}
