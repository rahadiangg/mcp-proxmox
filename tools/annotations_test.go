package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// registerAll builds a server with every tool, mirroring main.go's write mode.
func registerAll(t *testing.T) *server.MCPServer {
	t.Helper()
	client := &proxmox.Client{}
	s := server.NewMCPServer("test", "test")

	RegisterNodeTools(s, client)
	RegisterGuestTools(s, client)
	RegisterStorageTools(s, client)
	RegisterPoolTools(s, client)
	RegisterHATools(s, client)
	RegisterMetricsTools(s, client)
	RegisterUserTools(s, client)
	RegisterGroupTools(s, client)
	RegisterACMETools(s, client)
	RegisterResourceTools(s, client)
	RegisterStorageContentTools(s, client)
	RegisterSnapshotTools(s, client)
	RegisterQemuAgentTools(s, client)
	RegisterNodeNetworkTools(s, client)
	RegisterNetworkTools(s, client)
	RegisterFirewallTools(s, client)
	RegisterCreateTools(s, client)
	RegisterDiskBandwidthTools(s, client)

	RegisterLifecycleTools(s, client)
	RegisterCloneTools(s, client)
	RegisterDiskTools(s, client)
	RegisterDiskBandwidthWriteTools(s, client)
	RegisterMigrateTools(s, client)
	RegisterBackupTools(s, client)
	RegisterGroupWriteTools(s, client)
	RegisterACMEWriteTools(s, client)
	RegisterNodeWriteTools(s, client)

	return s
}

// TestReadOnlyToolsAreAnnotated pins the client-visible hints. Every tool
// previously shipped readOnlyHint:false and destructiveHint:true, telling
// hosts that list_nodes was as dangerous as delete_guest.
func TestReadOnlyToolsAreAnnotated(t *testing.T) {
	readOnly := []string{
		"list_nodes", "get_node_status", "list_guests", "get_guest_info",
		"get_guest_config", "get_guest_status", "get_guest_by_name",
		"list_storage", "get_storage_status", "get_storage_config",
		"list_pools", "list_ha_groups", "list_metrics_servers", "list_users",
		"list_groups", "get_group", "list_acme_accounts", "get_acme_account",
		"list_acme_plugins", "get_disk_bandwidth", "list_resources",
		"list_snapshots", "get_node_network", "get_guest_agent_network",
		"get_guest_firewall_options", "get_storage_content", "get_next_vmid",
		"ping_qemu_agent",
	}

	tools := registerAll(t).ListTools()
	for _, name := range readOnly {
		st, ok := tools[name]
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if st.Tool.Annotations.ReadOnlyHint == nil || !*st.Tool.Annotations.ReadOnlyHint {
			t.Errorf("%s should be annotated read-only", name)
		}
	}
}

func TestDestructiveToolsAreAnnotated(t *testing.T) {
	destructive := []string{
		"delete_guest", "stop_guest", "delete_group",
		"delete_acme_account", "delete_acme_plugin",
		"reboot_node", "shutdown_node", "resize_disk", "migrate_guest",
	}

	tools := registerAll(t).ListTools()
	for _, name := range destructive {
		st, ok := tools[name]
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if st.Tool.Annotations.DestructiveHint == nil || !*st.Tool.Annotations.DestructiveHint {
			t.Errorf("%s should be annotated destructive", name)
		}
		if st.Tool.Annotations.ReadOnlyHint != nil && *st.Tool.Annotations.ReadOnlyHint {
			t.Errorf("%s must not be annotated read-only", name)
		}
	}
}

// TestEveryToolHasADescription guards against a tool shipping with no
// explanation for the model.
func TestEveryToolHasADescription(t *testing.T) {
	for name, st := range registerAll(t).ListTools() {
		if st.Tool.Description == "" {
			t.Errorf("tool %s has no description", name)
		}
	}
}

// TestToolsRequiringArgumentsDeclareThem is the assertion that would have
// caught the nine stub tools: each dereferences a parameter, but none of them
// declared a schema at all, so the model could not target a VM or node.
func TestToolsRequiringArgumentsDeclareThem(t *testing.T) {
	required := map[string][]string{
		"backup_guest":               {"vmid", "storage"},
		"migrate_guest":              {"vmid", "target_node"},
		"resize_disk":                {"vmid", "disk", "size"},
		"list_snapshots":             {"vmid"},
		"ping_qemu_agent":            {"vmid"},
		"get_guest_agent_network":    {"vmid"},
		"get_guest_firewall_options": {"vmid"},
		"get_node_network":           {"node"},
		"get_storage_content":        {"node", "storage"},
		"get_guest_info":             {"vmid"},
		"set_disk_bandwidth":         {"vmid", "disk_id"},
	}

	tools := registerAll(t).ListTools()
	for name, params := range required {
		st, ok := tools[name]
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		for _, p := range params {
			if _, declared := st.Tool.InputSchema.Properties[p]; !declared {
				t.Errorf("%s reads %q but does not declare it in its schema", name, p)
			}
		}
		for _, p := range params {
			found := false
			for _, r := range st.Tool.InputSchema.Required {
				if r == p {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s should mark %q as required", name, p)
			}
		}
	}
}
