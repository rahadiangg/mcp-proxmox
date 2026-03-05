package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterUserTools(s *server.MCPServer, client *proxmox.Client) {
	listUsersTool := mcp.NewTool("list_users", mcp.WithDescription("List all users in the cluster"))
	s.AddTool(listUsersTool, listUsersHandler(client))
}

func listUsersHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawUsers, err := client.New().User.List(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list users: %v", err)), nil
		}

		users := rawUsers.AsArray()
		result, _ := json.MarshalIndent(users, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}
