package tools

import (
	"context"
	"strings"
	"testing"
)

// Tests for the handlers this remediation changed. Without these the fixes
// were asserted only by reading the code.

// TestRebootNode_DoesNotWaitForTheTask covers the switch away from
// PostWithTask: a node that is restarting cannot report the completion of its
// own reboot task, so waiting could only ever end in a timeout.
func TestRebootNode_DoesNotWaitForTheTask(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"POST /nodes/pve1/status": "UPID:pve1:00001234:00ABCDEF:65000000:srvreboot:pve1:root@pam:",
	})

	res, err := rebootNodeHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "pve1",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	body, ok := log.bodyFor("POST", "/nodes/pve1/status")
	if !ok {
		t.Fatalf("reboot_node did not post to the node status endpoint; calls: %v", log.calls())
	}
	if !strings.Contains(body, "command=reboot") {
		t.Errorf("expected command=reboot, got body: %s", body)
	}

	// The decisive assertion: no task-status polling.
	for _, c := range log.calls() {
		if strings.Contains(c, "/tasks/") {
			t.Errorf("reboot_node must not poll the task of a departing node, but called %s", c)
		}
	}
}

func TestShutdownNode_DoesNotWaitForTheTask(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"POST /nodes/pve1/status": "UPID:pve1:00001234:00ABCDEF:65000000:srvshutdown:pve1:root@pam:",
	})

	res, err := shutdownNodeHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "pve1",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	body, _ := log.bodyFor("POST", "/nodes/pve1/status")
	if !strings.Contains(body, "command=shutdown") {
		t.Errorf("expected command=shutdown, got body: %s", body)
	}
	for _, c := range log.calls() {
		if strings.Contains(c, "/tasks/") {
			t.Errorf("shutdown_node must not poll a task the node cannot answer, but called %s", c)
		}
	}
}

func TestNodeWriteToolsValidateTheNodeName(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{})

	res, _ := rebootNodeHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "../../access/users",
	}))
	if res == nil || !res.IsError {
		t.Error("reboot_node must reject a node name containing a path traversal")
	}

	res, _ = shutdownNodeHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "",
	}))
	if res == nil || !res.IsError {
		t.Error("shutdown_node must reject an empty node name")
	}

	if len(log.calls()) != 0 {
		t.Errorf("invalid node names must never reach Proxmox; calls: %v", log.calls())
	}
}

// TestUpdateGroup_PreservesCommentWhenOmitted covers the data-loss fix: the
// handler used to send Comment unconditionally, so updating a group without
// supplying one wiped its description.
func TestUpdateGroup_PreservesCommentWhenOmitted(t *testing.T) {
	// Omitting the comment must not report success. The SDK issues no request
	// at all when Comment is nil and still returns nil, so a handler that
	// reported success there would be fabricating a result -- and the caller
	// would believe an update happened that never did.
	t.Run("comment omitted is rejected, not silently ignored", func(t *testing.T) {
		client, log := fakePVE(t, map[string]interface{}{})

		res, err := updateGroupHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"groupid": "admins",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		assertToolError(t, res, "comment is required")

		if len(log.calls()) != 0 {
			t.Errorf("nothing should have been sent; calls: %v", log.calls())
		}
	})

	t.Run("comment supplied is sent", func(t *testing.T) {
		client, log := fakePVE(t, map[string]interface{}{
			"GET /access/groups/admins": map[string]interface{}{
				"comment": "Administrators", "members": []interface{}{},
			},
			"PUT /access/groups/admins": nil,
		})

		res, err := updateGroupHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"groupid": "admins", "comment": "Updated",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		assertToolSuccess(t, res)

		body, _ := log.bodyFor("PUT", "/access/groups/admins")
		if !strings.Contains(body, "Updated") {
			t.Errorf("a supplied comment must be written; body: %q", body)
		}
	})
}

// TestClearDiskBandwidth covers the rewritten write path: request context,
// forwarded digest, and removal of every throttle key.
func TestClearDiskBandwidth(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
			"scsi0":  "local-lvm:vm-100-disk-0,size=32G,ssd,mbps_rd=50,iops_wr=200",
			"digest": "cafebabe",
		},
		"POST /nodes/pve1/qemu/100/config": "UPID:pve1:00001234:00ABCDEF:65000000:qmconfig:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := clearDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "disk_id": "scsi0",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	body, ok := log.bodyFor("POST", "/nodes/pve1/qemu/100/config")
	if !ok {
		t.Fatalf("no config write was made; calls: %v", log.calls())
	}
	for _, gone := range []string{"mbps_rd", "iops_wr"} {
		if strings.Contains(body, gone) {
			t.Errorf("%s should have been cleared; body: %s", gone, body)
		}
	}
	// Non-throttle options, including bare flags, must survive.
	for _, kept := range []string{"size", "ssd"} {
		if !strings.Contains(body, kept) {
			t.Errorf("clearing throttles must not drop %q; body: %s", kept, body)
		}
	}
	if !strings.Contains(body, "cafebabe") {
		t.Errorf("the config digest must be forwarded; body: %s", body)
	}
}

