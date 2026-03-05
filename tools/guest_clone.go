package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	px "github.com/Telmate/proxmox-api-go/proxmox"
)

// RegisterCloneTools registers guest cloning tools
func RegisterCloneTools(s *server.MCPServer, client *proxmox.Client) {
	// Clone QEMU VM
	cloneQemuVmTool := mcp.NewTool("clone_qemu_vm",
		mcp.WithDescription("Clone a QEMU VM to a new VM"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("Source VM ID to clone"),
		),
		mcp.WithNumber("new_id",
			mcp.Required(),
			mcp.Description("New VM ID for the clone"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name for the cloned VM"),
		),
		mcp.WithString("node",
			mcp.Description("Target node (defaults to source node)"),
		),
		mcp.WithBoolean("full",
			mcp.Description("Create a full clone (true) or linked clone (false)"),
		),
	)
	s.AddTool(cloneQemuVmTool, cloneQemuVmHandler(client))

	// Clone LXC container
	cloneLxcContainerTool := mcp.NewTool("clone_lxc_container",
		mcp.WithDescription("Clone an LXC container to a new container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("Source container ID to clone"),
		),
		mcp.WithNumber("new_id",
			mcp.Required(),
			mcp.Description("New container ID for the clone"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name for the cloned container"),
		),
		mcp.WithString("node",
			mcp.Description("Target node (defaults to source node)"),
		),
		mcp.WithBoolean("full",
			mcp.Description("Create a full clone (true) or linked clone (false)"),
		),
	)
	s.AddTool(cloneLxcContainerTool, cloneLxcContainerHandler(client))

	// Create template
	createTemplateTool := mcp.NewTool("create_template",
		mcp.WithDescription("Convert a VM to a template"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID to convert to template"),
		),
	)
	s.AddTool(createTemplateTool, createTemplateHandler(client))
}

func cloneQemuVmHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		newID := int(req.GetFloat("new_id", 0))
		if newID == 0 {
			return mcp.NewToolResultError("new_id is required"), nil
		}

		name := req.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}

		// Get source VM info to determine node
		sourceVmr := px.NewVmRef(px.GuestID(vmid))
		sourceInfo, err := client.GetVmInfo(ctx, sourceVmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get source VM info: %v", err)), nil
		}

		node := req.GetString("node", "")
		if node == "" {
			// Get node from source VM
			if nodeVal, ok := sourceInfo["node"].(string); ok {
				node = nodeVal
			}
		}

		// Clone options
		fullClone := 1
		if !req.GetBool("full", true) {
			fullClone = 0
		}

		cloneParams := map[string]interface{}{
			"newid": newID,
			"name":  name,
			"full":  fullClone,
			"target": node,
		}

		upid, err := client.CloneQemuVm(ctx, sourceVmr, cloneParams)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to clone VM: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("VM %d cloned to %d (%s). UPID: %s", vmid, newID, name, upid)), nil
	}
}

func cloneLxcContainerHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		newID := int(req.GetFloat("new_id", 0))
		if newID == 0 {
			return mcp.NewToolResultError("new_id is required"), nil
		}

		name := req.GetString("name", "")
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}

		// Get source container info to determine node
		sourceVmr := px.NewVmRef(px.GuestID(vmid))
		sourceInfo, err := client.GetVmInfo(ctx, sourceVmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get source container info: %v", err)), nil
		}

		node := req.GetString("node", "")
		if node == "" {
			if nodeVal, ok := sourceInfo["node"].(string); ok {
				node = nodeVal
			}
		}

		fullClone := 1
		if !req.GetBool("full", true) {
			fullClone = 0
		}

		cloneParams := map[string]interface{}{
			"newid":    newID,
			"hostname": name,
			"full":     fullClone,
			"target":   node,
		}

		upid, err := client.CloneLxcContainer(ctx, sourceVmr, cloneParams)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to clone container: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Container %d cloned to %d (%s). UPID: %s", vmid, newID, name, upid)), nil
	}
}

func createTemplateHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))

		config := map[string]interface{}{
			"template": 1,
		}

		_, err := client.SetVmConfig(vmr, config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert to template: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("VM %d converted to template", vmid)), nil
	}
}
