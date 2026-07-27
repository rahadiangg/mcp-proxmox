package tools

import (
	"context"
	"fmt"
	"strconv"

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

		// CloneLxcContainer builds its URL from vmParams["vmid"], not from the
		// VmRef, so omitting this key produced /nodes/<node>/lxc/%!s(<nil>)/clone
		// and every LXC clone failed. It formats the value with %s, so this
		// must be a string -- an int renders as %!s(int=200).
		cloneParams := map[string]interface{}{
			"vmid":     strconv.Itoa(vmid),
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

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// CreateTemplate posts /nodes/{node}/{type}/{vmid}/template, which is
		// the documented conversion path. The previous implementation wrote
		// template=1 through /config using an unresolved VmRef, producing the
		// malformed URL /nodes///<vmid>/config -- so it never once succeeded.
		if err := client.CreateTemplate(ctx, vmr); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert guest %d to template: %v", vmid, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d converted to a template", vmid)), nil
	}
}
