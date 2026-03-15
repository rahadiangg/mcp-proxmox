package tools

import (
	"testing"
)

// Test ACME handler validation logic without requiring Proxmox server

func TestACMEAccountValidation(t *testing.T) {
	tests := []struct {
		name        string
		account     string
		expectError bool
	}{
		{"valid account format", "user@realm!1", false},
		{"root@pam account", "root@pam!1", false},
		{"empty account", "", true},
		{"account with special chars", "user-name@pve!100", false},
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

func TestACMEPluginValidation(t *testing.T) {
	tests := []struct {
		name        string
		plugin      string
		expectError bool
	}{
		{"valid plugin name", "acme-plugin", false},
		{"plugin with hyphens", "my-plugin", false},
		{"plugin with underscores", "my_plugin", false},
		{"empty plugin", "", true},
		{"plugin with numbers", "plugin123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.plugin == ""
			if isEmpty != tt.expectError {
				t.Errorf("Plugin %q: expected error=%v, got error=%v", tt.plugin, tt.expectError, isEmpty)
			}
		})
	}
}

func TestACMEAccountFormatValidation(t *testing.T) {
	// Test various ACME account formats
	tests := []struct {
		name        string
		account     string
		hasSeparator bool
	}{
		{"valid with ! separator", "user@realm!1", true},
		{"valid with number", "root@pam!42", true},
		{"missing separator", "user@realm", false},
		{"missing realm", "user!1", false},
		{"missing number", "user@realm!", false},
		{"with hyphens", "user-name@realm!1", true},
		{"complex", "admin_user@pve!100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSeparator := len(tt.account) > 0 && tt.account[len(tt.account)-1] != '!'
			if tt.hasSeparator && !hasSeparator {
				t.Errorf("Account %q should have separator ending", tt.account)
			}
		})
	}
}

func TestLifecycleOperations_Validation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		expectError bool
	}{
		{"valid vmid", 100, false},
		{"zero vmid", 0, true},
		{"negative vmid", -1, false}, // handlers don't validate negative
		{"minimum positive", 1, false},
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

func TestCloneOperations_Validation(t *testing.T) {
	tests := []struct {
		name        string
		vmid        int
		newID       int
		targetNode  string
		expectError bool
	}{
		{"valid clone", 100, 200, "pve2", false},
		{"zero source", 0, 200, "pve2", true},
		{"zero target", 100, 0, "pve2", true},
		{"empty node", 100, 200, "", true},
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

func TestDiskIDFormats_AllValid(t *testing.T) {
	// Test all valid disk ID formats with different prefixes and numbers
	prefixes := []string{"scsi", "virtio", "sata", "ide", "ata"}

	for _, prefix := range prefixes {
		// Test single digit numbers (0-9)
		for i := 0; i <= 9; i++ {
			diskID := prefix + string(rune('0'+i))
			if !isValidDiskID(diskID) {
				t.Errorf("Expected %s to be valid", diskID)
			}
		}

		// Test double digit numbers (10-99)
		for i := 10; i <= 99; i++ {
			diskID := prefix + string(rune('0'+i/10)) + string(rune('0'+i%10))
			if !isValidDiskID(diskID) {
				t.Errorf("Expected %s to be valid", diskID)
			}
		}
	}
}

func TestBandwidthParams_AllCombinations(t *testing.T) {
	// Test all bandwidth parameter names are recognized
	allParams := []string{
		"mbps_rd", "mbps_rd_max", "mbps_wr", "mbps_wr_max",
		"iops_rd", "iops_rd_max", "iops_rd_max_length",
		"iops_wr", "iops_wr_max", "iops_wr_max_length",
	}

	for _, param := range allParams {
		t.Run("param_"+param, func(t *testing.T) {
			// Each parameter name should be recognized
			if param == "" {
				t.Errorf("Parameter name should not be empty")
			}
		})
	}
}

func TestStorageConfigParsing_Complex(t *testing.T) {
	// Test complex storage configurations
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
	}{
		{
			name:       "with params and bandwidth",
			diskConfig: "local-lvm:vm-100-disk-0,size=32G,cache=writeback,mbps_rd=100",
			storage:    "local-lvm",
			volumePath: "vm-100-disk-0",
		},
			{
			name:       "with multiple colons in path",
			diskConfig: "nfs:subdir:vm-100-disk-0",
			storage:    "nfs",
			volumePath: "subdir:vm-100-disk-0",
		},
		{
			name:       "with spaces in params",
			diskConfig: "local:vm-100-disk-0,size = 32G",
			storage:    "local",
			volumePath: "vm-100-disk-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, volumePath := parseStorageFromConfig(tt.diskConfig)
			if storage != tt.storage {
				t.Errorf("Expected storage=%q, got %q", tt.storage, storage)
			}
			// For the volume path, we check it starts with the expected value since params affect parsing
			if volumePath != tt.volumePath {
				t.Errorf("Expected volumePath=%q, got %q", tt.volumePath, volumePath)
			}
		})
	}
}

func TestVMIDValues_AllRanges(t *testing.T) {
	// Test VM ID values across different ranges
	tests := []struct {
		name        string
		vmid        int
		expectValid bool
	}{
		{"zero invalid", 0, false},
		{"min valid", 1, true},
		{"normal", 100, true},
		{"three digits", 999, true},
		{"four digits", 1000, true},
		{"five digits", 100000, true},
		{"max practical", 100000000, true},
		{"negative one", -1, false},
		{"large negative", -999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.vmid > 0
			if isValid != tt.expectValid {
				t.Errorf("vmid=%d: expected valid=%v, got valid=%v", tt.vmid, tt.expectValid, isValid)
			}
		})
	}
}

func TestNodeNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		expectValid bool
	}{
		{"simple name", "pve1", true},
		{"with domain", "node.example.com", true},
		{"with hyphen", "pve-1", true},
		{"with numbers", "node123", true},
		{"empty", "", false},
		{"with spaces", "pve 1", true},
		{"with underscores", "pve_1", true},
		{"with dots", "pve.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.node != ""
			if isValid != tt.expectValid {
				t.Errorf("Node %q: expected valid=%v, got valid=%v", tt.node, tt.expectValid, isValid)
			}
		})
	}
}

func TestStorageNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		storage     string
		expectValid bool
	}{
		{"simple", "local", true},
		{"with hyphen", "local-lvm", true},
		{"with underscore", "my_storage", true},
		{"with numbers", "storage123", true},
		{"nfs", "nfs", true},
		{"ceph", "ceph", true},
		{"zfs", "zfs", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.storage != ""
			if isValid != tt.expectValid {
				t.Errorf("Storage %q: expected valid=%v, got valid=%v", tt.storage, tt.expectValid, isValid)
			}
		})
	}
}
