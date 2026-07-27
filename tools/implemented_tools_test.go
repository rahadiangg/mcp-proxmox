package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

// qemuGuest is the cluster-resource fixture CheckVmRef consumes to resolve a
// VmRef's node and guest type.
func qemuGuest(vmid float64, node string) map[string]interface{} {
	return map[string]interface{}{"vmid": vmid, "node": node, "type": "qemu", "name": "vm-test"}
}

func lxcGuest(vmid float64, node string) map[string]interface{} {
	return map[string]interface{}{"vmid": vmid, "node": node, "type": "lxc", "name": "ct-test"}
}

// resourcesRoute is the route CheckVmRef hits while resolving a guest.
func resourcesRoute(guests ...map[string]interface{}) (string, interface{}) {
	list := make([]interface{}, 0, len(guests))
	for _, g := range guests {
		list = append(list, g)
	}
	return "GET /cluster/resources", list
}

func withGuests(routes map[string]interface{}, guests ...map[string]interface{}) map[string]interface{} {
	k, v := resourcesRoute(guests...)
	routes[k] = v
	return routes
}

// --- The nine formerly-stubbed tools --------------------------------------
//
// Each previously returned a canned success string without contacting
// Proxmox. Every test below asserts a real request was made.

func TestListResources_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /cluster/resources": []interface{}{
			map[string]interface{}{"id": "qemu/100", "type": "qemu", "name": "web-1"},
		},
	})

	res, err := listResourcesHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "web-1") {
		t.Errorf("expected real resource data, got: %s", out)
	}
	if !log.called("GET", "/cluster/resources") {
		t.Error("list_resources did not contact Proxmox")
	}
}

func TestListSnapshots_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/snapshot": []interface{}{
			map[string]interface{}{"name": "before-upgrade", "snaptime": float64(1700000000)},
			map[string]interface{}{"name": "current"},
		},
	}, qemuGuest(100, "pve1")))

	res, err := listSnapshotsHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "before-upgrade") {
		t.Errorf("expected snapshot data, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/qemu/100/snapshot") {
		t.Errorf("list_snapshots did not query snapshots; calls: %v", log.calls())
	}
}

func TestListSnapshots_WorksForContainers(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/lxc/200/snapshot": []interface{}{
			map[string]interface{}{"name": "ct-snap"},
		},
	}, lxcGuest(200, "pve1")))

	res, err := listSnapshotsHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(200)}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "ct-snap") {
		t.Errorf("expected container snapshots, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/lxc/200/snapshot") {
		t.Errorf("expected the lxc snapshot path; calls: %v", log.calls())
	}
}

func TestPingQemuAgent_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"POST /nodes/pve1/qemu/100/agent/ping": nil,
	}, qemuGuest(100, "pve1")))

	res, err := pingQemuAgentHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)
	if !log.called("POST", "/nodes/pve1/qemu/100/agent/ping") {
		t.Errorf("ping_qemu_agent did not ping; calls: %v", log.calls())
	}
}

func TestPingQemuAgent_RejectsContainer(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, lxcGuest(200, "pve1")))
	res, _ := pingQemuAgentHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(200)}))
	assertToolError(t, res, "only supported for QEMU")
}

func TestGetGuestAgentNetwork_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/agent/network-get-interfaces": map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{"name": "eth0", "hardware-address": "aa:bb:cc:dd:ee:ff"},
			},
		},
	}, qemuGuest(100, "pve1")))

	res, err := getGuestAgentNetworkHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "eth0") {
		t.Errorf("expected interface data, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/qemu/100/agent/network-get-interfaces") {
		t.Errorf("did not query agent interfaces; calls: %v", log.calls())
	}
}

func TestGetGuestFirewallOptions_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/firewall/options": map[string]interface{}{
			"enable": float64(1), "policy_in": "DROP",
		},
	}, qemuGuest(100, "pve1")))

	res, err := getGuestFirewallOptionsHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "policy_in") {
		t.Errorf("expected firewall options, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/qemu/100/firewall/options") {
		t.Errorf("did not query firewall options; calls: %v", log.calls())
	}
}

