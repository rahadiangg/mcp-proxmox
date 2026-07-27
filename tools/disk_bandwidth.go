package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	px "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// RegisterDiskBandwidthTools registers read-only disk bandwidth tools (always available)
func RegisterDiskBandwidthTools(s *server.MCPServer, client *proxmox.Client) {
	getDiskBandwidthTool := mcp.NewTool("get_disk_bandwidth",
		mcp.WithDescription("Get current bandwidth settings for a VM's disks"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithString("disk_id",
			mcp.Description("Specific disk ID like 'scsi0' - if omitted, returns all disks"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
	s.AddTool(getDiskBandwidthTool, getDiskBandwidthHandler(client))
}

// RegisterDiskBandwidthWriteTools registers write-only disk bandwidth tools (requires write mode)
func RegisterDiskBandwidthWriteTools(s *server.MCPServer, client *proxmox.Client) {
	// Set disk bandwidth
	setDiskBandwidthTool := mcp.NewTool("set_disk_bandwidth",
		mcp.WithDescription("Set bandwidth limits for a VM's disk"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithString("disk_id",
			mcp.Required(),
			mcp.Description("Disk identifier (e.g., 'scsi0', 'virtio0', 'sata0')"),
		),
		mcp.WithNumber("mbps_rd",
			mcp.Description("Read limit MB/s (min 1, 0 = unlimited)"),
		),
		mcp.WithNumber("mbps_rd_max",
			mcp.Description("Read burst MB/s (min 1, 0 = unlimited)"),
		),
		mcp.WithNumber("mbps_wr",
			mcp.Description("Write limit MB/s (min 1, 0 = unlimited)"),
		),
		mcp.WithNumber("mbps_wr_max",
			mcp.Description("Write burst MB/s (min 1, 0 = unlimited)"),
		),
		mcp.WithNumber("iops_rd",
			mcp.Description("Read IOPS limit (min 10, 0 = unlimited)"),
		),
		mcp.WithNumber("iops_rd_max",
			mcp.Description("Read IOPS burst (min 10, 0 = unlimited)"),
		),
		mcp.WithNumber("iops_rd_max_length",
			mcp.Description("Read burst duration (seconds)"),
		),
		mcp.WithNumber("iops_wr",
			mcp.Description("Write IOPS limit (min 10, 0 = unlimited)"),
		),
		mcp.WithNumber("iops_wr_max",
			mcp.Description("Write IOPS burst (min 10, 0 = unlimited)"),
		),
		mcp.WithNumber("iops_wr_max_length",
			mcp.Description("Write burst duration (seconds)"),
		),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(setDiskBandwidthTool, setDiskBandwidthHandler(client))

	// Clear disk bandwidth
	clearDiskBandwidthTool := mcp.NewTool("clear_disk_bandwidth",
		mcp.WithDescription("Remove all bandwidth limits from a disk"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithString("disk_id",
			mcp.Required(),
			mcp.Description("Disk identifier (e.g., 'scsi0', 'virtio0')"),
		),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(clearDiskBandwidthTool, clearDiskBandwidthHandler(client))
}

// BandwidthInfo represents bandwidth settings for a disk
type BandwidthInfo struct {
	Storage    string                 `json:"storage,omitempty"`
	VolumePath string                 `json:"volume_path,omitempty"`
	Bandwidth  map[string]interface{} `json:"bandwidth,omitempty"`
}

// getDiskBandwidthHandler handles get_disk_bandwidth requests
func getDiskBandwidthHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		diskID := req.GetString("disk_id", "")

		// Validate disk ID format if provided
		if diskID != "" && !isValidDiskID(diskID) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid disk_id format: '%s'. Expected format: disk type prefix (scsi, virtio, sata, ide, ata) followed by a number (e.g., 'scsi0', 'virtio1', 'sata0')", diskID)), nil
		}

		vmr := px.NewVmRef(px.GuestID(vmid))
		config, err := client.GetVmConfig(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get VM config: %v", err)), nil
		}

		result := make(map[string]interface{})

		// If disk_id specified, return only that disk
		if diskID != "" {
			diskConfig, ok := config[diskID].(string)
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("disk '%s' not found on VM %d", diskID, vmid)), nil
			}

			storage, volumePath := parseStorageFromConfig(diskConfig)
			bandwidth := parseDiskBandwidth(diskConfig)

			result[diskID] = BandwidthInfo{
				Storage:    storage,
				VolumePath: volumePath,
				Bandwidth:  bandwidth,
			}
		} else {
			// Return all disks
			for key, value := range config {
				if isDiskKey(key) {
					if diskConfig, ok := value.(string); ok {
						storage, volumePath := parseStorageFromConfig(diskConfig)
						bandwidth := parseDiskBandwidth(diskConfig)

						result[key] = BandwidthInfo{
							Storage:    storage,
							VolumePath: volumePath,
							Bandwidth:  bandwidth,
						}
					}
				}
			}
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

// setDiskBandwidthHandler handles set_disk_bandwidth requests
func setDiskBandwidthHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		diskID := req.GetString("disk_id", "")
		if diskID == "" {
			return mcp.NewToolResultError("disk_id is required"), nil
		}

		if !isValidDiskID(diskID) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid disk_id format: '%s'. Expected format: disk type prefix (scsi, virtio, sata, ide, ata) followed by a number (e.g., 'scsi0', 'virtio1', 'sata0')", diskID)), nil
		}

		// Presence is detected by looking the key up, not by a -1 sentinel.
		// The sentinel could not tell "omitted" from "explicitly negative", so
		// a negative value was silently dropped while the tool reported success.
		args := req.GetArguments()
		bandwidthParams := make(map[string]interface{})

		for _, param := range bandwidthParamOrder {
			raw, present := args[param]
			if !present || raw == nil {
				continue // omitted: leave whatever is already on the disk
			}

			val, ok := toFloat(raw)
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("invalid value for %s: expected a number, got %T", param, raw)), nil
			}

			if val < 0 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid value for %s: %g. Must be >= 0 (0 means unlimited)", param, val)), nil
			}

			if val == 0 {
				// 0 is documented as unlimited, and unlimited is how Proxmox
				// represents an absent key. Remove it rather than writing 0.
				bandwidthParams[param] = nil
				continue
			}

			if minVal := bandwidthMinValues[param]; val < minVal {
				return mcp.NewToolResultError(fmt.Sprintf("invalid value for %s: %g. Minimum value is %g (0 means unlimited)", param, val, minVal)), nil
			}

			if isIntegerParam(param) {
				if val != math.Trunc(val) {
					return mcp.NewToolResultError(fmt.Sprintf("invalid value for %s: %g. Must be a whole number", param, val)), nil
				}
				bandwidthParams[param] = int(val)
				continue
			}

			// mbps_* accept fractional values; truncating changed what the
			// caller asked for without telling them.
			if val == math.Trunc(val) {
				bandwidthParams[param] = int(val)
			} else {
				bandwidthParams[param] = val
			}
		}

		if len(bandwidthParams) == 0 {
			return mcp.NewToolResultError("at least one bandwidth parameter is required"), nil
		}

		// Get current VM config to find the disk
		vmr := px.NewVmRef(px.GuestID(vmid))
		config, err := client.GetVmConfig(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get VM config: %v", err)), nil
		}

		diskConfig, ok := config[diskID].(string)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("disk '%s' not found on VM %d", diskID, vmid)), nil
		}

		// Merge the requested limits over what the disk already carries.
		newDiskConfig := buildDiskConfigString(diskConfig, bandwidthParams)

		status, err := writeDiskConfig(ctx, client, vmr, diskID, newDiskConfig, config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set disk bandwidth: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Disk bandwidth settings updated for VM %d, disk '%s'. Task status: %s", vmid, diskID, status)), nil
	}
}

