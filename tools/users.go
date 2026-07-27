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
	listUsersTool := mcp.NewTool("list_users", mcp.WithDescription("List all users in the cluster"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
	s.AddTool(listUsersTool, listUsersHandler(client))
}

func listUsersHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Read the endpoint directly. User.List().AsArray() returns values
		// whose only field is unexported, so marshalling them produced
		// [{},{}] -- every user's data was silently dropped.
		envelope, err := client.GetItemList(ctx, "/access/users")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list users: %v", err)), nil
		}

		users := []interface{}{}
		if data, ok := unwrapData(envelope).([]interface{}); ok {
			users = data
		}

		result, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to encode users: %v", err)), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}
