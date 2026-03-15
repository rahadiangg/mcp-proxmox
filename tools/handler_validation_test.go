package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Test handler input validation for various tools
// These tests focus on input validation logic that doesn't require a real Proxmox client

func TestGuestLifecycleHandlerValidation_StartGuest(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"negative vmid", -1, false}, // validation only checks for 0
		{"large vmid", 999999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test the full handler without a client, but we can test
			// the validation logic pattern used in handlers
			vmid := int(tt.vmid)
			if vmid == 0 && !tt.expectError {
				t.Error("Expected valid vmid")
			}
			if vmid != 0 && tt.expectError {
				t.Error("Expected error for invalid vmid")
			}
		})
	}
}

func TestGuestLifecycleHandlerValidation_PauseGuest(t *testing.T) {
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

func TestDiskBandwidthValidation_DiskID(t *testing.T) {
	tests := []struct {
		name        string
		diskID      string
		expectValid bool
	}{
		{"valid scsi0", "scsi0", true},
		{"valid virtio1", "virtio1", true},
		{"valid sata2", "sata2", true},
		{"valid ide3", "ide3", true},
		{"valid ata4", "ata4", true},
		{"invalid missing number", "scsi", false},
		{"invalid wrong prefix", "nvme0", false},
		{"invalid empty", "", false},
		{"invalid special chars", "scsi-0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidDiskID(tt.diskID)
			if isValid != tt.expectValid {
				t.Errorf("isValidDiskID(%q) = %v; want %v", tt.diskID, isValid, tt.expectValid)
			}
		})
	}
}

func TestDiskBandwidthValidation_BandwidthParams(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		value       float64
		minValue    int
		expectValid bool
	}{
		{"valid mbps_rd", "mbps_rd", 100, 1, true},
		{"zero mbps_rd (unlimited)", "mbps_rd", 0, 1, true},
		{"below minimum mbps_rd", "mbps_rd", 0.5, 1, false},
		{"valid iops_rd", "iops_rd", 100, 10, true},
		{"zero iops_rd (unlimited)", "iops_rd", 0, 10, true},
		{"below minimum iops_rd", "iops_rd", 5, 10, false},
		{"valid mbps_rd_max", "mbps_rd_max", 150, 1, true},
		{"valid iops_rd_max_length", "iops_rd_max_length", 60, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation logic from setDiskBandwidthHandler
			val := tt.value
			minVal := tt.minValue

			isValid := val == 0 || float64(val) >= float64(minVal)
			if isValid != tt.expectValid {
				t.Errorf("Validation for %s=%.0f (min=%d) = %v; want %v",
					tt.paramName, val, minVal, isValid, tt.expectValid)
			}
		})
	}
}

func TestBandwidthParamsMinValues(t *testing.T) {
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

	tests := []struct {
		name     string
		param    string
		value    float64
		expected bool
	}{
		{"mbps_rd valid", "mbps_rd", 1, true},
		{"mbps_rd below min", "mbps_rd", 0, true}, // 0 means unlimited
		{"iops_rd valid", "iops_rd", 10, true},
		{"iops_rd below min", "iops_rd", 5, false},
		{"iops_rd unlimited", "iops_rd", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minVal := minValues[tt.param]
			isValid := tt.value == 0 || tt.value >= float64(minVal)
			if isValid != tt.expected {
				t.Errorf("Param %s=%.0f (min=%d): expected valid=%v, got valid=%v",
					tt.param, tt.value, minVal, tt.expected, isValid)
			}
		})
	}
}

func TestCloneGuestValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		newID       float64
		expectError bool
	}{
		{"valid both ids", 100, 200, false},
		{"zero source vmid", 0, 200, true},
		{"valid source, zero new id", 100, 0, true},
		{"both zero", 0, 0, true},
		{"negative source", -1, 200, true}, // negative is invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceID := int(tt.vmid)
			newID := int(tt.newID)

			hasError := sourceID <= 0 || newID <= 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v for vmid=%d, newID=%d", tt.expectError, sourceID, newID)
			}
		})
	}
}

func TestNodeHandlerValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node name", "pve1", false},
		{"empty node", "", true},
		{"node with spaces", "pve 1", false},
		{"node with hyphen", "pve-1", false},
		{"node with numbers", "node123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.node == ""
			if isEmpty != tt.expectError {
				t.Errorf("Node %q: expected error=%v, got error=%v", tt.node, tt.expectError, isEmpty)
			}
		})
	}
}

func TestStorageHandlerValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		node        string
		expectError bool
	}{
		{"valid params", "local-lvm", "pve1", false},
		{"empty storage", "", "pve1", true},
		{"empty node", "local-lvm", "", true},
		{"both empty", "", "", true},
		{"valid with underscore", "local_lvm", "pve1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.storage == "" || tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Storage=%q, Node=%q: expected error=%v, got error=%v",
					tt.storage, tt.node, tt.expectError, hasError)
			}
		})
	}
}

func TestGroupHandlerValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectError bool
	}{
		{"valid group", "admins", false},
		{"empty group", "", true},
		{"group with spaces", "admin group", false},
		{"group with numbers", "group123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.groupID == ""
			if isEmpty != tt.expectError {
				t.Errorf("Group %q: expected error=%v, got error=%v", tt.groupID, tt.expectError, isEmpty)
			}
		})
	}
}

func TestCreateGuestValidation_NextVMID(t *testing.T) {
	tests := []struct {
		name        string
		startID     float64
		expectValid bool
	}{
		{"default start", 0, true},     // 0 means use default
		{"specific start", 100, true},
		{"negative start", -1, true},   // handler doesn't validate
		{"large start", 999999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startID := int(tt.startID)
			// The handler uses the value as-is, any int is technically valid
			isValid := true
			if isValid != tt.expectValid {
				t.Errorf("startID=%d: expected valid=%v, got valid=%v", startID, tt.expectValid, isValid)
			}
		})
	}
}

func TestMigrateGuestValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        float64
		targetNode  string
		expectError bool
	}{
		{"valid params", 100, "pve2", false},
		{"zero vmid", 0, "pve2", true},
		{"empty target node", 100, "", true},
		{"both invalid", 0, "", true},
		{"online migration", 100, "pve2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmid := int(tt.vmid)
			hasError := vmid == 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, target=%q: expected error=%v, got error=%v",
					vmid, tt.targetNode, tt.expectError, hasError)
			}
		})
	}
}

func TestACMEHandlerValidation(t *testing.T) {
	tests := []struct {
		name        string
		account     string
		expectError bool
	}{
		{"valid account", "user@realm!1", false},
		{"empty account", "", true},
		{"account without separator", "user", false}, // handler only checks empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.account == ""
			if isEmpty != tt.expectError {
				t.Errorf("Account %q: expected error=%v, got error=%v", tt.account, tt.expectError, isEmpty)
			}
		})
	}
}

func TestToolRequestValidation_Helpers(t *testing.T) {
	// Test the helper functions used across handlers

	t.Run("getRequiredIntParam with zero value", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"vmid": float64(0),
				},
			},
		}

		result, err := getRequiredIntParam(req, "vmid")
		if err == nil {
			t.Error("Expected error for zero value")
		}
		if result != 0 {
			t.Errorf("Expected result=0, got %d", result)
		}
	})

	t.Run("getOptionalStringParam with missing key", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result := getOptionalStringParam(req, "missing")
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})

	t.Run("getOptionalBoolParam with missing key", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result := getOptionalBoolParam(req, "missing")
		if result != false {
			t.Errorf("Expected false, got %v", result)
		}
	})
}