// clearDiskBandwidthHandler handles clear_disk_bandwidth requests
func clearDiskBandwidthHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid := int(req.GetFloat("vmid", 0))
		if vmid == 0 {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		diskID := req.GetString("disk_id", "")
		if diskID == "" {
			return mcp.NewToolResultError("disk_id is required"), nil
		}

		if !isValidDiskID(diskID) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid disk_id format: '%s'. Expected format: disk type prefix (scsi, virtio, sata, ide, ata) followed by a number (e.g., 'scsi0', 'virtio1', 'sata0')", diskID)), nil
		}

		// Get current VM config
		vmr := px.NewVmRef(px.GuestID(vmid))
		config, err := client.GetVmConfig(ctx, vmr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get VM config: %v", err)), nil
		}

		diskConfig, ok := config[diskID].(string)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("disk '%s' not found on VM %d", diskID, vmid)), nil
		}

		// Build new disk config without bandwidth parameters
		newDiskConfig := removeBandwidthParams(diskConfig)

		status, err := writeDiskConfig(ctx, client, vmr, diskID, newDiskConfig, config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to clear disk bandwidth: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Disk bandwidth limits cleared for VM %d, disk '%s'. Task status: %s", vmid, diskID, status)), nil
	}
}

// diskIDPattern matches a Proxmox disk identifier. Compiled once: the previous
// per-call regexp.MatchString also discarded its compile error.
var diskIDPattern = regexp.MustCompile(`^(scsi|virtio|sata|ide|ata)\d+$`)

