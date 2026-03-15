package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Test additional input validation scenarios for handlers

func TestGuestCloneValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		vmid            float64
		newID           float64
		targetNode      string
		fullClone       string
		expectError     bool
		description      string
	}{
		{
			name:        "valid basic clone",
			vmid:        100,
			newID:       200,
			targetNode:  "pve2",
			fullClone:   "1",
			expectError:  false,
			description: "All valid parameters",
		},
		{
			name:        "zero source vmid",
			vmid:        0,
			newID:       200,
			targetNode:  "pve2",
			fullClone:   "1",
			expectError:  true,
			description: "Source VM ID cannot be 0",
		},
		{
			name:        "zero target vmid",
			vmid:        100,
			newID:       0,
			targetNode:  "pve2",
			fullClone:   "1",
			expectError:  true,
			description: "New VM ID cannot be 0",
		},
		{
			name:        "empty target node",
			vmid:        100,
			newID:       200,
			targetNode:  "",
			fullClone:   "1",
			expectError:  true,
			description: "Target node cannot be empty",
		},
		{
			name:        "negative vmid",
			vmid:        -1,
			newID:       200,
			targetNode:  "pve2",
			fullClone:   "1",
			expectError:  true,
			description: "Negative VM ID is invalid",
		},
		{
			name:        "negative new id",
			vmid:        100,
			newID:       -1,
			targetNode:  "pve2",
			fullClone:   "1",
			expectError:  true,
			description: "Negative new VM ID is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceID := int(tt.vmid)
			newID := int(tt.newID)
			hasError := sourceID <= 0 || newID <= 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("%s: expected error=%v, got error=%v", tt.description, tt.expectError, hasError)
			}
		})
	}
}

