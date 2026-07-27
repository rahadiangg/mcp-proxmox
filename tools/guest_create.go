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

const (
	guestIDMinimum = 100
	guestIDMaximum = 999999999
	// maxVmidProbes bounds the scan when the caller supplies a start_id. The
	// SDK's own helper increments without any bound and only stops on a
	// non-API error, so a high start_id makes it hammer the cluster forever.
	maxVmidProbes = 64
)

func RegisterCreateTools(s *server.MCPServer, client *proxmox.Client) {
	getNextVmidTool := mcp.NewTool("get_next_vmid",
		mcp.WithDescription("Get the next available VM ID. Without start_id this asks the cluster directly; with start_id it probes upward from that ID."),
		mcp.WithNumber("start_id",
			mcp.Description("Search upward from this ID instead of asking the cluster for its next free ID"),
			mcp.Min(guestIDMinimum),
			mcp.Max(guestIDMaximum),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getNextVmidTool, getNextVmidHandler(client))
}

func getNextVmidHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Absent start_id: one request to /cluster/nextid, no scanning.
		raw, hasStart := req.GetArguments()["start_id"]
		if !hasStart || raw == nil {
			nextID, err := client.GetNextID(ctx, nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get next VMID: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{"vmid": int(nextID)}, "next VMID")
		}

		startID := int(req.GetFloat("start_id", guestIDMinimum))
		if startID < guestIDMinimum || startID > guestIDMaximum {
			return mcp.NewToolResultError(fmt.Sprintf(
				"invalid start_id %d: must be between %d and %d", startID, guestIDMinimum, guestIDMaximum)), nil
		}

		// Bounded probe. Each call asks the cluster whether a specific ID is
		// free; PVE errors when it is taken.
		for offset := 0; offset < maxVmidProbes; offset++ {
			candidate := startID + offset
			if candidate > guestIDMaximum {
				break
			}

			free, err := vmidIsFree(ctx, client, candidate)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to check VMID %d: %v", candidate, err)), nil
			}
			if free {
				return jsonResult(map[string]interface{}{
					"vmid":     candidate,
					"start_id": startID,
					"probes":   offset + 1,
				}, "next VMID")
			}
		}

		return mcp.NewToolResultError(fmt.Sprintf(
			"no free VMID found in [%d, %d) after %d probes. Omit start_id to use the cluster's next free ID",
			startID, startID+maxVmidProbes, maxVmidProbes)), nil
	}
}

// vmidIsFree reports whether a specific guest ID is available. A Proxmox API
// error means the ID is taken; any other failure is a real error.
func vmidIsFree(ctx context.Context, client *proxmox.Client, vmid int) (bool, error) {
	url := "/cluster/nextid?vmid=" + strconv.Itoa(vmid)
	envelope, err := client.GetItemList(ctx, url)
	if err != nil {
		if _, isAPIErr := err.(*px.ApiError); isAPIErr {
			return false, nil
		}
		return false, err
	}

	// PVE echoes the id back when it is free, as a string or a number
	// depending on version, so accept both rather than asserting one.
	switch v := unwrapData(envelope).(type) {
	case string:
		return v != "", nil
	case float64:
		return true, nil
	default:
		return false, nil
	}
}