// isValidDiskID checks if the disk ID matches a valid format
func isValidDiskID(diskID string) bool {
	return diskIDPattern.MatchString(diskID)
}

// isDiskKey checks if a config key is a disk key
func isDiskKey(key string) bool {
	return diskIDPattern.MatchString(key)
}

// bandwidthParamOrder is the canonical emission order for throttle keys. A
// fixed order keeps the rewritten disk line byte-stable; ranging over a map
// previously reshuffled it on every write, churning the VM config.
var bandwidthParamOrder = []string{
	"mbps_rd", "mbps_rd_max", "mbps_wr", "mbps_wr_max",
	"iops_rd", "iops_rd_max", "iops_rd_max_length",
	"iops_wr", "iops_wr_max", "iops_wr_max_length",
}

// bandwidthKeys is the set form of bandwidthParamOrder.
var bandwidthKeys = func() map[string]bool {
	m := make(map[string]bool, len(bandwidthParamOrder))
	for _, k := range bandwidthParamOrder {
		m[k] = true
	}
	return m
}()

// bandwidthMinValues is the single source of truth for per-parameter minimums.
// mbps_* are floats in the Proxmox API; iops_* are whole numbers.
var bandwidthMinValues = map[string]float64{
	"mbps_rd":            1,
	"mbps_rd_max":        1,
	"mbps_wr":            1,
	"mbps_wr_max":        1,
	"iops_rd":            10,
	"iops_rd_max":        10,
	"iops_rd_max_length": 1,
	"iops_wr":            10,
	"iops_wr_max":        10,
	"iops_wr_max_length": 1,
}

// isIntegerParam reports whether a throttle parameter must be a whole number.
func isIntegerParam(param string) bool {
	return strings.HasPrefix(param, "iops_")
}

