package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterHATools(s *server.MCPServer, client *proxmox.Client) {
	listHAGroupsTool := mcp.NewTool("list_ha_groups", mcp.WithDescription("List all high availability groups"))
	s.AddTool(listHAGroupsTool, listHAGroupsHandler(client))
}

func listHAGroupsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		haGroups, err := client.GetHAGroupList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list HA groups: %v", err)), nil
		}
		result, _ := json.MarshalIndent(haGroups, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}
