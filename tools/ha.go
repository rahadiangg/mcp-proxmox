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
		// Deliberately not client.GetHAGroupList: it type-asserts optional
		// fields (comment, nofailback, restricted) without a comma-ok, so a
		// group created without a comment panics inside the SDK. This handler
		// only serializes to JSON, so the typed struct buys nothing.
		list, err := client.GetItemList(ctx, "/cluster/ha/groups")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list HA groups: %v", err)), nil
		}

		groups := []interface{}{}
		if data, ok := list["data"].([]interface{}); ok {
			groups = data
		}

		result, err := json.MarshalIndent(groups, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to encode HA groups: %v", err)), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}