// toFloat coerces a JSON-decoded argument to a float64. MCP arguments arrive
// as float64, but a client may send an int or a numeric string.
func toFloat(raw interface{}) (float64, bool) {
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// writeDiskConfig persists a rewritten disk line.
//
// It uses PostWithTask rather than the deprecated SetVmConfig, which hardcodes
// context.Background() -- so a cancelled or timed-out tool call previously kept
// mutating the VM. It also forwards the config digest so a concurrent edit is
// rejected by Proxmox instead of silently clobbered.
func writeDiskConfig(
	ctx context.Context,
	client *proxmox.Client,
	vmr *px.VmRef,
	diskID, diskConfig string,
	currentConfig map[string]interface{},
) (string, error) {
	params := map[string]interface{}{diskID: diskConfig}
	if digest, ok := currentConfig["digest"].(string); ok && digest != "" {
		params["digest"] = digest
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.Node(), vmr.GetVmType(), int(vmr.VmId()))
	return client.PostWithTask(ctx, params, url)
}

// parseStorageFromConfig extracts storage and volume path from disk config string
func parseStorageFromConfig(diskConfig string) (storage, volumePath string) {
	parts := strings.Split(diskConfig, ",")
	if len(parts) == 0 {
		return "", ""
	}

	baseConfig := parts[0]
	colonIndex := strings.Index(baseConfig, ":")
	if colonIndex == -1 {
		// No volume path, just storage
		return baseConfig, ""
	}

	storage = baseConfig[:colonIndex]
	volumePath = baseConfig[colonIndex+1:]
	return storage, volumePath
}

// parseDiskBandwidth extracts bandwidth parameters from disk config string.
//
// mbps_* are floats in the Proxmox API. Parsing them with Atoi silently
// dropped any fractional limit, so a disk throttled at mbps_rd=10.5 was
// reported as having no throttle at all. Integral values are still returned
// as int so callers and tests see 100 rather than 100.0.
func parseDiskBandwidth(diskConfig string) map[string]interface{} {
	bandwidth := make(map[string]interface{})

	parts := strings.Split(diskConfig, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		for _, param := range bandwidthParamOrder {
			if !strings.HasPrefix(part, param+"=") {
				continue
			}
			valueStr := strings.TrimPrefix(part, param+"=")
			val, err := strconv.ParseFloat(valueStr, 64)
			if err != nil || math.IsInf(val, 0) || math.IsNaN(val) {
				continue
			}
			if val == math.Trunc(val) {
				bandwidth[param] = int(val)
			} else {
				bandwidth[param] = val
			}
		}
	}

	return bandwidth
}

// buildDiskConfigString rewrites a disk config line, MERGING the supplied
// throttle values over whatever the line already carries.
//
// A key absent from bandwidthParams keeps its existing value; a key mapped to
// nil is removed. This is the fix for a silent data-loss bug: the function
// used to drop every pre-existing throttle it was not explicitly given, so
// setting mbps_rd alone wiped mbps_wr and every iops_* limit on that disk.
//
// Output order is fixed -- base volume, then non-throttle options in their
// original order, then throttles in canonical order -- so a no-op write
// reproduces the input byte for byte.
func buildDiskConfigString(originalConfig string, bandwidthParams map[string]interface{}) string {
	parts := strings.Split(originalConfig, ",")

	baseConfig := strings.TrimSpace(parts[0])
	var otherParts []string
	existing := make(map[string]string)

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		eqIndex := strings.Index(part, "=")
		if eqIndex == -1 {
			// A bare flag such as "ssd". Preserved -- dropping these meant
			// set_disk_bandwidth silently discarded options that
			// clear_disk_bandwidth kept.
			otherParts = append(otherParts, part)
			continue
		}

		key := part[:eqIndex]
		if bandwidthKeys[key] {
			existing[key] = part[eqIndex+1:]
			continue
		}
		otherParts = append(otherParts, part)
	}

	resultParts := append([]string{baseConfig}, otherParts...)

	for _, key := range bandwidthParamOrder {
		if value, supplied := bandwidthParams[key]; supplied {
			if value == nil {
				continue // explicit removal
			}
			resultParts = append(resultParts, fmt.Sprintf("%s=%v", key, value))
			continue
		}
		if prev, ok := existing[key]; ok {
			resultParts = append(resultParts, fmt.Sprintf("%s=%s", key, prev))
		}
	}

	return strings.Join(resultParts, ",")
}

// removeBandwidthParams creates a new disk config string without bandwidth parameters
func removeBandwidthParams(originalConfig string) string {
	parts := strings.Split(originalConfig, ",")

	// Extract base storage config (first part)
	baseConfig := parts[0]

	bandwidthKeys := map[string]bool{
		"mbps_rd": true, "mbps_rd_max": true, "mbps_wr": true, "mbps_wr_max": true,
		"iops_rd": true, "iops_rd_max": true, "iops_rd_max_length": true,
		"iops_wr": true, "iops_wr_max": true, "iops_wr_max_length": true,
	}

	var resultParts []string
	resultParts = append(resultParts, baseConfig)

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		eqIndex := strings.Index(part, "=")
		if eqIndex == -1 {
			resultParts = append(resultParts, part)
			continue
		}

		key := part[:eqIndex]

		if !bandwidthKeys[key] {
			resultParts = append(resultParts, part)
		}
	}

	return strings.Join(resultParts, ",")
}
