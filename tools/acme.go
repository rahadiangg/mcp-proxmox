package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

func RegisterACMETools(s *server.MCPServer, client *proxmox.Client) {
	listACMEAccountsTool := mcp.NewTool("list_acme_accounts",
		mcp.WithDescription("List all ACME accounts"))
	s.AddTool(listACMEAccountsTool, listACMEAccountsHandler(client))

	getACMEAccountTool := mcp.NewTool("get_acme_account",
		mcp.WithDescription("Get ACME account information"),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("ACME account name"),
		),
	)
	s.AddTool(getACMEAccountTool, getACMEAccountHandler(client))

	listACMEPluginsTool := mcp.NewTool("list_acme_plugins",
		mcp.WithDescription("List all ACME plugins"))
	s.AddTool(listACMEPluginsTool, listACMEPluginsHandler(client))
}

// RegisterACMEWriteTools registers ACME write tools (requires PROXMOX_READ_ONLY=false)
func RegisterACMEWriteTools(s *server.MCPServer, client *proxmox.Client) {
	deleteACMEAccountTool := mcp.NewTool("delete_acme_account",
		mcp.WithDescription("Delete an ACME account"),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("ACME account name"),
		),
	)
	s.AddTool(deleteACMEAccountTool, deleteACMEAccountHandler(client))

	deleteACMEPluginTool := mcp.NewTool("delete_acme_plugin",
		mcp.WithDescription("Delete an ACME plugin"),
		mcp.WithString("plugin",
			mcp.Required(),
			mcp.Description("ACME plugin name"),
		),
	)
	s.AddTool(deleteACMEPluginTool, deleteACMEPluginHandler(client))
}

func listACMEAccountsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		accounts, err := client.GetAcmeAccountList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list ACME accounts: %v", err)), nil
		}
		result, _ := json.MarshalIndent(accounts, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func getACMEAccountHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("account is required"), nil
		}
		acmeAccount, err := client.GetAcmeAccountConfig(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get ACME account: %v", err)), nil
		}
		result, _ := json.MarshalIndent(acmeAccount, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func deleteACMEAccountHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("account is required"), nil
		}
		_, err := client.DeleteAcmeAccount(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete ACME account: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("ACME account '%s' deleted", account)), nil
	}
}

func listACMEPluginsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		plugins, err := client.GetAcmePluginList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list ACME plugins: %v", err)), nil
		}
		result, _ := json.MarshalIndent(plugins, "", "  ")
		return mcp.NewToolResultText(string(result)), nil
	}
}

func deleteACMEPluginHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		plugin := req.GetString("plugin", "")
		if plugin == "" {
			return mcp.NewToolResultError("plugin is required"), nil
		}
		err := client.DeleteAcmePlugin(ctx, plugin)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete ACME plugin: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("ACME plugin '%s' deleted", plugin)), nil
	}
}
