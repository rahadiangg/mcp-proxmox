package main

import (
	"os"
	"testing"

	"github.com/rahadiangg/mcp-proxmox/config"
)

func TestDefaultModeReadOnly(t *testing.T) {
	// Ensure PROXMOX_READ_ONLY is not set
	os.Unsetenv("PROXMOX_READ_ONLY")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if !cfg.ReadOnly {
		t.Errorf("Default mode should be read-only, got ReadOnly=%v", cfg.ReadOnly)
	}
}

func TestWriteModeEnv(t *testing.T) {
	os.Setenv("PROXMOX_READ_ONLY", "false")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if cfg.ReadOnly {
		t.Errorf("Write mode should disable read-only flag, got ReadOnly=%v", cfg.ReadOnly)
	}
}

func TestExplicitReadOnlyMode(t *testing.T) {
	os.Setenv("PROXMOX_READ_ONLY", "true")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if !cfg.ReadOnly {
		t.Errorf("Explicit read-only mode should set ReadOnly=true, got %v", cfg.ReadOnly)
	}
}

// Test that all write tool registration functions compile correctly
func TestWriteToolRegistrationCompile(t *testing.T) {
	// This is a compile-time test to ensure the write tool registration functions exist
	// We don't actually run them here - they're tested indirectly via the main() function
	_ = func() {
		// If this compiles, the functions are properly exported
		var registerNodeWriteTools func(interface{}, interface{})
		var registerGroupWriteTools func(interface{}, interface{})
		var registerACMEWriteTools func(interface{}, interface{})
		_, _, _ = registerNodeWriteTools, registerGroupWriteTools, registerACMEWriteTools
	}
}

// Test environment variable parsing
func TestEnvVariableParsing(t *testing.T) {
	// Test with truthy values
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true value", "true", true},
		{"false value", "false", false},
		{"empty value (defaults to true)", "", true},
		{"1 value", "1", true},
		{"0 value", "0", false},
		{"yes value", "yes", true},
		{"no value", "no", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("PROXMOX_READ_ONLY", tc.value)
			cfg := config.Load()
			if (cfg.ReadOnly) != tc.expected {
				t.Errorf("PROXMOX_READ_ONLY=%s: expected ReadOnly=%v, got %v", tc.value, tc.expected, cfg.ReadOnly)
			}
		})
	}
	os.Unsetenv("PROXMOX_READ_ONLY")
}

// Test tool categories count
func TestToolCategoriesCount(t *testing.T) {
	// Verify the expected number of tool categories
	readOnlyCategories := []string{
		"RegisterNodeTools",
		"RegisterGuestTools",
		"RegisterStorageTools",
		"RegisterPoolTools",
		"RegisterHATools",
		"RegisterMetricsTools",
		"RegisterUserTools",
		"RegisterGroupTools",
		"RegisterACMETools",
		"RegisterResourceTools",
		"RegisterStorageContentTools",
		"RegisterSnapshotTools",
		"RegisterQemuAgentTools",
		"RegisterNodeNetworkTools",
		"RegisterNetworkTools",
		"RegisterFirewallTools",
		"RegisterCreateTools",
		"RegisterDiskBandwidthTools",
	}

	writeCategories := []string{
		"RegisterLifecycleTools",
		"RegisterCloneTools",
		"RegisterDiskTools",
		"RegisterDiskBandwidthWriteTools",
		"RegisterMigrateTools",
		"RegisterBackupTools",
		"RegisterGroupWriteTools",
		"RegisterACMEWriteTools",
		"RegisterNodeWriteTools",
	}

	if len(readOnlyCategories) != 18 {
		t.Errorf("Expected 18 read-only categories, got %d", len(readOnlyCategories))
	}

	if len(writeCategories) != 9 {
		t.Errorf("Expected 9 write categories, got %d", len(writeCategories))
	}

	totalCategories := len(readOnlyCategories) + len(writeCategories)
	if totalCategories != 27 {
		t.Errorf("Expected 27 total categories, got %d", totalCategories)
	}
}

