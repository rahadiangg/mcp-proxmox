package tools

import (
	"context"
	"strings"
	"testing"
)

// TestListGroups_ReturnsRealData is the regression test for a tool that
// returned [{},{}]: the SDK's AsArray values expose only an unexported field,
// so marshalling them silently dropped every group.
func TestListGroups_ReturnsRealData(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /access/groups": []interface{}{
			map[string]interface{}{"groupid": "admins", "comment": "Administrators", "users": "root@pam"},
			map[string]interface{}{"groupid": "ops"},
		},
	})

	res, err := listGroupsHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	for _, want := range []string{"admins", "Administrators", "ops"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the group listing, got: %s", want, out)
		}
	}
	if strings.Contains(out, "{}") {
		t.Errorf("group entries encoded as empty objects: %s", out)
	}
	if !log.called("GET", "/access/groups") {
		t.Errorf("list_groups did not query groups; calls: %v", log.calls())
	}
}

func TestListUsers_ReturnsRealData(t *testing.T) {
	client, log := fakePVE(t, map[string]interface{}{
		"GET /access/users": []interface{}{
			map[string]interface{}{"userid": "root@pam", "enable": float64(1), "comment": "Superuser"},
		},
	})

	res, err := listUsersHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}

	out := assertToolSuccess(t, res)
	for _, want := range []string{"root@pam", "Superuser"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the user listing, got: %s", want, out)
		}
	}
	if strings.Contains(out, "{}") {
		t.Errorf("user entries encoded as empty objects: %s", out)
	}
	if !log.called("GET", "/access/users") {
		t.Errorf("list_users did not query users; calls: %v", log.calls())
	}
}

func TestListGroups_Empty(t *testing.T) {
	client, _ := fakePVE(t, map[string]interface{}{
		"GET /access/groups": []interface{}{},
	})
	res, err := listGroupsHandler(client)(context.Background(), newRequest(nil))
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if got := strings.TrimSpace(assertToolSuccess(t, res)); got != "[]" {
		t.Errorf("expected [], got %q", got)
	}
}
