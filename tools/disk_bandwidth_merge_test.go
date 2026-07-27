package tools

import (
	"context"
	"strings"
	"testing"
)

// TestSetDiskBandwidth_PreservesUnmentionedLimits is the regression test for
// silent data loss: setting one throttle used to wipe every other throttle on
// that disk, because the rewrite was a full replace rather than a merge.
func TestSetDiskBandwidth_PreservesUnmentionedLimits(t *testing.T) {
	const disk = "local-lvm:vm-100-disk-0,size=32G,mbps_rd=50,mbps_wr=60,iops_rd=500"

	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
			"scsi0":  disk,
			"digest": "abc123",
		},
		"POST /nodes/pve1/qemu/100/config": "UPID:pve1:00001234:00ABCDEF:65000000:qmconfig:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := setDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "disk_id": "scsi0", "mbps_rd": float64(100),
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	body, ok := log.bodyFor("POST", "/nodes/pve1/qemu/100/config")
	if !ok {
		t.Fatalf("no config write was made; calls: %v", log.calls())
	}

	// The value we asked for changed...
	if !strings.Contains(body, "mbps_rd%3D100") && !strings.Contains(body, "mbps_rd=100") {
		t.Errorf("expected mbps_rd=100 in the write; body: %s", body)
	}
	// ...and the ones we never mentioned survived.
	for _, want := range []string{"mbps_wr", "iops_rd", "size"} {
		if !strings.Contains(body, want) {
			t.Errorf("merge dropped %q, which the caller never mentioned; body: %s", want, body)
		}
	}
}

// TestSetDiskBandwidth_DigestIsForwarded covers the lost-update guard.
func TestSetDiskBandwidth_DigestIsForwarded(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
			"scsi0":  "local-lvm:vm-100-disk-0,size=32G",
			"digest": "deadbeef",
		},
		"POST /nodes/pve1/qemu/100/config": "UPID:pve1:00001234:00ABCDEF:65000000:qmconfig:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := setDiskBandwidthHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100), "disk_id": "scsi0", "mbps_rd": float64(50),
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	body, _ := log.bodyFor("POST", "/nodes/pve1/qemu/100/config")
	if !strings.Contains(body, "deadbeef") {
		t.Errorf("config digest must be forwarded so concurrent edits fail loudly; body: %s", body)
	}
}

