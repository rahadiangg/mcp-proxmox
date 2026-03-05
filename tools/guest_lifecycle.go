package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	px "github.com/Telmate/proxmox-api-go/proxmox"
)

// RegisterLifecycleTools registers guest lifecycle tools (start, stop, reboot, etc.)
func RegisterLifecycleTools(s *server.MCPServer, client *proxmox.Client) {
	// Start guest
	startGuestTool := mcp.NewTool("start_guest",
		mcp.WithDescription("Start a VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to start"),
		),
	)
	s.AddTool(startGuestTool, startGuestHandler(client))

	// Stop guest
	stopGuestTool := mcp.NewTool("stop_guest",
		mcp.WithDescription("Force stop a VM or container (immediate power off)"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to stop"),
		),
	)
	s.AddTool(stopGuestTool, stopGuestHandler(client))

	// Shutdown guest
	shutdownGuestTool := mcp.NewTool("shutdown_guest",
		mcp.WithDescription("Gracefully shutdown a VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to shutdown"),
		),
		mcp.WithBoolean("force",
			mcp.Description("Force shutdown (skip ACPI) - NOTE: may not be respected in old SDK versions"),
		),
	)
	s.AddTool(shutdownGuestTool, shutdownGuestHandler(client))

	// Reboot guest
	rebootGuestTool := mcp.NewTool("reboot_guest",
		mcp.WithDescription("Reboot a VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to reboot"),
		),
	)
	s.AddTool(rebootGuestTool, rebootGuestHandler(client))

	// Pause guest
	pauseGuestTool := mcp.NewTool("pause_guest",
		mcp.WithDescription("Pause a VM (QEMU only) - saves state to memory"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID to pause"),
		),
	)
	s.AddTool(pauseGuestTool, pauseGuestHandler(client))

	// Resume guest
	resumeGuestTool := mcp.NewTool("resume_guest",
		mcp.WithDescription("Resume a paused VM"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID to resume"),
		),
	)
	s.AddTool(resumeGuestTool, resumeGuestHandler(client))

	// Hibernate guest
	hibernateGuestTool := mcp.NewTool("hibernate_guest",
		mcp.WithDescription("Hibernate a VM (QEMU only) - saves state to disk"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID to hibernate"),
		),
	)
	s.AddTool(hibernateGuestTool, hibernateGuestHandler(client))

	// Delete guest
	deleteGuestTool := mcp.NewTool("delete_guest",
		mcp.WithDescription("Delete a VM or container (WARNING: irreversible)"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to delete"),
		),
	)
	s.AddTool(deleteGuestTool, deleteGuestHandler(client))
}

func startGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.StartVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to start guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d started. UPID: %s", vmid, upid)), nil
	}
}

func stopGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.StopVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to stop guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d stopped. UPID: %s", vmid, upid)), nil
	}
}

func shutdownGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		_ = req.GetBool("force", false) // force parameter accepted but not used in this SDK version

		upid, err := client.ShutdownVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to shutdown guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d shutdown initiated. UPID: %s", vmid, upid)), nil
	}
}

func rebootGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.RebootVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reboot guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d reboot initiated. UPID: %s", vmid, upid)), nil
	}
}

func pauseGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.StatusChangeVm(ctx, vmr, nil, "suspend")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to pause guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d paused. UPID: %s", vmid, upid)), nil
	}
}

func resumeGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.ResumeVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to resume guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d resumed. UPID: %s", vmid, upid)), nil
	}
}

func hibernateGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.StatusChangeVm(ctx, vmr, nil, "suspend")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to hibernate guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d hibernated. UPID: %s", vmid, upid)), nil
	}
}

func deleteGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		upid, err := client.DeleteVm(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete guest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d deletion initiated. UPID: %s", vmid, upid)), nil
	}
}
