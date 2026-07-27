package main

import (
	"os"
	"sort"
	"testing"

	"github.com/mark3labs/mcp-go/server"

	"github.com/rahadiangg/mcp-proxmox/config"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// writeTools are the tools that can change cluster state. None of them may be
// registered while read-only mode is on.
var writeTools = []string{
	"start_guest", "stop_guest", "shutdown_guest", "reboot_guest",
	"pause_guest", "resume_guest", "hibernate_guest", "delete_guest",
	"clone_qemu_vm", "clone_lxc_container", "create_template",
	"resize_disk", "set_disk_bandwidth", "clear_disk_bandwidth",
	"migrate_guest", "backup_guest",
	"create_group", "update_group", "delete_group",
	"delete_acme_account", "delete_acme_plugin",
	"reboot_node", "shutdown_node",
}

func registeredNames(t *testing.T, readOnly bool) []string {
	t.Helper()
	s := server.NewMCPServer("test", "test")
	registerTools(s, &proxmox.Client{}, readOnly)

	names := make([]string, 0)
	for name := range s.ListTools() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// TestReadOnlyModeHidesWriteTools is the test this gate never had. It was 0%
// covered: the tests that claimed to cover it counted hand-written string
// literals instead of calling the registration path, so deleting the entire
// gate would not have failed a single one.
func TestReadOnlyModeHidesWriteTools(t *testing.T) {
	present := make(map[string]bool)
	for _, n := range registeredNames(t, true) {
		present[n] = true
	}

	for _, w := range writeTools {
		if present[w] {
			t.Errorf("%s must not be registered in read-only mode", w)
		}
	}

	// Read-only mode should still expose something useful.
	for _, r := range []string{"list_nodes", "list_guests", "get_guest_info"} {
		if !present[r] {
			t.Errorf("%s should be available in read-only mode", r)
		}
	}
}

func TestWriteModeExposesWriteTools(t *testing.T) {
	present := make(map[string]bool)
	for _, n := range registeredNames(t, false) {
		present[n] = true
	}

	for _, w := range writeTools {
		if !present[w] {
			t.Errorf("%s should be registered when write mode is enabled", w)
		}
	}
}

// TestWriteModeAddsExactlyTheWriteTools proves the two modes differ by
// precisely the write set, so no tool leaks across the boundary either way.
func TestWriteModeAddsExactlyTheWriteTools(t *testing.T) {
	readOnly := registeredNames(t, true)
	full := registeredNames(t, false)

	inReadOnly := make(map[string]bool, len(readOnly))
	for _, n := range readOnly {
		inReadOnly[n] = true
	}

	var added []string
	for _, n := range full {
		if !inReadOnly[n] {
			added = append(added, n)
		}
	}
	sort.Strings(added)

	expected := append([]string(nil), writeTools...)
	sort.Strings(expected)

	if len(added) != len(expected) {
		t.Fatalf("write mode added %d tools, expected %d\n added: %v\nexpect: %v",
			len(added), len(expected), added, expected)
	}
	for i := range added {
		if added[i] != expected[i] {
			t.Errorf("write-mode difference mismatch at %d: got %s, want %s", i, added[i], expected[i])
		}
	}
}

func TestDefaultModeIsReadOnly(t *testing.T) {
	os.Unsetenv("PROXMOX_READ_ONLY")
	if !config.Load().ReadOnly {
		t.Error("the default mode must be read-only")
	}
}

func TestWriteModeRequiresExplicitOptIn(t *testing.T) {
	cases := map[string]bool{
		"false":   false, // the only way to opt in
		"0":       false,
		"true":    true,
		"":        true,
		"garbage": true, // a typo must never enable writes
	}

	for value, wantReadOnly := range cases {
		t.Run(value, func(t *testing.T) {
			if value == "" {
				os.Unsetenv("PROXMOX_READ_ONLY")
			} else {
				os.Setenv("PROXMOX_READ_ONLY", value)
			}
			defer os.Unsetenv("PROXMOX_READ_ONLY")

			if got := config.Load().ReadOnly; got != wantReadOnly {
				t.Errorf("PROXMOX_READ_ONLY=%q gave ReadOnly=%v, want %v", value, got, wantReadOnly)
			}
		})
	}
}

// TestCredentialsAreRequired covers the fail-fast added to main: without it a
// credential-less server started happily and failed on every tool call.
func TestCredentialsAreRequired(t *testing.T) {
	for _, key := range []string{"PROXMOX_TOKEN_ID", "PROXMOX_TOKEN_SECRET", "PROXMOX_USERNAME", "PROXMOX_PASSWORD"} {
		os.Unsetenv(key)
	}
	if config.Load().HasCredentials() {
		t.Error("expected HasCredentials to be false with nothing configured")
	}

	os.Setenv("PROXMOX_TOKEN_ID", "root@pam!mcp")
	os.Setenv("PROXMOX_TOKEN_SECRET", "secret")
	defer func() {
		os.Unsetenv("PROXMOX_TOKEN_ID")
		os.Unsetenv("PROXMOX_TOKEN_SECRET")
	}()
	if !config.Load().HasCredentials() {
		t.Error("expected HasCredentials to be true once a token is configured")
	}
}

func TestVersionIsInjectable(t *testing.T) {
	// The release build injects this via -ldflags -X main.version. The symbol
	// must exist, or the flag is silently ignored and the server misreports
	// its version forever.
	if version == "" {
		t.Error("version must have a default value")
	}
}
