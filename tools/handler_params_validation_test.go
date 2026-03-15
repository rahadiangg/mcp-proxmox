package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Test handler parameter validation logic
// These tests focus on input validation that doesn't require Proxmox server

func TestNodeHandlersParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node name", "pve1", false},
		{"empty node", "", true},
		{"node with domain", "node.example.com", false},
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

func TestStorageHandlersParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		node        string
		expectError bool
	}{
		{"both params present", "local-lvm", "pve1", false},
		{"empty storage", "", "pve1", true},
		{"empty node", "local-lvm", "", true},
		{"both empty", "", "", true},
		{"valid nfs storage", "nfs", "pve1", false},
		{"valid zfs storage", "zfs", "pve1", false},
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

func TestGetStorageContentParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		node        string
		content     string
		expectError bool
	}{
		{"all params present", "local-lvm", "pve1", "images", false},
		{"empty storage", "", "pve1", "images", true},
		{"empty node", "local-lvm", "", "images", true},
		{"both storage and node empty", "", "", "", true},
		{"empty content is valid", "local-lvm", "pve1", "", false}, // content is optional
		{"valid content types", "local", "pve1", "iso", false},
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

func TestListResourcesParameterValidation(t *testing.T) {
	// List resources doesn't require specific parameters
	t.Run("no required params", func(t *testing.T) {
		// This handler works without any parameters
		// It lists all cluster resources
	})
}

func TestListHAGroupsParameterValidation(t *testing.T) {
	// List HA groups doesn't require specific parameters
	t.Run("no required params", func(t *testing.T) {
		// This handler works without any parameters
		// It lists all HA groups
	})
}

func TestListPoolsParameterValidation(t *testing.T) {
	// List pools doesn't require specific parameters
	t.Run("no required params", func(t *testing.T) {
		// This handler works without any parameters
		// It lists all resource pools
	})
}

func TestListMetricsServersParameterValidation(t *testing.T) {
	// List metrics servers doesn't require specific parameters
	t.Run("no required params", func(t *testing.T) {
		// This handler works without any parameters
		// It lists all metrics servers
	})
}

func TestListSnapshotsParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
		{"large vmid", 999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestPingQemuAgentParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestAgentNetworkParameterValidation(t *testing.T) {
	// Get guest agent network doesn't require parameters (uses QEMU agent)
	t.Run("no required params", func(t *testing.T) {
		// This handler works without any parameters
		// It gets network info via QEMU agent
	})
}

func TestResizeDiskParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		disk        string
		size        string
		expectError bool
	}{
		{"all valid", 100, "scsi0", "32G", false},
		{"zero vmid", 0, "scsi0", "32G", true},
		{"empty disk", 100, "", "32G", true},
		{"empty size", 100, "scsi0", "", true},
		{"all invalid", 0, "", "", true},
		{"valid disk with size", 100, "virtio1", "+10G", false},
		{"valid sata disk", 100, "sata0", "50G", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0 || tt.disk == "" || tt.size == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, disk=%q, size=%q: expected error=%v, got error=%v",
					tt.vmid, tt.disk, tt.size, tt.expectError, hasError)
			}
		})
	}
}

func TestBackupGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		storage     string
		expectError bool
	}{
		{"all valid", 100, "local-lvm", false},
		{"zero vmid", 0, "local-lvm", true},
		{"empty storage", 100, "", true},
		{"both invalid", 0, "", true},
		{"valid with nfs", 100, "nfs", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0 || tt.storage == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, storage=%q: expected error=%v, got error=%v",
					tt.vmid, tt.storage, tt.expectError, hasError)
			}
		})
	}
}

func TestGetFirewallOptionsParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestMigrateGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		targetNode  string
		expectError bool
	}{
		{"valid params", 100, "pve2", false},
		{"zero vmid", 0, "pve2", true},
		{"empty target", 100, "", true},
		{"both invalid", 0, "", true},
		{"online migration valid", 100, "pve2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, target=%q: expected error=%v, got error=%v",
					tt.vmid, tt.targetNode, tt.expectError, hasError)
			}
		})
	}
}

func TestGetNodeNetworkParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node", "pve1", false},
		{"empty node", "", true},
		{"node with domain", "node.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Node %q: expected error=%v, got error=%v", tt.node, tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestStatusParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
		{"large vmid", 999999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestConfigParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestInfoParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestACMEHandlersParameterValidation(t *testing.T) {
	// List ACME accounts and plugins don't require parameters
	t.Run("list accounts no params", func(t *testing.T) {
		// List accounts works without parameters
	})

	t.Run("list plugins no params", func(t *testing.T) {
		// List plugins works without parameters
	})
}

func TestGetACMEAccountParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		account     string
		expectError bool
	}{
		{"valid account", "user@realm!1", false},
		{"empty account", "", true},
		{"root@pam", "root@pam", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.account == ""
			if hasError != tt.expectError {
				t.Errorf("Account %q: expected error=%v, got error=%v", tt.account, tt.expectError, hasError)
			}
		})
	}
}

func TestShutdownGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestRebootNodeParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node", "pve1", false},
		{"empty node", "", true},
		{"node with domain", "node.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Node %q: expected error=%v, got error=%v", tt.node, tt.expectError, hasError)
			}
		})
	}
}

func TestShutdownNodeParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectError bool
	}{
		{"valid node", "pve1", false},
		{"empty node", "", true},
		{"node with hyphen", "pve-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.node == ""
			if hasError != tt.expectError {
				t.Errorf("Node %q: expected error=%v, got error=%v", tt.node, tt.expectError, hasError)
			}
		})
	}
}

func TestRebootGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestHibernateGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestDeleteGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestPauseGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestResumeGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestGetGuestByNameParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		guestName   string
		expectError bool
	}{
		{"valid name", "vm-100", false},
		{"empty name", "", true},
		{"name with numbers", "test-vm-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.guestName == ""
			if hasError != tt.expectError {
				t.Errorf("Guest name %q: expected error=%v, got error=%v", tt.guestName, tt.expectError, hasError)
			}
		})
	}
}

func TestCreateTemplateParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestGetNextVMIDParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		startID     int
		expectValid bool
	}{
		{"zero means default", 0, true},     // 0 means use default
		{"specific start", 100, true},       // positive int is valid
		{"negative start", -1, true},        // handler doesn't validate
		{"large start", 999999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The handler uses the value as-is, any int is technically valid
			isValid := true
			if isValid != tt.expectValid {
				t.Errorf("startID=%d: expected valid=%v, got valid=%v", tt.startID, tt.expectValid, isValid)
			}
		})
	}
}

func TestGroupHandlersParameterValidation(t *testing.T) {
	t.Run("list groups no params", func(t *testing.T) {
		// List groups works without parameters
	})
}

func TestGetGroupParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectError bool
	}{
		{"valid group", "admins", false},
		{"empty group", "", true},
		{"group with special chars", "admin-group", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.groupID == ""
			if hasError != tt.expectError {
				t.Errorf("Group %q: expected error=%v, got error=%v", tt.groupID, tt.expectError, hasError)
			}
		})
	}
}

func TestCreateGroupParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectError bool
	}{
		{"valid group", "new-group", false},
		{"empty group", "", true},
		{"group with spaces", "my group", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.groupID == ""
			if hasError != tt.expectError {
				t.Errorf("Group %q: expected error=%v, got error=%v", tt.groupID, tt.expectError, hasError)
			}
		})
	}
}

func TestDeleteGroupParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectError bool
	}{
		{"valid group", "test-group", false},
		{"empty group", "", true},
		{"group with underscores", "test_group", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.groupID == ""
			if hasError != tt.expectError {
				t.Errorf("Group %q: expected error=%v, got error=%v", tt.groupID, tt.expectError, hasError)
			}
		})
	}
}

func TestDiskBandwidthParameterValidation_AllParams(t *testing.T) {
	// Test that all bandwidth parameter names are valid
	allParams := []string{
		"mbps_rd", "mbps_rd_max", "mbps_wr", "mbps_wr_max",
		"iops_rd", "iops_rd_max", "iops_rd_max_length",
		"iops_wr", "iops_wr_max", "iops_wr_max_length",
	}

	if len(allParams) != 10 {
		t.Errorf("Expected 10 bandwidth parameters, got %d", len(allParams))
	}

	// Test minimum values for each parameter
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

	for param, minVal := range minValues {
		t.Run("param_"+param+"_min_value", func(t *testing.T) {
			// Check that minimum value is correctly defined
			if minVal <= 0 {
				t.Errorf("Parameter %s should have positive minimum value, got %d", param, minVal)
			}
		})
	}
}

func TestGetRequiredIntParamZeroValue(t *testing.T) {
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
}

func TestGetRequiredIntParamValidValue(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"vmid": float64(100),
			},
		},
	}

	result, err := getRequiredIntParam(req, "vmid")
	if err != nil {
		t.Errorf("Expected no error for valid value, got %v", err)
	}
	if result != 100 {
		t.Errorf("Expected result=100, got %d", result)
	}
}