func TestClearDiskBandwidth_Validation(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
	h := clearDiskBandwidthHandler(client)

	res, _ := h(context.Background(), newRequest(map[string]interface{}{"disk_id": "scsi0"}))
	assertToolError(t, res, "vmid is required")

	res, _ = h(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
	assertToolError(t, res, "disk_id is required")

	res, _ = h(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100), "disk_id": "nope"}))
	assertToolError(t, res, "invalid disk_id format")
}

// TestGetDiskBandwidth_ReportsFractionalLimits covers the read-side data loss:
// mbps_* are floats in the PVE API, and parsing them with Atoi meant a
// throttled disk was reported as having no throttle at all.
func TestGetDiskBandwidth_ReportsFractionalLimits(t *testing.T) {
	client, _ := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
			"scsi0": "local-lvm:vm-100-disk-0,size=32G,mbps_rd=10.5,iops_wr=200",
		},
	}, qemuGuest(100, "pve1")))

	res, err := getDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "disk_id": "scsi0",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	if !strings.Contains(out, "10.5") {
		t.Errorf("a fractional mbps limit must be reported, not dropped; got: %s", out)
	}
	if !strings.Contains(out, "200") {
		t.Errorf("expected the iops limit in the output; got: %s", out)
	}
	if !strings.Contains(out, "local-lvm") {
		t.Errorf("expected the storage name in the output; got: %s", out)
	}
}

func TestGetDiskBandwidth_AllDisksAndErrors(t *testing.T) {
	t.Run("all disks when disk_id omitted", func(t *testing.T) {
		client, _ := fakePVE(t, withGuests(map[string]interface{}{
			"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
				"scsi0":   "local-lvm:vm-100-disk-0,mbps_rd=50",
				"virtio1": "local-lvm:vm-100-disk-1,mbps_wr=60",
				"net0":    "virtio=AA:BB:CC:DD:EE:FF",
			},
		}, qemuGuest(100, "pve1")))

		res, err := getDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"vmid": float64(100),
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		out := assertToolSuccess(t, res)
		for _, want := range []string{"scsi0", "virtio1"} {
			if !strings.Contains(out, want) {
				t.Errorf("expected %s in the all-disks listing; got: %s", want, out)
			}
		}
		if strings.Contains(out, "net0") {
			t.Errorf("network devices are not disks; got: %s", out)
		}
	})

	t.Run("unknown disk is reported", func(t *testing.T) {
		client, _ := fakePVE(t, withGuests(map[string]interface{}{
			"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
				"scsi0": "local-lvm:vm-100-disk-0",
			},
		}, qemuGuest(100, "pve1")))

		res, _ := getDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"vmid": float64(100), "disk_id": "scsi9",
		}))
		assertToolError(t, res, "not found")
	})

	t.Run("invalid disk id is rejected", func(t *testing.T) {
		client, _ := fakePVE(t, withGuests(map[string]interface{}{}, qemuGuest(100, "pve1")))
		res, _ := getDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"vmid": float64(100), "disk_id": "notadisk",
		}))
		assertToolError(t, res, "invalid disk_id format")
	})
}

// TestDeleteACMEAccount_ChecksTaskStatus covers the discarded exit status:
// deletion is async, so a failing task could return with a nil error and the
// tool reported success anyway.
func TestDeleteACMEAccount_ChecksTaskStatus(t *testing.T) {
	t.Run("failing task is reported as an error", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"DELETE /cluster/acme/account/prod": "UPID:pve1:00001234:00ABCDEF:65000000:acmedel:prod:root@pam:",
			"GET /nodes/pve1/tasks/UPID:pve1:00001234:00ABCDEF:65000000:acmedel:prod:root@pam:/status": map[string]interface{}{
				"status": "stopped", "exitstatus": "unable to delete account",
			},
		})

		res, err := deleteACMEAccountHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"account": "prod",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if res == nil || !res.IsError {
			t.Errorf("a failed deletion task must surface as an error, got: %s", resultText(t, res))
		}
	})

	t.Run("successful task reports success", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"DELETE /cluster/acme/account/prod": "UPID:pve1:00001234:00ABCDEF:65000000:acmedel:prod:root@pam:",
		})

		res, err := deleteACMEAccountHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"account": "prod",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		assertToolSuccess(t, res)
	})
}
