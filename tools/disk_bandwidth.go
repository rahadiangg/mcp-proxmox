package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	px "github.com/Telmate/proxmox-api-go/proxmox"
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
	)
	s.AddTool(clearDiskBandwidthTool, clearDiskBandwidthHandler(client))
}

// BandwidthInfo represents bandwidth settings for a disk
type BandwidthInfo struct {
	Storage      string                    `json:"storage,omitempty"`
	VolumePath   string                    `json:"volume_path,omitempty"`
	Bandwidth    map[string]interface{}    `json:"bandwidth,omitempty"`
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

		// Build bandwidth parameters map
		bandwidthParams := make(map[string]interface{})
		minValues := map[string]int{
			"mbps_rd":              1,
			"mbps_rd_max":          1,
			"mbps_wr":              1,
			"mbps_wr_max":          1,
			"iops_rd":              10,
			"iops_rd_max":          10,
			"iops_rd_max_length":   1,
			"iops_wr":              10,
			"iops_wr_max":          10,
			"iops_wr_max_length":   1,
		}

		paramNames := []string{
			"mbps_rd", "mbps_rd_max", "mbps_wr", "mbps_wr_max",
			"iops_rd", "iops_rd_max", "iops_rd_max_length",
			"iops_wr", "iops_wr_max", "iops_wr_max_length",
		}

		for _, param := range paramNames {
			val := req.GetFloat(param, -1)
			if val >= 0 {
				minVal := minValues[param]
				if val != 0 && float64(val) < float64(minVal) {
					return mcp.NewToolResultError(fmt.Sprintf("invalid value for %s: %.0f. Minimum value is %d (0 means unlimited)", param, val, minVal)), nil
				}
				bandwidthParams[param] = int(val)
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

		// Build new disk config string with bandwidth parameters
		newDiskConfig := buildDiskConfigString(diskConfig, bandwidthParams)

		// Update VM config
		updateConfig := map[string]interface{}{
			diskID: newDiskConfig,
		}

		upid, err := client.SetVmConfig(vmr, updateConfig)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set disk bandwidth: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Disk bandwidth settings updated for VM %d, disk '%s'. UPID: %s", vmid, diskID, upid)), nil
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

		// Update VM config
		updateConfig := map[string]interface{}{
			diskID: newDiskConfig,
		}

		upid, err := client.SetVmConfig(vmr, updateConfig)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to clear disk bandwidth: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Disk bandwidth limits cleared for VM %d, disk '%s'. UPID: %s", vmid, diskID, upid)), nil
	}
}

// isValidDiskID checks if the disk ID matches a valid format
func isValidDiskID(diskID string) bool {
	// Valid disk ID prefixes: scsi, virtio, sata, ide, ata followed by a number
	pattern := `^(scsi|virtio|sata|ide|ata)\d+$`
	matched, _ := regexp.MatchString(pattern, diskID)
	return matched
}

// isDiskKey checks if a config key is a disk key
func isDiskKey(key string) bool {
	pattern := `^(scsi|virtio|sata|ide|ata)\d+$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
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

// parseDiskBandwidth extracts bandwidth parameters from disk config string
func parseDiskBandwidth(diskConfig string) map[string]interface{} {
	bandwidthParams := []string{
		"mbps_rd", "mbps_rd_max", "mbps_wr", "mbps_wr_max",
		"iops_rd", "iops_rd_max", "iops_rd_max_length",
		"iops_wr", "iops_wr_max", "iops_wr_max_length",
	}

	bandwidth := make(map[string]interface{})

	parts := strings.Split(diskConfig, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		for _, param := range bandwidthParams {
			if strings.HasPrefix(part, param+"=") {
				valueStr := strings.TrimPrefix(part, param+"=")
				if val, err := strconv.Atoi(valueStr); err == nil {
					bandwidth[param] = val
				}
			}
		}
	}

	return bandwidth
}

// buildDiskConfigString creates a new disk config string with updated bandwidth parameters
func buildDiskConfigString(originalConfig string, bandwidthParams map[string]interface{}) string {
	parts := strings.Split(originalConfig, ",")

	// Extract base storage config (first part)
	baseConfig := parts[0]
	otherParams := make(map[string]string)

	// Collect non-bandwidth parameters
	bandwidthKeys := map[string]bool{
		"mbps_rd": true, "mbps_rd_max": true, "mbps_wr": true, "mbps_wr_max": true,
		"iops_rd": true, "iops_rd_max": true, "iops_rd_max_length": true,
		"iops_wr": true, "iops_wr_max": true, "iops_wr_max_length": true,
	}

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		eqIndex := strings.Index(part, "=")
		if eqIndex == -1 {
			// No value, skip
			continue
		}

		key := part[:eqIndex]
		value := part[eqIndex+1:]

		if !bandwidthKeys[key] {
			otherParams[key] = value
		}
	}

	// Build new config string
	var resultParts []string
	resultParts = append(resultParts, baseConfig)

	// Add other non-bandwidth parameters first
	for key, value := range otherParams {
		resultParts = append(resultParts, fmt.Sprintf("%s=%s", key, value))
	}

	// Add new bandwidth parameters
	for key, value := range bandwidthParams {
		resultParts = append(resultParts, fmt.Sprintf("%s=%v", key, value))
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
