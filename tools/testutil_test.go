package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// testUPID is a syntactically valid Proxmox task id. The SDK extracts the node
// name from the second field to build the task-status URL, so it must parse.
const testUPID = "UPID:pve1:00001234:00ABCDEF:65000000:qmstart:100:root@pam:"

var upidNodeRe = regexp.MustCompile(`^UPID:([^:]+):`)

// fakePVE starts an httptest server that speaks just enough of the Proxmox API
// for handler tests, and returns a client pointed at it plus a recorder of the
// requests it received.
//
// routes is keyed by "METHOD /path" where /path excludes the /api2/json prefix
// (e.g. "GET /nodes"). Each value is marshalled as the body's "data" field, so
// a route may return an object, an array, or a bare string.
//
// A POST route whose value is a UPID string automatically gets a matching
// /nodes/<node>/tasks/<upid>/status route returning exitstatus OK, so handlers
// that call PostWithTask complete without extra wiring.
//
// Requests to unregistered routes fail the test rather than returning silently
// wrong data — a handler calling an endpoint the test did not anticipate is a
// bug worth surfacing.
func fakePVE(t *testing.T, routes map[string]interface{}) (*proxmox.Client, *requestLog) {
	t.Helper()

	all := make(map[string]interface{}, len(routes)+1)
	for k, v := range routes {
		all[k] = v
		if s, ok := v.(string); ok && strings.HasPrefix(s, "UPID:") {
			if m := upidNodeRe.FindStringSubmatch(s); m != nil {
				statusPath := fmt.Sprintf("GET /nodes/%s/tasks/%s/status", m[1], s)
				if _, exists := all[statusPath]; !exists {
					all[statusPath] = map[string]interface{}{"exitstatus": "OK", "status": "stopped"}
				}
			}
		}
	}

	log := &requestLog{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api2/json")
		if err := r.ParseForm(); err == nil && len(r.PostForm) > 0 {
			log.record(r.Method, path, r.PostForm.Encode())
		} else {
			log.record(r.Method, path, "")
		}

		// A query-qualified route wins over the bare path, so tests can give
		// different answers per query (e.g. /cluster/nextid?vmid=101).
		body, ok := lookupRoute(all, r.Method, path, r.URL.RawQuery)
		if !ok {
			t.Errorf("fakePVE: unexpected request %s %s (registered: %v)", r.Method, path, keysOf(all))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte(`{"data":null,"message":"route not registered"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if apiErr, isErr := body.(apiError); isErr {
			w.WriteHeader(apiErr.Status)
			_, _ = w.Write([]byte(`{"data":null,"message":"` + apiErr.Message + `"}`))
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": body}); err != nil {
			t.Errorf("fakePVE: encoding response for %s %s: %v", r.Method, path, err)
		}
	}))
	t.Cleanup(srv.Close)

	// Token auth performs no network round-trip, so the client is ready immediately.
	client, err := proxmox.NewClientWithToken(srv.URL+"/api2/json", "root@pam!test", "secret", proxmox.DefaultOptions())
	if err != nil {
		t.Fatalf("fakePVE: building client: %v", err)
	}
	return client, log
}

// requestLog records the requests a fake server received so tests can assert
// that a handler actually called Proxmox, and with what.
type requestLog struct {
	mu   sync.Mutex
	seen []loggedRequest
}

type loggedRequest struct {
	Method string
	Path   string
	Body   string
}

func (l *requestLog) record(method, path, body string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seen = append(l.seen, loggedRequest{Method: method, Path: path, Body: body})
}

// calls returns the recorded requests as "METHOD /path" strings.
func (l *requestLog) calls() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, 0, len(l.seen))
	for _, r := range l.seen {
		out = append(out, r.Method+" "+r.Path)
	}
	return out
}

// bodyFor returns the recorded request body for the first matching call.
func (l *requestLog) bodyFor(method, path string) (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range l.seen {
		if r.Method == method && r.Path == path {
			return r.Body, true
		}
	}
	return "", false
}

// called reports whether "METHOD /path" was requested at least once.
func (l *requestLog) called(method, path string) bool {
	_, ok := l.bodyFor(method, path)
	return ok
}

// apiError is a route value that makes the fake reply with a Proxmox-style
// error, which is how PVE signals things like "this VMID is already taken".
type apiError struct {
	Status  int
	Message string
}

// lookupRoute resolves a request against the registered routes, preferring an
// exact query match, then the bare path, then a trailing-slash variant.
func lookupRoute(routes map[string]interface{}, method, path, rawQuery string) (interface{}, bool) {
	if rawQuery != "" {
		if body, ok := routes[method+" "+path+"?"+rawQuery]; ok {
			return body, true
		}
	}
	if body, ok := routes[method+" "+path]; ok {
		return body, true
	}
	// The SDK builds some URLs with a trailing slash.
	if body, ok := routes[method+" "+strings.TrimSuffix(path, "/")]; ok {
		return body, true
	}
	return nil, false
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// newRequest builds a CallToolRequest carrying the given arguments.
func newRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	}
}

// resultText extracts the text payload of a tool result.
func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil {
		t.Fatal("nil tool result")
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}

// assertToolError fails unless the result is an error containing want.
func assertToolError(t *testing.T, res *mcp.CallToolResult, want string) {
	t.Helper()
	if res == nil || !res.IsError {
		t.Fatalf("expected an error result containing %q, got success: %s", want, resultText(t, res))
	}
	if got := resultText(t, res); !strings.Contains(got, want) {
		t.Errorf("expected error containing %q, got %q", want, got)
	}
}

// assertToolSuccess fails if the result is an error, and returns its text.
func assertToolSuccess(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil {
		t.Fatal("nil tool result")
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, res))
	}
	return resultText(t, res)
}