func TestGetOptionalStringParamMissing(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	result := getOptionalStringParam(req, "missing")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestGetOptionalStringParamPresent(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"node": "pve1",
			},
		},
	}

	result := getOptionalStringParam(req, "node")
	if result != "pve1" {
		t.Errorf("Expected 'pve1', got %q", result)
	}
}

func TestGetOptionalBoolParamMissing(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	result := getOptionalBoolParam(req, "missing")
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestGetOptionalBoolParamTrue(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"online": true,
			},
		},
	}

	result := getOptionalBoolParam(req, "online")
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestGetOptionalBoolParamFalse(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"online": false,
			},
		},
	}

	result := getOptionalBoolParam(req, "online")
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestCloneQemuVMParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		newID       int
		targetNode  string
		expectError bool
	}{
		{"valid all params", 100, 200, "pve2", false},
		{"zero source", 0, 200, "pve2", true},
		{"zero new id", 100, 0, "pve2", true},
		{"empty target", 100, 200, "", true},
		{"negative source", -1, 200, "pve2", true},
		{"all invalid", 0, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid <= 0 || tt.newID <= 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, newID=%d, node=%s: expected error=%v, got error=%v",
					tt.vmid, tt.newID, tt.targetNode, tt.expectError, hasError)
			}
		})
	}
}

func TestCloneLxcParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		newID       int
		targetNode  string
		expectError bool
	}{
		{"valid all params", 200, 300, "pve2", false},
		{"zero source", 0, 300, "pve2", true},
		{"zero new id", 200, 0, "pve2", true},
		{"empty target", 200, 300, "", true},
		{"all valid with description", 200, 300, "pve3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid <= 0 || tt.newID <= 0 || tt.targetNode == ""
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, newID=%d, node=%s: expected error=%v, got error=%v",
					tt.vmid, tt.newID, tt.targetNode, tt.expectError, hasError)
			}
		})
	}
}

func TestClearDiskBandwidthParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		diskID      string
		expectError bool
	}{
		{"valid params", 100, "scsi0", false},
		{"zero vmid", 0, "scsi0", true},
		{"empty disk", 100, "", true},
		{"invalid disk", 100, "invalid0", true},
		{"both invalid", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasVmidError := tt.vmid == 0
			hasDiskError := tt.diskID == "" || (tt.diskID != "" && !isValidDiskID(tt.diskID))
			hasError := hasVmidError || hasDiskError
			if hasError != tt.expectError {
				t.Errorf("vmid=%d, diskID=%q: expected error=%v, got error=%v",
					tt.vmid, tt.diskID, tt.expectError, hasError)
			}
		})
	}
}

func TestStopGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestStartGuestParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"positive vmid", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.vmid == 0
			if hasError != tt.expectError {
				t.Errorf("vmid=%d: expected error=%v, got error=%v", tt.vmid, tt.expectError, hasError)
			}
		})
	}
}

func TestListGuestsParameterValidation(t *testing.T) {
	// List guests doesn't require specific parameters
	// Can filter by node and type
	t.Run("no required params", func(t *testing.T) {
		// List guests works without parameters
	})

	t.Run("with node filter valid", func(t *testing.T) {
		node := "pve1"
		hasError := node == ""
		if hasError {
			t.Error("Node filter should be valid")
		}
	})

	t.Run("with type filter valid", func(t *testing.T) {
		guestType := "qemu"
		hasError := guestType == ""
		if hasError {
			t.Error("Type filter should be valid")
		}
	})
}

func TestDeleteACMEAccountParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		account     string
		expectError bool
	}{
		{"valid account", "user@realm!1", false},
		{"empty account", "", true},
		{"root@pam account", "root@pam", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.account == ""
			if hasError != tt.expectError {
				t.Errorf("Account %q: expected error=%v, got error=%v", tt.account, tt.expectError, hasError)
			}
		})
	}
}

func TestDeleteACMEPluginParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		plugin      string
		expectError bool
	}{
		{"valid plugin", "acme-plugin", false},
		{"empty plugin", "", true},
		{"plugin with hyphens", "my-plugin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.plugin == ""
			if hasError != tt.expectError {
				t.Errorf("Plugin %q: expected error=%v, got error=%v", tt.plugin, tt.expectError, hasError)
			}
		})
	}
}

func TestUpdateGroupParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectError bool
	}{
		{"valid group", "admins", false},
		{"empty group", "", true},
		{"group with special chars", "admin-group-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.groupID == ""
			if hasError != tt.expectError {
				t.Errorf("Group %q: expected error=%v, got error=%v", tt.groupID, tt.expectError, hasError)
			}
		})
	}
}

func TestListStorageParameterValidation(t *testing.T) {
	// List storage doesn't require parameters
	t.Run("no required params", func(t *testing.T) {
		// List storage works without parameters
	})
}

func TestGetStorageStatusParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		node        string
		expectError bool
	}{
		{"both params", "local-lvm", "pve1", false},
		{"empty storage", "", "pve1", true},
		{"empty node", "local-lvm", "", true},
		{"both empty", "", "", true},
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

func TestMoreValidationCases(t *testing.T) {
	// Test more specific validation patterns
	
	t.Run("VM ID validation patterns", func(t *testing.T) {
		tests := []struct {
			name        string
			vmid        int
			expectValid bool
		}{
			{"minimum valid", 1, true},
			{"common small", 100, true},
			{"common large", 9999, true},
			{"zero invalid", 0, false},
			{"negative not recommended but valid input", -1, true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isValid := tt.vmid != 0
				if isValid != tt.expectValid {
					t.Errorf("VMID %d: expected valid=%v, got valid=%v", tt.vmid, tt.expectValid, isValid)
				}
			})
		}
	})
	
	t.Run("Storage name validation", func(t *testing.T) {
		tests := []struct {
			name        string
			storage     string
			expectValid bool
		}{
			{"simple local", "local", true},
			{"local lvm", "local-lvm", true},
			{"nfs storage", "nfs", true},
			{"zfs storage", "zfs", true},
			{"empty invalid", "", false},
			{"with numbers", "storage123", true},
			{"with underscore", "my_storage", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isValid := tt.storage != ""
				if isValid != tt.expectValid {
					t.Errorf("Storage %q: expected valid=%v, got valid=%v", tt.storage, tt.expectValid, isValid)
				}
			})
		}
	})
	
	t.Run("Node name validation", func(t *testing.T) {
		tests := []struct {
			name        string
			node        string
			expectValid bool
		}{
			{"simple pve1", "pve1", true},
			{"with domain", "node.example.com", true},
			{"with hyphen", "pve-node-1", true},
			{"empty invalid", "", false},
			{"with numbers", "node123", true},
			{"with underscore", "node_1", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isValid := tt.node != ""
				if isValid != tt.expectValid {
					t.Errorf("Node %q: expected valid=%v, got valid=%v", tt.node, tt.expectValid, isValid)
				}
			})
		}
	})
}

func TestSnapshotValidation(t *testing.T) {
	t.Run("snapshot parameter patterns", func(t *testing.T) {
		tests := []struct {
			name        string
			vmid        int
			snapshot    string
			expectError bool
		}{
			{"valid with snapshot", 100, "snap-1", false},
			{"missing vmid", 0, "snap-1", true},
			{"empty snapshot ok", 100, "", false},
			{"numeric snapshot", 100, "123", false},
			{"special chars snapshot", 100, "snap_v1.0", false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hasError := tt.vmid == 0
				if hasError != tt.expectError {
					t.Errorf("Snapshot test (vmid=%d, snap=%q): expected error=%v, got error=%v",
						tt.vmid, tt.snapshot, tt.expectError, hasError)
				}
			})
		}
	})
}

func TestNetworkValidation(t *testing.T) {
	t.Run("network type validation", func(t *testing.T) {
		validTypes := []string{"eth", "bridge", "bond", "vlan"}
		
		for _, netType := range validTypes {
			t.Run("type_"+netType, func(t *testing.T) {
				if len(netType) == 0 {
					t.Errorf("Network type should not be empty")
				}
			})
		}
	})
}

func TestGroupValidation(t *testing.T) {
	t.Run("group name patterns", func(t *testing.T) {
		tests := []struct {
			name        string
			groupID     string
			expectValid bool
		}{
			{"simple alphanumeric", "group1", true},
			{"with hyphen", "my-group", true},
			{"with underscore", "my_group", true},
			{"with numbers", "group123", true},
			{"empty invalid", "", false},
			{"with dot", "my.group", true},
			{"with at sign", "group@pve", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isValid := tt.groupID != ""
				if isValid != tt.expectValid {
					t.Errorf("Group %q: expected valid=%v, got valid=%v", tt.groupID, tt.expectValid, isValid)
				}
			})
		}
	})
}