func TestSetDiskBandwidth_ValueSemantics(t *testing.T) {
	setup := func(t *testing.T) (*requestLog, func(map[string]interface{}) string) {
		t.Helper()
		client, log := fakePVE(t, withGuests(map[string]interface{}{
			"GET /nodes/pve1/qemu/100/config": map[string]interface{}{
				"scsi0":  "local-lvm:vm-100-disk-0,size=32G,mbps_rd=50",
				"digest": "d1",
			},
			"POST /nodes/pve1/qemu/100/config": "UPID:pve1:00001234:00ABCDEF:65000000:qmconfig:100:root@pam:",
		}, qemuGuest(100, "pve1")))

		return log, func(args map[string]interface{}) string {
			args["vmid"] = float64(100)
			args["disk_id"] = "scsi0"
			res, err := setDiskBandwidthHandler(client)(context.Background(), newRequest(args))
			if err != nil {
				t.Fatalf("transport error: %v", err)
			}
			if res.IsError {
				return "ERROR:" + resultText(t, res)
			}
			body, _ := log.bodyFor("POST", "/nodes/pve1/qemu/100/config")
			return body
		}
	}

	t.Run("zero removes the key", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{"mbps_rd": float64(0)})
		if strings.Contains(body, "ERROR:") {
			t.Fatalf("0 should be accepted as unlimited, got %s", body)
		}
		if strings.Contains(body, "mbps_rd") {
			t.Errorf("0 means unlimited, so the key should be removed; body: %s", body)
		}
	})

	t.Run("negative is rejected, not silently dropped", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{"mbps_rd": float64(-5)})
		if !strings.Contains(body, "ERROR:") || !strings.Contains(body, "Must be >= 0") {
			t.Errorf("expected a negative value to be rejected, got: %s", body)
		}
	})

	t.Run("fractional mbps is preserved", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{"mbps_rd": 12.5})
		if strings.Contains(body, "ERROR:") {
			t.Fatalf("fractional mbps should be accepted, got %s", body)
		}
		if !strings.Contains(body, "12.5") {
			t.Errorf("fractional mbps must not be truncated; body: %s", body)
		}
	})

	t.Run("fractional iops is rejected", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{"iops_rd": 10.5})
		if !strings.Contains(body, "ERROR:") || !strings.Contains(body, "whole number") {
			t.Errorf("expected fractional iops to be rejected, got: %s", body)
		}
	})

	t.Run("below minimum is rejected with a readable value", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{"iops_rd": float64(5)})
		if !strings.Contains(body, "ERROR:") || !strings.Contains(body, "Minimum value is 10") {
			t.Errorf("expected a minimum-value rejection, got: %s", body)
		}
		// The old %.0f format rendered 0.4 as "0", i.e. the value documented
		// as "unlimited" appeared to be rejected.
		_, run2 := setup(t)
		body2 := run2(map[string]interface{}{"mbps_rd": 0.4})
		if strings.Contains(body2, ": 0.") == false && strings.Contains(body2, "0.4") == false {
			t.Errorf("rejection should name the actual value 0.4, got: %s", body2)
		}
	})

	t.Run("no parameters at all is rejected", func(t *testing.T) {
		_, run := setup(t)
		body := run(map[string]interface{}{})
		if !strings.Contains(body, "ERROR:") || !strings.Contains(body, "at least one") {
			t.Errorf("expected a rejection when nothing was supplied, got: %s", body)
		}
	})
}

// TestBuildDiskConfigString_Merge covers the pure function directly.
func TestBuildDiskConfigString_Merge(t *testing.T) {
	tests := []struct {
		name     string
		original string
		params   map[string]interface{}
		want     string
	}{
		{
			name:     "omitted keys are preserved",
			original: "st:vol,size=32G,mbps_rd=50,mbps_wr=60",
			params:   map[string]interface{}{"mbps_rd": 100},
			want:     "st:vol,size=32G,mbps_rd=100,mbps_wr=60",
		},
		{
			name:     "nil removes a key",
			original: "st:vol,size=32G,mbps_rd=50,mbps_wr=60",
			params:   map[string]interface{}{"mbps_rd": nil},
			want:     "st:vol,size=32G,mbps_wr=60",
		},
		{
			name:     "bare flags survive",
			original: "st:vol,ssd,size=32G",
			params:   map[string]interface{}{"mbps_rd": 100},
			want:     "st:vol,ssd,size=32G,mbps_rd=100",
		},
		{
			name:     "canonical ordering",
			original: "st:vol",
			params: map[string]interface{}{
				"iops_wr": 20, "mbps_rd": 1, "iops_rd": 10, "mbps_wr": 2,
			},
			want: "st:vol,mbps_rd=1,mbps_wr=2,iops_rd=10,iops_wr=20",
		},
		{
			name:     "no-op write reproduces the input",
			original: "st:vol,size=32G,mbps_rd=50",
			params:   map[string]interface{}{},
			want:     "st:vol,size=32G,mbps_rd=50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDiskConfigString(tt.original, tt.params); got != tt.want {
				t.Errorf("buildDiskConfigString()\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

// TestBuildDiskConfigString_IsDeterministic guards against the map-ordering
// churn that made every write produce a different config line.
func TestBuildDiskConfigString_IsDeterministic(t *testing.T) {
	const original = "st:vol,size=32G,cache=writeback,ssd,mbps_rd=50,iops_wr=200"
	params := map[string]interface{}{"mbps_wr": 75, "iops_rd": 300}

	first := buildDiskConfigString(original, params)
	for i := 0; i < 200; i++ {
		if got := buildDiskConfigString(original, params); got != first {
			t.Fatalf("output is not stable across runs:\n  %s\n  %s", first, got)
		}
	}
}
