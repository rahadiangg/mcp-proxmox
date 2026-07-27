package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// TestCloneLxcContainer_BuildsValidURL is the regression test for a tool that
// could never succeed: the SDK reads the source id from the params map, and
// the handler never set it, so every request went to
// /nodes/pve1/lxc/%!s(<nil>)/clone.
func TestCloneLxcContainer_BuildsValidURL(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"POST /nodes/pve1/lxc/200/clone": "UPID:pve1:00001234:00ABCDEF:65000000:vzclone:200:root@pam:",
	}, lxcGuest(200, "pve1")))

	res, err := cloneLxcContainerHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(200), "new_id": float64(201), "name": "ct-clone",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	if !log.called("POST", "/nodes/pve1/lxc/200/clone") {
		t.Fatalf("clone did not reach the correct URL; calls: %v", log.calls())
	}
	body, _ := log.bodyFor("POST", "/nodes/pve1/lxc/200/clone")
	for _, want := range []string{"newid=201", "hostname=ct-clone"} {
		if !strings.Contains(body, want) {
			t.Errorf("clone request missing %q; body: %s", want, body)
		}
	}
}

// TestCreateTemplate_UsesTemplateEndpoint is the regression test for the other
// never-working tool: it previously wrote template=1 through /config with an
// unresolved VmRef, producing /nodes///<vmid>/config.
func TestCreateTemplate_UsesTemplateEndpoint(t *testing.T) {
	client, log := fakePVE(t, withGuests(map[string]interface{}{
		"POST /nodes/pve1/qemu/100/template": "UPID:pve1:00001234:00ABCDEF:65000000:qmtemplate:100:root@pam:",
	}, qemuGuest(100, "pve1")))

	res, err := createTemplateHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"vmid": float64(100),
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	assertToolSuccess(t, res)

	if !log.called("POST", "/nodes/pve1/qemu/100/template") {
		t.Errorf("expected the template endpoint; calls: %v", log.calls())
	}
	for _, c := range log.calls() {
		if strings.HasSuffix(c, "/config") {
			t.Errorf("create_template must not write through /config; saw %s", c)
		}
	}
}

