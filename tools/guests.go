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

// RegisterGuestTools registers guest (VM/LXC) listing tools
func RegisterGuestTools(s *server.MCPServer, client *proxmox.Client) {
	// List guests
	listGuestsTool := mcp.NewTool("list_guests",
		mcp.WithDescription("List all VMs and containers in the Proxmox cluster"),
		mcp.WithString("node",
			mcp.Description("Filter by node name (optional)"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by guest type: 'qemu' or 'lxc' (optional)"),
		),
	)
	s.AddTool(listGuestsTool, listGuestsHandler(client))

	// Get guest info
	getGuestInfoTool := mcp.NewTool("get_guest_info",
		mcp.WithDescription("Get detailed information about a specific VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID"),
		),
	)
	s.AddTool(getGuestInfoTool, getGuestInfoHandler(client))

	// Get guest config
	getGuestConfigTool := mcp.NewTool("get_guest_config",
		mcp.WithDescription("Get configuration of a specific VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID"),
		),
	)
	s.AddTool(getGuestConfigTool, getGuestConfigHandler(client))

	// Get guest status
	getGuestStatusTool := mcp.NewTool("get_guest_status",
		mcp.WithDescription("Get current status (running/stopped) of a VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID"),
		),
	)
	s.AddTool(getGuestStatusTool, getGuestStatusHandler(client))

	// Get guest by name
	getGuestByNameTool := mcp.NewTool("get_guest_by_name",
		mcp.WithDescription("Find a guest (VM or container) by name"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Guest name to search for"),
		),
	)
	s.AddTool(getGuestByNameTool, getGuestByNameHandler(client))
}

func listGuestsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmsList, err := client.GetVmList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list guests: %v", err)), nil
		}

		// Extract the data from the response
		var result []map[string]interface{}
		if data, ok := vmsList["data"].([]interface{}); ok {
			for _, vm := range data {
				if vmMap, ok := vm.(map[string]interface{}); ok {
					result = append(result, vmMap)
				}
			}
		}

		// Apply filters if provided
		nodeFilter := req.GetString("node", "")
		typeFilter := req.GetString("type", "")

		if nodeFilter != "" {
			filtered := make([]map[string]interface{}, 0)
			for _, vm := range result {
				if node, ok := vm["node"].(string); ok && node == nodeFilter {
					filtered = append(filtered, vm)
				}
			}
			result = filtered
		}

		if typeFilter != "" {
			filtered := make([]map[string]interface{}, 0)
			for _, vm := range result {
				if vmType, ok := vm["type"].(string); ok && vmType == typeFilter {
					filtered = append(filtered, vm)
				}
			}
			result = filtered
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func getGuestInfoHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		info, err := client.GetVmInfo(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get guest info: %v", err)), nil
		}

		jsonResult, _ := json.MarshalIndent(info, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func getGuestConfigHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		config, err := client.GetVmConfig(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get guest config: %v", err)), nil
		}

		jsonResult, _ := json.MarshalIndent(config, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func getGuestStatusHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		status, err := client.GetVmState(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get guest status: %v", err)), nil
		}

		jsonResult, _ := json.MarshalIndent(status, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func getGuestByNameHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}

		vmsList, err := client.GetVmList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list guests: %v", err)), nil
		}

		// Extract the data from the response
		var guests []map[string]interface{}
		if data, ok := vmsList["data"].([]interface{}); ok {
			for _, vm := range data {
				if vmMap, ok := vm.(map[string]interface{}); ok {
					guests = append(guests, vmMap)
				}
			}
		}

		for _, guest := range guests {
			if guestName, ok := guest["name"].(string); ok && guestName == name {
				jsonResult, _ := json.MarshalIndent(guest, "", "  ")
				return mcp.NewToolResultText(string(jsonResult)), nil
			}
		}

		return mcp.NewToolResultError(fmt.Sprintf("Guest '%s' not found", name)), nil
	}
}
