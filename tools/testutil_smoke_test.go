package tools

import (
	"context"
	"strings"
	"testing"
)

// TestFakePVE_DrivesARealHandler proves the harness works end to end: a real
// handler closure, a real *proxmox.Client, a real HTTP round-trip.
func TestFakePVE_DrivesARealHandler(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /cluster/resources": []interface{}{
			map[string]interface{}{"vmid": float64(100), "name": "web-1", "node": "pve1", "type": "qemu"},
			map[string]interface{}{"vmid": float64(101), "name": "db-1", "node": "pve2", "type": "lxc"},
		},
	})

	res, err := listGuestsHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("handler returned a transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	for _, want := range []string{"web-1", "db-1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected result to contain %q, got: %s", want, out)
		}
	}

	if !log.called("GET", "/cluster/resources") {
		t.Errorf("handler never called Proxmox; recorded calls: %v", log.calls())
	}
}

// TestFakePVE_FiltersAreApplied confirms the recorder plus result assertions
// are sharp enough to catch a handler that ignores its arguments.
func TestFakePVE_FiltersAreApplied(t *testing.T) {
	client, _ := fakePVE(t, map[string]interface{}{
		"GET /cluster/resources": []interface{}{
			map[string]interface{}{"vmid": float64(100), "name": "web-1", "node": "pve1", "type": "qemu"},
			map[string]interface{}{"vmid": float64(101), "name": "db-1", "node": "pve2", "type": "lxc"},
		},
	})

	res, err := listGuestsHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "pve2",
	}))
	if err != nil {
		t.Fatalf("handler returned a transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	if !strings.Contains(out, "db-1") {
		t.Errorf("expected the pve2 guest in the result, got: %s", out)
	}
	if strings.Contains(out, "web-1") {
		t.Errorf("node filter was not applied; got: %s", out)
	}
}