// TestHibernateIsNotPause pins the distinction the two tools previously
// lacked: hibernate must request suspend-to-disk, pause must not.
func TestHibernateIsNotPause(t *testing.T) {
	t.Run("hibernate sends todisk", func(t *testing.T) {
		client, log := fakePVE(t, withGuests(map[string]interface{}{
			"POST /nodes/pve1/qemu/100/status/suspend": "UPID:pve1:00001234:00ABCDEF:65000000:qmsuspend:100:root@pam:",
		}, qemuGuest(100, "pve1")))

		res, err := hibernateGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		assertToolSuccess(t, res)

		body, ok := log.bodyFor("POST", "/nodes/pve1/qemu/100/status/suspend")
		if !ok {
			t.Fatalf("hibernate did not call suspend; calls: %v", log.calls())
		}
		if !strings.Contains(body, "todisk") {
			t.Errorf("hibernate must request suspend-to-disk; body was %q", body)
		}
	})

	t.Run("pause does not send todisk", func(t *testing.T) {
		client, log := fakePVE(t, withGuests(map[string]interface{}{
			"POST /nodes/pve1/qemu/100/status/suspend": "UPID:pve1:00001234:00ABCDEF:65000000:qmsuspend:100:root@pam:",
		}, qemuGuest(100, "pve1")))

		res, err := pauseGuestHandler(client)(context.Background(), newRequest(map[string]interface{}{"vmid": float64(100)}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		assertToolSuccess(t, res)

		body, _ := log.bodyFor("POST", "/nodes/pve1/qemu/100/status/suspend")
		if strings.Contains(body, "todisk") {
			t.Errorf("pause must suspend to RAM, not disk; body was %q", body)
		}
	})
}

// TestShutdownForceIsHonored covers the flag that was read and thrown away.
func TestShutdownForceIsHonored(t *testing.T) {
	cases := []struct {
		name      string
		args      map[string]interface{}
		wantForce bool
	}{
		{"force true", map[string]interface{}{"vmid": float64(100), "force": true}, true},
		{"force false", map[string]interface{}{"vmid": float64(100), "force": false}, false},
		{"force omitted", map[string]interface{}{"vmid": float64(100)}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, log := fakePVE(t, withGuests(map[string]interface{}{
				"POST /nodes/pve1/qemu/100/status/shutdown": "UPID:pve1:00001234:00ABCDEF:65000000:qmshutdown:100:root@pam:",
			}, qemuGuest(100, "pve1")))

			res, err := shutdownGuestHandler(client)(context.Background(), newRequest(tc.args))
			if err != nil {
				t.Fatalf("transport error: %v", err)
			}
			assertToolSuccess(t, res)

			body, _ := log.bodyFor("POST", "/nodes/pve1/qemu/100/status/shutdown")
			gotForce := strings.Contains(body, "forceStop")
			if gotForce != tc.wantForce {
				t.Errorf("forceStop present = %v, want %v; body: %q", gotForce, tc.wantForce, body)
			}
		})
	}
}

// TestGetNextVmid_Bounded is the regression test for the unbounded request
// flood: the SDK's scan increments forever, breaking only on a non-API error.
func TestGetNextVmid_Bounded(t *testing.T) {
	t.Run("without start_id makes one request", func(t *testing.T) {
		client, log := fakePVE(t, map[string]interface{}{
			"GET /cluster/nextid": "100",
		})
		res, err := getNextVmidHandler(client)(context.Background(), newRequest(nil))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if out := assertToolSuccess(t, res); !strings.Contains(out, "100") {
			t.Errorf("expected the cluster's next id, got: %s", out)
		}
		if n := len(log.calls()); n != 1 {
			t.Errorf("expected exactly one request, got %d: %v", n, log.calls())
		}
	})

	t.Run("start_id out of range is rejected without any request", func(t *testing.T) {
		client, log := fakePVE(t, map[string]interface{}{})
		for _, bad := range []float64{1, 99, 1000000000} {
			res, _ := getNextVmidHandler(client)(context.Background(), newRequest(map[string]interface{}{"start_id": bad}))
			assertToolError(t, res, "must be between")
		}
		if n := len(log.calls()); n != 0 {
			t.Errorf("out-of-range start_id must not reach Proxmox; calls: %v", log.calls())
		}
	})

	t.Run("finds the first free id above start_id", func(t *testing.T) {
		// 100 and 101 are taken, 102 is free.
		client, log := fakePVE(t, map[string]interface{}{
			"GET /cluster/nextid?vmid=100": apiError{Status: 500, Message: "VM 100 already exists"},
			"GET /cluster/nextid?vmid=101": apiError{Status: 500, Message: "VM 101 already exists"},
			"GET /cluster/nextid?vmid=102": "102",
		})

		res, err := getNextVmidHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"start_id": float64(100),
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if out := assertToolSuccess(t, res); !strings.Contains(out, "102") {
			t.Errorf("expected the first free id (102), got: %s", out)
		}
		if n := len(log.calls()); n != 3 {
			t.Errorf("expected exactly 3 probes, got %d: %v", n, log.calls())
		}
	})

	t.Run("probe count is bounded when every id is taken", func(t *testing.T) {
		// Every probe reports the id as taken, which is precisely the input
		// that made the SDK's own scan loop without any bound.
		routes := map[string]interface{}{}
		for i := 0; i < maxVmidProbes+10; i++ {
			routes[fmt.Sprintf("GET /cluster/nextid?vmid=%d", 100+i)] =
				apiError{Status: 500, Message: "already exists"}
		}

		client, log := fakePVE(t, routes)
		res, err := getNextVmidHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"start_id": float64(100),
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if res == nil || !res.IsError {
			t.Fatal("expected an error when no free id is found")
		}
		if n := len(log.calls()); n != maxVmidProbes {
			t.Errorf("probe count must be exactly %d, got %d", maxVmidProbes, n)
		}
	})
}