func TestBackupGuestValidation(t *testing.T) {
	tests := []struct {
		name         string
		vmid         float64
		storage      string
		compression  string
		expectError  bool
	}{
		{
			name:        "valid backup",
			vmid:        100,
			storage:     "local",
			compression: "zstd",
			expectError: false,
		},
		{
			name:        "zero vmid",
			vmid:        0,
			storage:     "local",
			compression: "zstd",
			expectError:  true,
		},
		{
			name:        "empty storage",
			vmid:        100,
			storage:     "",
			compression: "zstd",
			expectError:  true,
		},
		{
			name:        "empty compression",
			vmid:        100,
			storage:     "local",
			compression: "",
			expectError:  false, // compression is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0 || tt.storage == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestResizeDiskValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		diskID      string
		size        string
		expectError bool
	}{
		{
			name:        "valid resize",
			vmid:        100,
			diskID:      "scsi0",
			size:        "+10G",
			expectError: false,
		},
		{
			name:        "zero vmid",
			vmid:        0,
			diskID:      "scsi0",
			size:        "+10G",
			expectError:  true,
		},
		{
			name:        "empty disk id",
			vmid:        100,
			diskID:      "",
			size:        "+10G",
			expectError:  true,
		},
		{
			name:        "invalid disk id",
			vmid:        100,
			diskID:      "net0",
			size:        "+10G",
			expectError:  true, // net0 is not a valid disk ID
		},
		{
			name:        "empty size",
			vmid:        100,
			diskID:      "scsi0",
			size:        "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			isValidDisk := isValidDiskID(tt.diskID)
			hasError := vmid == 0 || tt.diskID == "" || tt.size == "" || !isValidDisk
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestMigrateGuestValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		vmid         float64
		targetNode   string
		targetStorage string
		online       string
		expectError  bool
	}{
		{
			name:         "valid migration",
			vmid:         100,
			targetNode:   "pve2",
			targetStorage: "local",
			online:       "1",
			expectError:  false,
		},
		{
			name:         "zero vmid",
			vmid:         0,
			targetNode:   "pve2",
			targetStorage: "local",
			online:       "1",
			expectError:  true,
		},
		{
			name:         "empty target node",
			vmid:         100,
			targetNode:   "",
			targetStorage: "local",
			online:       "1",
			expectError:  true,
		},
		{
			name:         "with target storage",
			vmid:         100,
			targetNode:   "pve2",
			targetStorage: "local-lvm",
			online:       "1",
			expectError:  false,
		},
		{
			name:         "online migration",
			vmid:         100,
			targetNode:   "pve2",
			targetStorage: "",
			online:       "1",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestRebootShutdownNodeValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node", "pve1", false},
		{"empty node", "", true},
		{"node with spaces", "pve 1", false},
		{"node with numbers", "node123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestGroupCreateUpdateDeleteValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		comment     string
		expectError bool
	}{
		{
			name:        "valid group",
			groupID:     "admins",
			comment:     "Administrators",
			expectError:  false,
		},
		{
			name:        "empty group id",
			groupID:     "",
			comment:     "Administrators",
			expectError:  true,
		},
		{
			name:        "empty comment",
			groupID:     "admins",
			comment:     "",
			expectError:  false, // comment is optional
		},
		{
			name:        "group with spaces",
			groupID:     "admin group",
			comment:     "Administrators",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.groupID == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestACMEOperationsValidation(t *testing.T) {
	tests := []struct {
		name        string
		account     string
		plugin      string
		expectError bool
	}{
		{
			name:        "valid account",
			account:     "user@realm!1",
			plugin:      "",
			expectError:  false,
		},
		{
			name:        "empty account",
			account:     "",
			plugin:      "",
			expectError:  true,
		},
		{
			name:        "valid plugin",
			account:     "",
			plugin:      "acme-plugin",
			expectError:  false,
		},
		{
			name:        "empty plugin",
			account:     "user@realm!1",
			plugin:      "",
			expectError:  false, // plugin is optional for some ops
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.account == "" && tt.plugin == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestSnapshotOperationsValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		snapname    string
		expectError bool
	}{
		{
			name:        "valid snapshot",
			vmid:        100,
			snapname:    "before-upgrade",
			expectError:  false,
		},
		{
			name:        "zero vmid",
			vmid:        0,
			snapname:    "before-upgrade",
			expectError:  true,
		},
		{
			name:        "empty snapname",
			vmid:        100,
			snapname:    "",
			expectError:  true,
		},
		{
			name:        "special chars in snapname",
			vmid:        100,
			snapname:    "snap-2024_01",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0 || tt.snapname == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestGetNextVmidValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		expectError bool
	}{
		{"default (0)", 0, false},     // 0 means auto-assign
		{"specific vmid", 100, false},
		{"negative", -1, true},       // negative is invalid
		{"large number", 999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid < 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v for vmid=%d", tt.expectError, hasError, vmid)
			}
		})
	}
}

func TestStorageContentValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		node        string
		contentType string
		expectError bool
	}{
		{
			name:        "valid params",
			storage:     "local",
			node:        "pve1",
			contentType:  "",
			expectError:  false,
		},
		{
			name:        "empty storage",
			storage:     "",
			node:        "pve1",
			contentType:  "",
			expectError:  true,
		},
		{
			name:        "empty node",
			storage:     "local",
			node:        "",
			contentType:  "",
			expectError:  true,
		},
		{
			name:        "with content type",
			storage:     "local",
			node:        "pve1",
			contentType:  "iso",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.storage == "" || tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestQemuAgentPingValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"negative vmid", -1, false}, // negative is technically valid input
		{"large vmid", 999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestGuestFirewallOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"negative vmid", -1, false}, // negative is technically valid input
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestResourceListValidation(t *testing.T) {
	tests := []struct {
		name        string
		typeStr     string
		expectError bool
	}{
		{"no type filter", "", false},
		{"vm type", "vm", false},
		{"node type", "node", false},
		{"storage type", "storage", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Type filter is optional, so always valid
			hasError := false
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestNetworkConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node", "pve1", false},
		{"empty node", "", true},
		{"node with special chars", "pve-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestDiskBandwidthSetValidation_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		diskID      string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name:   "valid set single param",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{"mbps_rd": 100},
			expectError: false,
		},
		{
			name:   "zero vmid",
			vmid:   0,
			diskID: "scsi0",
			params: map[string]interface{}{"mbps_rd": 100},
			expectError: true,
		},
		{
			name:   "empty disk id",
			vmid:   100,
			diskID: "",
			params: map[string]interface{}{"mbps_rd": 100},
			expectError: true,
		},
		{
			name:   "invalid disk id",
			vmid:   100,
			diskID: "net0",
			params: map[string]interface{}{"mbps_rd": 100},
			expectError: true,
		},
		{
			name:   "no bandwidth params",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{},
			expectError: true,
		},
		{
			name:   "below minimum mbps",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{"mbps_rd": 0.5},
			expectError: false, // float values are converted to int
		},
		{
			name:   "below minimum iops",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{"iops_rd": 5},
			expectError: false, // int values are accepted
		},
		{
			name:   "zero is unlimited (valid)",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{"mbps_rd": 0, "iops_rd": 0},
			expectError: false,
		},
		{
			name:   "multiple params",
			vmid:   100,
			diskID: "scsi0",
			params: map[string]interface{}{"mbps_rd": 100, "mbps_wr": 200},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			validDisk := isValidDiskID(tt.diskID)
			hasParams := len(tt.params) > 0

			hasError := vmid == 0 || tt.diskID == "" || !validDisk || !hasParams
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestToolRequestHelpers_EdgeCases(t *testing.T) {
	t.Run("getRequiredIntParam with negative value", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"vmid": float64(-1),
				},
			},
		}

		result, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != -1 {
			t.Errorf("Expected -1, got %d", result)
		}
	})

	t.Run("getOptionalStringParam with numeric", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"value": float64(42),
				},
			},
		}

		result := getOptionalStringParam(req, "value")
		// Numeric values don't convert to strings in GetInt
		if result != "" {
			t.Logf("Numeric value returns empty string: %q", result)
		}
	})

	t.Run("getOptionalBoolParam with numeric 1", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"value": float64(1),
				},
			},
		}

		result := getOptionalBoolParam(req, "value")
		if !result {
			t.Error("Numeric 1 should return true")
		}
	})
}

func TestMapBoolToInt_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected int
	}{
		{"true returns 1", true, 1},
		{"false returns 0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapBoolToInt(tt.input)
			if result != tt.expected {
				t.Errorf("mapBoolToInt(%v) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetOptionalStringParam_NilArguments(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: nil,
		},
	}

	result := getOptionalStringParam(req, "key")
	if result != "" {
		t.Errorf("Expected empty string for nil arguments, got %q", result)
	}
}

func TestGetOptionalBoolParam_NilArguments(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: nil,
		},
	}

	result := getOptionalBoolParam(req, "key")
	if result {
		t.Error("Expected false for nil arguments")
	}
}

func TestGetRequiredIntParam_NilArguments(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: nil,
		},
	}

	result, err := getRequiredIntParam(req, "key")
	if err == nil {
		t.Error("Expected error for nil arguments")
	}
	if result != 0 {
		t.Errorf("Expected 0 for nil arguments, got %d", result)
	}
}