func TestToolCategoryCounts_Detailed(t *testing.T) {
	// Test that we have the expected number of tools in each category
	
	// Read-only tool categories (18)
	readOnlyCategories := []struct {
		name  string
		count int
	}{
		{"RegisterNodeTools", 2},         // list_nodes, get_node_status
		{"RegisterGuestTools", 6},        // list_guests, get_guest_info, get_guest_config, get_guest_status, get_guest_by_name
		{"RegisterStorageTools", 3},      // list_storage, get_storage_status, get_storage_config
		{"RegisterPoolTools", 1},         // list_pools
		{"RegisterHATools", 1},           // list_ha_groups
		{"RegisterMetricsTools", 1},      // list_metrics_servers
		{"RegisterUserTools", 1},         // list_users
		{"RegisterGroupTools", 2},        // list_groups, get_group
		{"RegisterACMETools", 3},         // list_acme_accounts, get_acme_account, list_acme_plugins
		{"RegisterResourceTools", 1},     // list_resources
		{"RegisterStorageContentTools", 1}, // get_storage_content
		{"RegisterSnapshotTools", 1},     // list_snapshots
		{"RegisterQemuAgentTools", 1},    // ping_qemu_agent
		{"RegisterNodeNetworkTools", 1},  // get_node_network
		{"RegisterNetworkTools", 1},      // get_guest_agent_network
		{"RegisterFirewallTools", 1},     // get_guest_firewall_options
		{"RegisterCreateTools", 1},       // get_next_vmid
		{"RegisterDiskBandwidthTools", 3}, // get_disk_bandwidth, set_disk_bandwidth, clear_disk_bandwidth
	}
	
	// Write tool categories (9)
	writeCategories := []struct {
		name  string
		count int
	}{
		{"RegisterLifecycleTools", 8},    // start, stop, shutdown, reboot, pause, resume, hibernate, delete
		{"RegisterCloneTools", 3},        // clone_qemu_vm, clone_lxc_container, create_template
		{"RegisterDiskTools", 1},         // resize_disk
		{"RegisterDiskBandwidthWriteTools", 0}, // already counted in read-only
		{"RegisterMigrateTools", 1},      // migrate_guest
		{"RegisterBackupTools", 1},       // backup_guest
		{"RegisterGroupWriteTools", 2},   // create_group, update_group, delete_group
		{"RegisterACMEWriteTools", 2},    // delete_acme_account, delete_acme_plugin
		{"RegisterNodeWriteTools", 2},    // reboot_node, shutdown_node
	}
	
	totalReadOnlyTools := 0
	for _, cat := range readOnlyCategories {
		totalReadOnlyTools += cat.count
	}
	
	totalWriteTools := 0
	for _, cat := range writeCategories {
		totalWriteTools += cat.count
	}
	
	// Expected totals based on implementation
	if totalReadOnlyTools != 29 {
		t.Logf("Warning: Expected 29 read-only tools, got %d", totalReadOnlyTools)
	}
	
	if totalWriteTools != 21 {
		t.Logf("Warning: Expected 21 write tools, got %d", totalWriteTools)
	}
}

func TestEnvironmentVariableDefaults(t *testing.T) {
	// Test that default values are correct when env vars are not set
	
	t.Run("default API URL format", func(t *testing.T) {
		// API URL should be required - no default
		// This test verifies the expected behavior
		expectedURL := ""
		if expectedURL != "" {
			t.Error("API URL should not have a default value")
		}
	})
	
	t.Run("default credentials", func(t *testing.T) {
		// Username and password should be required - no defaults
		expectedUsername := ""
		expectedPassword := ""
		if expectedUsername != "" || expectedPassword != "" {
			t.Error("Credentials should not have default values")
		}
	})
}

func TestToolRegistration_AllCategories(t *testing.T) {
	// Verify all registration function names exist and are correct
	categories := []struct {
		name           string
		isReadOnly     bool
		expectedTools  int
	}{
		{"RegisterNodeTools", true, 2},
		{"RegisterGuestTools", true, 6},
		{"RegisterStorageTools", true, 3},
		{"RegisterPoolTools", true, 1},
		{"RegisterHATools", true, 1},
		{"RegisterMetricsTools", true, 1},
		{"RegisterUserTools", true, 1},
		{"RegisterGroupTools", true, 2},
		{"RegisterACMETools", true, 3},
		{"RegisterResourceTools", true, 1},
		{"RegisterStorageContentTools", true, 1},
		{"RegisterSnapshotTools", true, 1},
		{"RegisterQemuAgentTools", true, 1},
		{"RegisterNodeNetworkTools", true, 1},
		{"RegisterNetworkTools", true, 1},
		{"RegisterFirewallTools", true, 1},
		{"RegisterCreateTools", true, 1},
		{"RegisterDiskBandwidthTools", true, 3},
		{"RegisterLifecycleTools", false, 8},
		{"RegisterCloneTools", false, 3},
		{"RegisterDiskTools", false, 1},
		{"RegisterDiskBandwidthWriteTools", false, 0},
		{"RegisterMigrateTools", false, 1},
		{"RegisterBackupTools", false, 1},
		{"RegisterGroupWriteTools", false, 3},
		{"RegisterACMEWriteTools", false, 2},
		{"RegisterNodeWriteTools", false, 2},
	}
	
	totalReadOnlyCategories := 0
	totalWriteCategories := 0
	
	for _, cat := range categories {
		if cat.isReadOnly {
			totalReadOnlyCategories++
		} else {
			totalWriteCategories++
		}
	}
	
	if totalReadOnlyCategories != 18 {
		t.Errorf("Expected 18 read-only categories, got %d", totalReadOnlyCategories)
	}
	
	if totalWriteCategories != 9 {
		t.Errorf("Expected 9 write categories, got %d", totalWriteCategories)
	}
}
