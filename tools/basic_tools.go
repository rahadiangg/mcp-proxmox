package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
	px "github.com/Telmate/proxmox-api-go/proxmox"
)

// diskSizePattern matches the size argument accepted by the PVE resize API:
// an absolute target ("100G") or a relative increase ("+10G"). Proxmox cannot
// shrink a disk, so a negative form is deliberately not accepted.
var diskSizePattern = regexp.MustCompile(`^\+?\d+(\.\d+)?[KMGT]?$`)

func RegisterBackupTools(s *server.MCPServer, client *proxmox.Client) {
	backupGuestTool := mcp.NewTool("backup_guest",
		mcp.WithDescription("Create a vzdump backup of a VM or container. Waits for the backup task to finish, which can take a long time for large guests."),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to back up"),
		),
		mcp.WithString("storage",
			mcp.Required(),
			mcp.Description("Target storage for the backup (e.g., 'local')"),
		),
		mcp.WithString("mode",
			mcp.Description("Backup mode: 'snapshot' (default), 'suspend', or 'stop'"),
			mcp.Enum("snapshot", "suspend", "stop"),
		),
		mcp.WithString("compress",
			mcp.Description("Compression algorithm: 'zstd', 'lzo', 'gzip', or '0' for none"),
			mcp.Enum("zstd", "lzo", "gzip", "0"),
		),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(backupGuestTool, backupGuestHandler(client))
}

func RegisterFirewallTools(s *server.MCPServer, client *proxmox.Client) {
	getGuestFirewallOptionsTool := mcp.NewTool("get_guest_firewall_options",
		mcp.WithDescription("Get firewall options for a QEMU VM"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getGuestFirewallOptionsTool, getGuestFirewallOptionsHandler(client))
}

func RegisterDiskTools(s *server.MCPServer, client *proxmox.Client) {
	resizeDiskTool := mcp.NewTool("resize_disk",
		mcp.WithDescription("Grow a guest disk. Proxmox cannot shrink disks, so only absolute targets larger than the current size or relative increases are accepted."),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithString("disk",
			mcp.Required(),
			mcp.Description("Disk identifier (e.g., 'scsi0', 'virtio0')"),
		),
		mcp.WithString("size",
			mcp.Required(),
			mcp.Description("New size as an absolute value ('100G') or a relative increase ('+10G')"),
		),
		mcp.WithDestructiveHintAnnotation(true),
	)
	s.AddTool(resizeDiskTool, resizeDiskHandler(client))
}

func RegisterNetworkTools(s *server.MCPServer, client *proxmox.Client) {
	getGuestAgentNetworkTool := mcp.NewTool("get_guest_agent_network",
		mcp.WithDescription("Get network interfaces reported by the QEMU guest agent. Requires the agent to be installed and running inside the VM."),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getGuestAgentNetworkTool, getGuestAgentNetworkHandler(client))
}

func RegisterMigrateTools(s *server.MCPServer, client *proxmox.Client) {
	migrateGuestTool := mcp.NewTool("migrate_guest",
		mcp.WithDescription("Migrate a VM or container to another node. Waits for the migration task to finish."),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID to migrate"),
		),
		mcp.WithString("target_node",
			mcp.Required(),
			mcp.Description("Destination node name (e.g., 'pve2')"),
		),
		mcp.WithBoolean("online",
			mcp.Description("Perform a live migration without stopping the guest (default: false)"),
			mcp.DefaultBool(false),
		),
		mcp.WithDestructiveHintAnnotation(true),
	)
	s.AddTool(migrateGuestTool, migrateGuestHandler(client))
}

