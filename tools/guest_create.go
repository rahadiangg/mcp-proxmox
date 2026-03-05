package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	px "github.com/Telmate/proxmox-api-go/proxmox"
)

func RegisterCreateTools(s *server.MCPServer, client *proxmox.Client) {
	getNextVmidTool := mcp.NewTool("get_next_vmid",
		mcp.WithDescription("Get the next available VM ID"),
		mcp.WithNumber("start_id",
			mcp.Description("Starting ID to search from (default: 100)"),
		),
	)
	s.AddTool(getNextVmidTool, getNextVmidHandler(client))
}

func getNextVmidHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startID := int(req.GetFloat("start_id", 100))
		startIDGuest := px.GuestID(startID)

		nextID, err := client.GetNextID(ctx, &startIDGuest)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get next VMID: %v", err)), nil
		}

		result := map[string]interface{}{
			"vmid": int(nextID),
			"next_id": int(nextID),
			"start_id": startID,
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}
