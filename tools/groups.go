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

// RegisterGroupTools registers read-only group management tools
func RegisterGroupTools(s *server.MCPServer, client *proxmox.Client) {
	// List groups
	listGroupsTool := mcp.NewTool("list_groups",
		mcp.WithDescription("List all user groups in the cluster"),
	)
	s.AddTool(listGroupsTool, listGroupsHandler(client))

	// Get group
	getGroupTool := mcp.NewTool("get_group",
		mcp.WithDescription("Get group information"),
		mcp.WithString("groupid",
			mcp.Required(),
			mcp.Description("Group ID"),
		),
	)
	s.AddTool(getGroupTool, getGroupHandler(client))
}

// RegisterGroupWriteTools registers group write tools (requires PROXMOX_READ_ONLY=false)
func RegisterGroupWriteTools(s *server.MCPServer, client *proxmox.Client) {
	// Create group
	createGroupTool := mcp.NewTool("create_group",
		mcp.WithDescription("Create a new group"),
		mcp.WithString("groupid",
			mcp.Required(),
			mcp.Description("Group ID"),
		),
		mcp.WithString("comment",
			mcp.Description("Group description"),
		),
	)
	s.AddTool(createGroupTool, createGroupHandler(client))

	// Update group
	updateGroupTool := mcp.NewTool("update_group",
		mcp.WithDescription("Update group information"),
		mcp.WithString("groupid",
			mcp.Required(),
			mcp.Description("Group ID"),
		),
		mcp.WithString("comment",
			mcp.Required(),
			mcp.Description("Group description"),
		),
	)
	s.AddTool(updateGroupTool, updateGroupHandler(client))

	// Delete group
	deleteGroupTool := mcp.NewTool("delete_group",
		mcp.WithDescription("Delete a group"),
		mcp.WithString("groupid",
			mcp.Required(),
			mcp.Description("Group ID to delete"),
		),
	)
	s.AddTool(deleteGroupTool, deleteGroupHandler(client))
}

func listGroupsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawGroups, err := client.New().Group.List(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list groups: %v", err)), nil
		}

		groups := rawGroups.AsArray()
		result, _ := json.MarshalIndent(groups, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func getGroupHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupID := req.GetString("groupid", "")
		if groupID == "" {
			return mcp.NewToolResultError("groupid is required"), nil
		}

		rawGroup, err := client.New().Group.Read(ctx, px.GroupName(groupID))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get group: %v", err)), nil
		}

		group := rawGroup.Get()
		result, _ := json.MarshalIndent(group, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func createGroupHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupID := req.GetString("groupid", "")
		if groupID == "" {
			return mcp.NewToolResultError("groupid is required"), nil
		}

		config := px.ConfigGroup{
			Name: px.GroupName(groupID),
		}

		if comment := req.GetString("comment", ""); comment != "" {
			config.Comment = &comment
		}

		if err := client.New().Group.Create(ctx, config); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create group: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Group '%s' created", groupID)), nil
	}
}

func updateGroupHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupID := req.GetString("groupid", "")
		if groupID == "" {
			return mcp.NewToolResultError("groupid is required"), nil
		}

		comment := req.GetString("comment", "")

		config := px.ConfigGroup{
			Name:    px.GroupName(groupID),
			Comment: &comment,
		}

		if err := client.New().Group.Update(ctx, config); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update group: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Group '%s' updated", groupID)), nil
	}
}

func deleteGroupHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupID := req.GetString("groupid", "")
		if groupID == "" {
			return mcp.NewToolResultError("groupid is required"), nil
		}

		deleted, err := client.New().Group.Delete(ctx, px.GroupName(groupID))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete group: %v", err)), nil
		}

		if !deleted {
			return mcp.NewToolResultText(fmt.Sprintf("Group '%s' does not exist", groupID)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Group '%s' deleted", groupID)), nil
	}
}