func RegisterNodeNetworkTools(s *server.MCPServer, client *proxmox.Client) {
	getNodeNetworkTool := mcp.NewTool("get_node_network",
		mcp.WithDescription("Get the network interface configuration of a node"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name (e.g., 'pve1')"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by interface type (e.g., 'bridge', 'bond', 'eth')"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getNodeNetworkTool, getNodeNetworkHandler(client))
}

func RegisterQemuAgentTools(s *server.MCPServer, client *proxmox.Client) {
	pingQemuAgentTool := mcp.NewTool("ping_qemu_agent",
		mcp.WithDescription("Ping the QEMU guest agent to check whether it is responding"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM ID"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(pingQemuAgentTool, pingQemuAgentHandler(client))
}

func RegisterResourceTools(s *server.MCPServer, client *proxmox.Client) {
	listResourcesTool := mcp.NewTool("list_resources",
		mcp.WithDescription("List cluster resources (guests, nodes, storage, pools)"),
		mcp.WithString("type",
			mcp.Description("Filter by resource type"),
			mcp.Enum("vm", "storage", "node", "sdn"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listResourcesTool, listResourcesHandler(client))
}

func RegisterStorageContentTools(s *server.MCPServer, client *proxmox.Client) {
	getStorageContentTool := mcp.NewTool("get_storage_content",
		mcp.WithDescription("Get storage contents including disks, ISOs, and templates"),
		mcp.WithString("node",
			mcp.Required(),
			mcp.Description("Node name where the storage is located (e.g., 'pve1')"),
		),
		mcp.WithString("storage",
			mcp.Required(),
			mcp.Description("Storage name (e.g., 'local', 'local-lvm')"),
		),
		mcp.WithString("content",
			mcp.Description("Filter by content type - optional"),
			mcp.Enum("images", "iso", "rootdir", "snippets", "backup", "vztmpl"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getStorageContentTool, getStorageContentHandler(client))
}

func RegisterSnapshotTools(s *server.MCPServer, client *proxmox.Client) {
	listSnapshotsTool := mcp.NewTool("list_snapshots",
		mcp.WithDescription("List snapshots of a VM or container"),
		mcp.WithNumber("vmid",
			mcp.Required(),
			mcp.Description("VM or container ID"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listSnapshotsTool, listSnapshotsHandler(client))
}

func backupGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		storage, err := getRequiredNameParam(req, "storage")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]interface{}{
			"vmid":    vmid,
			"storage": storage,
			"mode":    req.GetString("mode", "snapshot"),
		}
		if compress := req.GetString("compress", ""); compress != "" {
			params["compress"] = compress
		}

		// No SDK wrapper exists for vzdump, so post the endpoint directly.
		url := fmt.Sprintf("/nodes/%s/vzdump", vmr.Node())
		status, err := client.PostWithTask(ctx, params, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to back up guest %d: %v", vmid, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Backup of guest %d to storage %q completed with status: %s", vmid, storage, status)), nil
	}
}

func getGuestFirewallOptionsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := requireQemu(vmr, "get_guest_firewall_options"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.Node(), vmid)
		envelope, err := client.GetItemList(ctx, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get firewall options for guest %d: %v", vmid, err)), nil
		}

		return jsonResult(unwrapData(envelope), "firewall options")
	}
}

func resizeDiskHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		disk := req.GetString("disk", "")
		if disk == "" {
			return mcp.NewToolResultError("disk is required"), nil
		}
		if !isValidDiskID(disk) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid disk format: '%s'. Expected a disk type prefix (scsi, virtio, sata, ide, ata) followed by a number (e.g., 'scsi0')", disk)), nil
		}

		size := req.GetString("size", "")
		if size == "" {
			return mcp.NewToolResultError("size is required"), nil
		}
		if !diskSizePattern.MatchString(size) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid size: '%s'. Expected an absolute size like '100G' or a relative increase like '+10G'. Proxmox cannot shrink disks", size)), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		status, err := client.ResizeQemuDiskRaw(ctx, vmr, disk, size)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to resize disk %s on guest %d: %v", disk, vmid, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Disk %s on guest %d resized to %s. Status: %v", disk, vmid, size, status)), nil
	}
}

func getGuestAgentNetworkHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := requireQemu(vmr, "get_guest_agent_network"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", vmr.Node(), vmid)
		envelope, err := client.GetItemList(ctx, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get agent network interfaces for guest %d (is the guest agent running?): %v", vmid, err)), nil
		}

		return jsonResult(unwrapData(envelope), "agent network interfaces")
	}
}

func migrateGuestHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		targetNode, err := getRequiredNameParam(req, "target_node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if string(vmr.Node()) == targetNode {
			return mcp.NewToolResultError(fmt.Sprintf("guest %d already runs on node %q", vmid, targetNode)), nil
		}

		online := req.GetBool("online", false)
		status, err := client.MigrateNode(ctx, vmr, px.NodeName(targetNode), online)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to migrate guest %d to %s: %v", vmid, targetNode, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest %d migrated to node %s (online=%t). Status: %v", vmid, targetNode, online, status)), nil
	}
}

func getNodeNetworkHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node, err := getRequiredNameParam(req, "node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		url := fmt.Sprintf("/nodes/%s/network", node)
		if typeFilter := req.GetString("type", ""); typeFilter != "" {
			if !isValidPVEName(typeFilter) {
				return mcp.NewToolResultError(fmt.Sprintf("invalid type filter: %q", typeFilter)), nil
			}
			url += "?type=" + typeFilter
		}

		envelope, err := client.GetItemList(ctx, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get network configuration for node %s: %v", node, err)), nil
		}

		return jsonResult(unwrapData(envelope), "node network configuration")
	}
}

func pingQemuAgentHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := requireQemu(vmr, "ping_qemu_agent"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Posted directly rather than via QemuAgentPing, which type-asserts the
		// response without a comma-ok and swallows its own error via shadowing.
		url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/ping", vmr.Node(), vmid)
		if _, err := client.PostWithTask(ctx, nil, url); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Guest agent on VM %d did not respond: %v", vmid, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Guest agent on VM %d responded to ping", vmid)), nil
	}
}

func listResourcesHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resourceType := req.GetString("type", "")
		if resourceType != "" && !isValidPVEName(resourceType) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid resource type: %q", resourceType)), nil
		}

		resources, err := client.GetResourceList(ctx, resourceType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list resources: %v", err)), nil
		}
		if resources == nil {
			resources = []interface{}{}
		}

		return jsonResult(resources, "cluster resources")
	}
}

func getStorageContentHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node, err := getRequiredNameParam(req, "node")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		storage, err := getRequiredNameParam(req, "storage")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		contentType := req.GetString("content", "")

		url := fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage)
		// This endpoint returns a JSON array. GetItemConfigMapStringInterface
		// would assert it into a map without a comma-ok and panic. Of the two
		// array helpers, only GetItemListInterfaceArray casts with a comma-ok,
		// so an unexpected response shape becomes an error, not a crash.
		items, err := client.GetItemListInterfaceArray(ctx, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get storage content: %v", err)), nil
		}

		// Non-nil so an empty or fully filtered result encodes as [], not null.
		result := []map[string]interface{}{}
		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if contentType != "" {
				if itemContentType, ok := itemMap["content"].(string); !ok || itemContentType != contentType {
					continue
				}
			}
			result = append(result, itemMap)
		}

		return jsonResult(result, "storage content")
	}
}

func listSnapshotsHandler(client *proxmox.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vmid, err := getRequiredIntParam(req, "vmid")
		if err != nil {
			return mcp.NewToolResultError("vmid is required"), nil
		}

		vmr, err := resolveGuest(ctx, client, vmid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// The raw endpoint works for both QEMU and LXC, unlike the SDK's
		// ListQemuSnapshot which is QEMU-only and drops the caller's context.
		url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot", vmr.Node(), vmr.GetVmType(), vmid)
		envelope, err := client.GetItemList(ctx, url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list snapshots for guest %d: %v", vmid, err)), nil
		}

		snapshots := []interface{}{}
		if data, ok := unwrapData(envelope).([]interface{}); ok {
			snapshots = data
		}

		return jsonResult(snapshots, "snapshots")
	}
}

// jsonResult renders a payload as indented JSON, reporting an encoding failure
// as a tool error rather than returning an empty body.
func jsonResult(payload interface{}, what string) (*mcp.CallToolResult, error) {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to encode %s: %v", what, err)), nil
	}
	return mcp.NewToolResultText(string(encoded)), nil
}