func TestGetNodeNetwork_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /nodes/pve1/network": []interface{}{
			map[string]interface{}{"iface": "vmbr0", "type": "bridge"},
		},
	})

	res, err := getNodeNetworkHandler(client)(context.Background(), newRequest(map[string]interface{}{"node": "pve1"}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if out := assertToolSuccess(t, res); !strings.Contains(out, "vmbr0") {
		t.Errorf("expected network config, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/network") {
		t.Errorf("did not query node network; calls: %v", log.calls())
	}
}

func TestResizeDisk_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"PUT /nodes/pve1/qemu/100/resize": "UPID:pve1:00001234:00ABCDEF:65000000:qmresize:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := resizeDiskHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "disk": "scsi0", "size": "+10G",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)
	if !log.called("PUT", "/nodes/pve1/qemu/100/resize") {
		t.Errorf("resize_disk did not call the resize endpoint; calls: %v", log.calls())
	}
}

func TestResizeDisk_ValidatesArguments(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
	h := resizeDiskHandler(client)

	cases := []struct {
		name string
		args map[string]interface{}
		want string
	}{
		{"missing vmid", map[string]interface{}{"disk": "scsi0", "size": "10G"}, "vmid is required"},
		{"missing disk", map[string]interface{}{"vmid": float64(100), "size": "10G"}, "disk is required"},
		{"missing size", map[string]interface{}{"vmid": float64(100), "disk": "scsi0"}, "size is required"},
		{"bad disk id", map[string]interface{}{"vmid": float64(100), "disk": "notadisk", "size": "10G"}, "invalid disk format"},
		{"bad size", map[string]interface{}{"vmid": float64(100), "disk": "scsi0", "size": "big"}, "invalid size"},
		{"shrink attempt", map[string]interface{}{"vmid": float64(100), "disk": "scsi0", "size": "-10G"}, "invalid size"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := h(context.Background(), newRequest(tc.args))
			assertToolError(t, res, tc.want)
		})
	}
}

func TestMigrateGuest_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"POST /nodes/pve1/qemu/100/migrate": "UPID:pve1:00001234:00ABCDEF:65000000:qmigrate:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := migrateGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "target_node": "pve2", "online": true,
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)
	if !log.called("POST", "/nodes/pve1/qemu/100/migrate") {
		t.Errorf("migrate_guest did not call migrate; calls: %v", log.calls())
	}
	if body, ok := log.bodyFor("POST", "/nodes/pve1/qemu/100/migrate"); ok && !strings.Contains(body, "pve2") {
		t.Errorf("migration request did not carry the target node: %s", body)
	}
}

func TestMigrateGuest_RejectsSameNode(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
	res, _ := migrateGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "target_node": "pve1",
	}))
	assertToolError(t, res, "already runs on node")
}

func TestBackupGuest_CallsAPI(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"POST /nodes/pve1/vzdump": "UPID:pve1:00001234:00ABCDEF:65000000:vzdump:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := backupGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "storage": "local",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)
	if !log.called("POST", "/nodes/pve1/vzdump") {
		t.Errorf("backup_guest did not call vzdump; calls: %v", log.calls())
	}
	body, _ := log.bodyFor("POST", "/nodes/pve1/vzdump")
	for _, want := range []string{"storage=local", "vmid=100"} {
		if !strings.Contains(body, want) {
			t.Errorf("backup request missing %q; body: %s", want, body)
		}
	}
}

func TestBackupGuest_RequiresStorage(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
	res, _ := backupGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	assertToolError(t, res, "storage is required")
}

// TestGuestNotFound covers the shared resolution failure: a vmid absent from
// the cluster must produce a clear error, not a fabricated success.
func TestGuestNotFound(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
	args := map[string]interface{}{
		"vmid": float64(999), "storage": "local",
		"disk": "scsi0", "size": "10G", "target_node": "pve2",
	}

	handlers := map[string]server.ToolHandlerFunc{
		"list_snapshots":          listSnapshotsHandler(client),
		"ping_qemu_agent":         pingQemuAgentHandler(client),
		"backup_guest":            backupGuestHandler(client),
		"resize_disk":             resizeDiskHandler(client),
		"migrate_guest":           migrateGuestHandler(client),
		"get_guest_agent_network": getGuestAgentNetworkHandler(client),
	}

	for name, h := range handlers {
		t.Run(name, func(t *testing.T) {
			res, err := h(context.Background(), newRequest(args))
			if err != nil {
				t.Fatalf("transport error: %v", err)
			}
			assertToolError(t, res, "not found")
		})
	}
}

// TestNodeNameRejectedBeforeReachingAPI pins the path-injection guard.
func TestNodeNameRejectedBeforeReachingAPI(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{})

	bad := []string{"../../access/users", "pve1/../pve2", "pve 1", "pve1?x=1", "pve1#frag", ""}
	for _, node := range bad {
		t.Run("rejects "+node, func(t *testing.T) {
			res, _ := getNodeNetworkHandler(client)(context.Background(), newRequest(map[string]interface{}{"node": node}))
			if res == nil || !res.IsError {
				t.Fatalf("expected node %q to be rejected", node)
			}
		})
	}
	if len(log.calls()) != 0 {
		t.Errorf("invalid node names must never reach Proxmox; calls: %v", log.calls())
	}
}
