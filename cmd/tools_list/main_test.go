package main

import (
	"strings"
	"testing"
)

// The previous tests here asserted a hardcoded 46-name slice against a second
// hand-written copy of the same list, so they proved nothing about the server
// and had already drifted five tools behind reality.

func TestToolNamesComeFromTheRealRegistration(t *testing.T) {
	s := buildServer(true)
	names := toolNames(s)

	if len(names) != len(s.ListTools()) {
		t.Errorf("toolNames returned %d entries for %d registered tools", len(names), len(s.ListTools()))
	}

	// Tools added after the old hardcoded list was written; their absence is
	// exactly the drift that list suffered from.
	for _, name := range []string{
		"get_storage_status", "get_storage_config",
		"get_disk_bandwidth", "set_disk_bandwidth", "clear_disk_bandwidth",
	} {
		if !contains(names, name) {
			t.Errorf("%s is registered but missing from the listing", name)
		}
	}
}

func TestReadOnlyListingExcludesWriteTools(t *testing.T) {
	readOnly := toolNames(buildServer(false))

	for _, w := range []string{"delete_guest", "stop_guest", "reboot_node", "set_disk_bandwidth"} {
		if contains(readOnly, w) {
			t.Errorf("%s must not appear in the read-only listing", w)
		}
	}

	for _, r := range []string{"list_nodes", "list_guests", "get_disk_bandwidth"} {
		if !contains(readOnly, r) {
			t.Errorf("%s should appear in the read-only listing", r)
		}
	}
}

func TestWriteListingIsASuperset(t *testing.T) {
	readOnly := toolNames(buildServer(false))
	all := toolNames(buildServer(true))

	if len(all) <= len(readOnly) {
		t.Fatalf("write mode should expose more tools: %d vs %d", len(all), len(readOnly))
	}

	for _, n := range readOnly {
		if !contains(all, n) {
			t.Errorf("%s disappeared in write mode", n)
		}
	}
}

func TestDifference(t *testing.T) {
	all := []string{"a", "b", "c"}
	base := []string{"a", "c"}

	got := difference(all, base)
	if len(got) != 1 || got[0] != "b" {
		t.Errorf("difference(%v, %v) = %v, want [b]", all, base, got)
	}
}

func TestToolNamesAreSortedAndUnique(t *testing.T) {
	names := toolNames(buildServer(true))

	seen := make(map[string]bool, len(names))
	for i, n := range names {
		if seen[n] {
			t.Errorf("duplicate tool name %q", n)
		}
		seen[n] = true

		if i > 0 && names[i-1] > n {
			t.Errorf("names are not sorted: %q precedes %q", names[i-1], n)
		}
		if strings.TrimSpace(n) == "" {
			t.Error("empty tool name registered")
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
