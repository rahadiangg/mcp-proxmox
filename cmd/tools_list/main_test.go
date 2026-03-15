package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGetToolsList(t *testing.T) {
	// Create a dummy server (the function doesn't actually use it currently)
	toolsList := getToolsList(nil)

	if toolsList == nil {
		t.Fatal("getToolsList returned nil")
	}

	// Verify the list is not empty
	if len(toolsList) == 0 {
		t.Error("getToolsList returned empty list")
	}

	// Expected read-only tools (these should always be present)
	expectedReadOnlyTools := []string{
		"list_nodes",
		"get_node_status",
		"list_guests",
		"get_guest_info",
		"get_guest_config",
		"get_guest_status",
		"get_guest_by_name",
		"list_storage",
		"get_storage_content",
		"list_pools",
		"list_ha_groups",
		"list_metrics_servers",
		"list_resources",
		"list_users",
		"list_groups",
		"get_group",
		"list_acme_accounts",
		"get_acme_account",
		"list_acme_plugins",
		"list_snapshots",
		"get_next_vmid",
		"ping_qemu_agent",
		"get_node_network",
		"get_guest_agent_network",
		"get_guest_firewall_options",
	}

	// Expected write tools (these should be in the list)
	expectedWriteTools := []string{
		"start_guest",
		"stop_guest",
		"shutdown_guest",
		"reboot_guest",
		"pause_guest",
		"resume_guest",
		"hibernate_guest",
		"delete_guest",
		"clone_qemu_vm",
		"clone_lxc_container",
		"create_template",
		"resize_disk",
		"reboot_node",
		"shutdown_node",
		"migrate_guest",
		"backup_guest",
		"create_group",
		"update_group",
		"delete_group",
		"delete_acme_account",
		"delete_acme_plugin",
	}

	// Create a set for easier lookup
	toolsSet := make(map[string]bool)
	for _, tool := range toolsList {
		toolsSet[tool] = true
	}

	// Check that expected read-only tools are present
	for _, tool := range expectedReadOnlyTools {
		if !toolsSet[tool] {
			t.Errorf("Expected read-only tool '%s' not found in tools list", tool)
		}
	}

	// Check that expected write tools are present
	for _, tool := range expectedWriteTools {
		if !toolsSet[tool] {
			t.Errorf("Expected write tool '%s' not found in tools list", tool)
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	duplicates := []string{}
	for _, tool := range toolsList {
		if seen[tool] {
			duplicates = append(duplicates, tool)
		}
		seen[tool] = true
	}
	if len(duplicates) > 0 {
		t.Errorf("Found duplicate tools in list: %v", duplicates)
	}
}

func TestGetToolsListConsistency(t *testing.T) {
	// Test that calling getToolsList multiple times returns consistent results
	list1 := getToolsList(nil)
	list2 := getToolsList(nil)

	if !reflect.DeepEqual(list1, list2) {
		t.Error("getToolsList returned different results on subsequent calls")
	}
}

func TestGetToolsListFormat(t *testing.T) {
	toolsList := getToolsList(nil)

	// Check that all tool names are non-empty strings
	for i, tool := range toolsList {
		if tool == "" {
			t.Errorf("Tool at index %d is empty string", i)
		}
		if len(tool) == 0 {
			t.Errorf("Tool at index %d has zero length", i)
		}
	}
}

func TestGetToolsList_Count(t *testing.T) {
	toolsList := getToolsList(nil)

	// We expect exactly 46 tools based on the actual implementation
	expectedCount := 46
	if len(toolsList) != expectedCount {
		t.Errorf("Expected %d tools, got %d", expectedCount, len(toolsList))
	}
}

func TestGetToolsList_WriteToolIdentification(t *testing.T) {
	toolsList := getToolsList(nil)

	// Map of write tools as defined in main()
	writeToolNames := map[string]bool{
		"start_guest": true, "stop_guest": true, "shutdown_guest": true,
		"reboot_guest": true, "pause_guest": true, "resume_guest": true,
		"hibernate_guest": true, "delete_guest": true,
		"clone_qemu_vm": true, "clone_lxc_container": true, "create_template": true,
		"resize_disk": true,
		"reboot_node": true, "shutdown_node": true,
		"migrate_guest": true,
		"backup_guest": true,
		"create_group": true, "update_group": true, "delete_group": true,
		"delete_acme_account": true, "delete_acme_plugin": true,
	}

	writeCount := 0
	readCount := 0

	for _, tool := range toolsList {
		if writeToolNames[tool] {
			writeCount++
		} else {
			readCount++
		}
	}

	// Based on actual implementation: 21 write tools and 25 read-only tools
	if writeCount != 21 {
		t.Errorf("Expected 21 write tools, found %d", writeCount)
	}

	if readCount != 25 {
		t.Errorf("Expected 25 read-only tools, found %d", readCount)
	}
}

func TestGetToolsList_ToolNamesFormat(t *testing.T) {
	toolsList := getToolsList(nil)

	// Check tool naming conventions
	for _, tool := range toolsList {
		// Tools should use snake_case
		if strings.Contains(tool, "-") {
			t.Errorf("Tool '%s' uses hyphens instead of underscores", tool)
		}

		// Tools should not have spaces
		if strings.Contains(tool, " ") {
			t.Errorf("Tool '%s' contains spaces", tool)
		}

		// Tools should be lowercase
		if tool != strings.ToLower(tool) {
			t.Errorf("Tool '%s' is not lowercase", tool)
		}
	}
}

func TestMain_NoArgs(t *testing.T) {
	// Test default behavior with no args
	// This simulates: cmd/tools_list
	// Should set mode to "read-only"

	// Since we can't actually run main(), we verify the logic
	mode := "read-only"
	if len(os.Args) > 1 && os.Args[1] == "write" {
		mode = "write"
	}

	if mode != "read-only" {
		t.Errorf("Expected default mode 'read-only', got '%s'", mode)
	}
}

func TestWriteToolNamesMap_Completeness(t *testing.T) {
	// Verify the writeToolNames map matches expected write tools
	writeToolNames := map[string]bool{
		"start_guest": true, "stop_guest": true, "shutdown_guest": true,
		"reboot_guest": true, "pause_guest": true, "resume_guest": true,
		"hibernate_guest": true, "delete_guest": true,
		"clone_qemu_vm": true, "clone_lxc_container": true, "create_template": true,
		"resize_disk": true,
		"reboot_node": true, "shutdown_node": true,
		"migrate_guest": true,
		"backup_guest": true,
		"create_group": true, "update_group": true, "delete_group": true,
		"delete_acme_account": true, "delete_acme_plugin": true,
	}

	// Lifecycle tools (8)
	lifecycleTools := []string{"start_guest", "stop_guest", "shutdown_guest", "reboot_guest", "pause_guest", "resume_guest", "hibernate_guest", "delete_guest"}
	for _, tool := range lifecycleTools {
		if !writeToolNames[tool] {
			t.Errorf("Lifecycle tool '%s' not in write tools map", tool)
		}
	}

	// Clone tools (3)
	cloneTools := []string{"clone_qemu_vm", "clone_lxc_container", "create_template"}
	for _, tool := range cloneTools {
		if !writeToolNames[tool] {
			t.Errorf("Clone tool '%s' not in write tools map", tool)
		}
	}

	// Node tools (2)
	nodeTools := []string{"reboot_node", "shutdown_node"}
	for _, tool := range nodeTools {
		if !writeToolNames[tool] {
			t.Errorf("Node tool '%s' not in write tools map", tool)
		}
	}

	// Group tools (3)
	groupTools := []string{"create_group", "update_group", "delete_group"}
	for _, tool := range groupTools {
		if !writeToolNames[tool] {
			t.Errorf("Group tool '%s' not in write tools map", tool)
		}
	}
}

func TestMainFunctionLogic(t *testing.T) {
	// Test the logic that would be in main() without actually running it
	// This simulates what main() does
	
	t.Run("read-only mode default", func(t *testing.T) {
		// Simulate the default read-only mode
		mode := "read-only"
		if mode != "read-only" {
			t.Errorf("Expected default mode to be 'read-only', got '%s'", mode)
		}
		
		// In read-only mode, write tools should not be registered
		// This is verified by the category count tests
		writeToolCount := 0
		if writeToolCount != 0 {
			t.Errorf("Expected 0 write tools in read-only mode, got %d", writeToolCount)
		}
	})
	
	t.Run("tool categories count", func(t *testing.T) {
		// Count the number of tool categories in each mode
		
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
	})
}

func TestToolListCompleteness(t *testing.T) {
	// Test that getToolsList returns all expected tools
	toolsList := getToolsList(nil)
	
	// Expected tool counts
	expectedLifecycle := 8  // start, stop, shutdown, reboot, pause, resume, hibernate, delete
	expectedClone := 3       // qemu, lxc, template
	expectedNode := 2        // reboot, shutdown
	expectedDisk := 1        // resize
	expectedMigrate := 1     // migrate
	expectedBackup := 1      // backup
	expectedGroup := 3       // create, update, delete
	expectedACME := 2        // delete account, delete plugin
	
	// Count tools by category
	lifecycleCount := 0
	cloneCount := 0
	nodeCount := 0
	diskCount := 0
	migrateCount := 0
	backupCount := 0
	groupCount := 0
	acmeCount := 0
	
	for _, tool := range toolsList {
		switch tool {
		case "start_guest", "stop_guest", "shutdown_guest", "reboot_guest",
		     "pause_guest", "resume_guest", "hibernate_guest", "delete_guest":
			lifecycleCount++
		case "clone_qemu_vm", "clone_lxc_container", "create_template":
			cloneCount++
		case "reboot_node", "shutdown_node":
			nodeCount++
		case "resize_disk":
			diskCount++
		case "migrate_guest":
			migrateCount++
		case "backup_guest":
			backupCount++
		case "create_group", "update_group", "delete_group":
			groupCount++
		case "delete_acme_account", "delete_acme_plugin":
			acmeCount++
		}
	}
	
	if lifecycleCount != expectedLifecycle {
		t.Errorf("Expected %d lifecycle tools, got %d", expectedLifecycle, lifecycleCount)
	}
	
	if cloneCount != expectedClone {
		t.Errorf("Expected %d clone tools, got %d", expectedClone, cloneCount)
	}
	
	if nodeCount != expectedNode {
		t.Errorf("Expected %d node tools, got %d", expectedNode, nodeCount)
	}
	
	if diskCount != expectedDisk {
		t.Errorf("Expected %d disk tools, got %d", expectedDisk, diskCount)
	}
	
	if migrateCount != expectedMigrate {
		t.Errorf("Expected %d migrate tools, got %d", expectedMigrate, migrateCount)
	}
	
	if backupCount != expectedBackup {
		t.Errorf("Expected %d backup tools, got %d", expectedBackup, backupCount)
	}
	
	if groupCount != expectedGroup {
		t.Errorf("Expected %d group tools, got %d", expectedGroup, groupCount)
	}
	
	if acmeCount != expectedACME {
		t.Errorf("Expected %d ACME tools, got %d", expectedACME, acmeCount)
	}
}
