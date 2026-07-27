package tools

import (
	"context"
	"strings"
	"testing"
)

// TestGetStorageContent_ArrayResponse is the regression test for the crash
// that killed the whole server: the endpoint returns a JSON array, and the
// previous implementation asserted it into a map without a comma-ok.
func TestGetStorageContent_ArrayResponse(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /nodes/pve1/storage/local/content": []interface{}{
			map[string]interface{}{"volid": "local:iso/debian.iso", "content": "iso", "size": float64(1024)},
			map[string]interface{}{"volid": "local:vztmpl/alpine.tar.gz", "content": "vztmpl", "size": float64(512)},
		},
	})

	res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
		"node": "pve1", "storage": "local",
	}))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	if !strings.Contains(out, "debian.iso") || !strings.Contains(out, "alpine.tar.gz") {
		t.Errorf("expected both volumes in the result, got: %s", out)
	}
	if !log.called("GET", "/nodes/pve1/storage/local/content") {
		t.Errorf("handler never queried storage content; calls: %v", log.calls())
	}
}

func TestGetStorageContent_EdgeCases(t *testing.T) {
	t.Run("empty array returns [] not null", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"GET /nodes/pve1/storage/local/content": []interface{}{},
		})
		res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"node": "pve1", "storage": "local",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if got := strings.TrimSpace(assertToolSuccess(t, res)); got != "[]" {
			t.Errorf("expected [], got %q", got)
		}
	})

	t.Run("filter matching nothing returns []", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"GET /nodes/pve1/storage/local/content": []interface{}{
				map[string]interface{}{"volid": "local:iso/debian.iso", "content": "iso"},
			},
		})
		res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"node": "pve1", "storage": "local", "content": "backup",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if got := strings.TrimSpace(assertToolSuccess(t, res)); got != "[]" {
			t.Errorf("expected [] for a non-matching filter, got %q", got)
		}
	})

	t.Run("filter selects matching items only", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"GET /nodes/pve1/storage/local/content": []interface{}{
				map[string]interface{}{"volid": "local:iso/debian.iso", "content": "iso"},
				map[string]interface{}{"volid": "local:backup/vzdump.tar", "content": "backup"},
			},
		})
		res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"node": "pve1", "storage": "local", "content": "iso",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		out := assertToolSuccess(t, res)
		if !strings.Contains(out, "debian.iso") {
			t.Errorf("expected the iso volume, got: %s", out)
		}
		if strings.Contains(out, "vzdump.tar") {
			t.Errorf("content filter was not applied, got: %s", out)
		}
	})

	t.Run("non-object array entries are skipped", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"GET /nodes/pve1/storage/local/content": []interface{}{
				"unexpected-string",
				map[string]interface{}{"volid": "local:iso/debian.iso", "content": "iso"},
			},
		})
		res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"node": "pve1", "storage": "local",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if out := assertToolSuccess(t, res); !strings.Contains(out, "debian.iso") {
			t.Errorf("expected the valid entry to survive, got: %s", out)
		}
	})

	t.Run("item missing content key is dropped when filtering", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{
			"GET /nodes/pve1/storage/local/content": []interface{}{
				map[string]interface{}{"volid": "local:iso/nocontent.iso"},
			},
		})
		res, err := getStorageContentHandler(client)(context.Background(), newRequest(map[string]interface{}{
			"node": "pve1", "storage": "local", "content": "iso",
		}))
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		if got := strings.TrimSpace(assertToolSuccess(t, res)); got != "[]" {
			t.Errorf("expected an item without a content key to be filtered out, got %q", got)
		}
	})

	t.Run("missing required args", func(t *testing.T) {
		client, _ := fakePVE(t, map[string]interface{}{})
		h := getStorageContentHandler(client)

		res, _ := h(context.Background(), newRequest(map[string]interface{}{"storage": "local"}))
		assertToolError(t, res, "node is required")

		res, _ = h(context.Background(), newRequest(map[string]interface{}{"node": "pve1"}))
		assertToolError(t, res, "storage is required")
	})
}

// TestListHAGroups_OptionalComment is the regression test for the second
// process-killing panic: the SDK's typed helper asserts the optional comment
// field into a string, so a group without one crashed the server.
func TestListHAGroups_OptionalComment(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /cluster/ha/groups": []interface{}{
			map[string]interface{}{"group": "ha-nocomment", "nodes": "pve1,pve2"},
			map[string]interface{}{"group": "ha-full", "nodes": "pve1", "comment": "primary", "restricted": float64(1)},
		},
	})

	res, err := listHAGroupsHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	if !strings.Contains(out, "ha-nocomment") || !strings.Contains(out, "ha-full") {
		t.Errorf("expected both HA groups, got: %s", out)
	}
	if !log.called("GET", "/cluster/ha/groups") {
		t.Errorf("handler never queried HA groups; calls: %v", log.calls())
	}
}

func TestListHAGroups_Empty(t *testing.T) {
	client, _ := fakePVE(t, map[string]interface{}{
		"GET /cluster/ha/groups": []interface{}{},
	})
	res, err := listHAGroupsHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if got := strings.TrimSpace(assertToolSuccess(t, res)); got != "[]" {
		t.Errorf("expected [], got %q", got)
	}
}
