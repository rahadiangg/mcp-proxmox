// Command tools_list prints the MCP tools this server exposes, in read-only
// mode by default or in write mode when passed "write".
//
// The listing is derived from the real registration path rather than a
// hand-maintained slice, so it cannot drift from what the server actually
// serves. The previous implementation returned a hardcoded list that had
// already fallen five tools behind.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	"github.com/rahadiangg/mcp-proxmox/tools"
)

func main() {
	writeMode := len(os.Args) > 1 && os.Args[1] == "write"

	readOnlyNames := toolNames(buildServer(false))
	allNames := toolNames(buildServer(true))
	writeNames := difference(allNames, readOnlyNames)

	mode := "read-only"
	names := readOnlyNames
	if writeMode {
		mode = "write"
		names = allNames
	}

	fmt.Printf("Proxmox MCP Server - %s mode\n", mode)
	fmt.Printf("%d tools available (%d read-only, %d write)\n\n", len(names), len(readOnlyNames), len(writeNames))

	writeSet := make(map[string]bool, len(writeNames))
	for _, n := range writeNames {
		writeSet[n] = true
	}

	for _, name := range names {
		kind := "read-only"
		if writeSet[name] {
			kind = "write"
		}
		fmt.Printf("  %-32s %s\n", name, kind)
	}

	if !writeMode {
		fmt.Printf("\n%d write tools are hidden. Run with \"write\" to list them.\n", len(writeNames))
	}
}

// buildServer registers the real tool set for the given mode.
func buildServer(writeMode bool) *server.MCPServer {
	s := server.NewMCPServer("Proxmox MCP Server", "tools-list", server.WithToolCapabilities(true))
	client := &proxmox.Client{}

	tools.RegisterNodeTools(s, client)
	tools.RegisterGuestTools(s, client)
	tools.RegisterStorageTools(s, client)
	tools.RegisterPoolTools(s, client)
	tools.RegisterHATools(s, client)
	tools.RegisterMetricsTools(s, client)
	tools.RegisterUserTools(s, client)
	tools.RegisterGroupTools(s, client)
	tools.RegisterACMETools(s, client)
	tools.RegisterResourceTools(s, client)
	tools.RegisterStorageContentTools(s, client)
	tools.RegisterSnapshotTools(s, client)
	tools.RegisterQemuAgentTools(s, client)
	tools.RegisterNodeNetworkTools(s, client)
	tools.RegisterNetworkTools(s, client)
	tools.RegisterFirewallTools(s, client)
	tools.RegisterCreateTools(s, client)
	tools.RegisterDiskBandwidthTools(s, client)

	if writeMode {
		tools.RegisterLifecycleTools(s, client)
		tools.RegisterCloneTools(s, client)
		tools.RegisterDiskTools(s, client)
		tools.RegisterDiskBandwidthWriteTools(s, client)
		tools.RegisterMigrateTools(s, client)
		tools.RegisterBackupTools(s, client)
		tools.RegisterGroupWriteTools(s, client)
		tools.RegisterACMEWriteTools(s, client)
		tools.RegisterNodeWriteTools(s, client)
	}

	return s
}

// toolNames returns the sorted names actually registered on a server.
func toolNames(s *server.MCPServer) []string {
	names := make([]string, 0)
	for name := range s.ListTools() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// difference returns the entries of all that are absent from base.
func difference(all, base []string) []string {
	inBase := make(map[string]bool, len(base))
	for _, n := range base {
		inBase[n] = true
	}

	var out []string
	for _, n := range all {
		if !inBase[n] {
			out = append(out, n)
		}
	}
	return out
}
