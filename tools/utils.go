package tools

import (
	"context"
	"fmt"
	"regexp"

	px "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// pveNamePattern matches Proxmox node, storage and pool names. Anchored so a
// value cannot smuggle path separators or a query string into an API URL.
var pveNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,62}$`)

// isValidPVEName reports whether s is safe to interpolate into an API path.
// Rejects empty strings, "..", and anything containing / ? # or whitespace.
func isValidPVEName(s string) bool {
	return pveNamePattern.MatchString(s)
}

// getRequiredIntParam extracts a required integer parameter from the tool request.
// A zero value is treated as missing, which suits identifiers such as vmid.
func getRequiredIntParam(req mcp.CallToolRequest, key string) (int, error) {
	val := req.GetFloat(key, 0)
	if val == 0 {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	return int(val), nil
}

// getRequiredNameParam extracts and validates a required Proxmox name such as
// a node or storage identifier.
func getRequiredNameParam(req mcp.CallToolRequest, key string) (string, error) {
	val := req.GetString(key, "")
	if val == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	if !isValidPVEName(val) {
		return "", fmt.Errorf("invalid %s: %q. Expected letters, digits, dots, dashes or underscores", key, val)
	}
	return val, nil
}

// getOptionalStringParam extracts an optional string parameter from the tool request
func getOptionalStringParam(req mcp.CallToolRequest, key string) string {
	return req.GetString(key, "")
}

// getOptionalBoolParam extracts an optional boolean parameter from the tool request
func getOptionalBoolParam(req mcp.CallToolRequest, key string) bool {
	return req.GetBool(key, false)
}

// mapBoolToInt converts a boolean to int (1 for true, 0 for false)
func mapBoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// resolveGuest builds a VmRef and resolves its node and guest type, which the
// SDK needs before it can construct any per-guest URL. Returns an error if the
// guest does not exist.
func resolveGuest(ctx context.Context, client *proxmox.Client, vmid int) (*px.VmRef, error) {
	vmr := px.NewVmRef(px.GuestID(vmid))
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return nil, fmt.Errorf("guest %d not found: %w", vmid, err)
	}
	return vmr, nil
}

// requireQemu reports an error when a QEMU-only operation targets a container.
func requireQemu(vmr *px.VmRef, operation string) error {
	if vmr.GetVmType() != px.GuestQemu {
		return fmt.Errorf("%s is only supported for QEMU VMs; guest %d is a container", operation, int(vmr.VmId()))
	}
	return nil
}

// unwrapData pulls the "data" payload out of a raw Proxmox response envelope.
// Returns nil when the key is absent, which callers render as an empty result.
func unwrapData(envelope map[string]interface{}) interface{} {
	if envelope == nil {
		return nil
	}
	return envelope["data"]
}
