# Proxmox MCP Server

A Model Context Protocol (MCP) server for Proxmox VE, enabling AI assistants like Claude to interact with your Proxmox cluster.

## Features

- List and inspect nodes, VMs, and containers
- Guest lifecycle: start, stop, shutdown, reboot, pause, resume, hibernate, delete
- Cloning and template creation for both QEMU VMs and LXC containers
- Storage inspection and per-disk bandwidth throttling
- Snapshots, cluster resources, backups, and live migration
- Node and guest-agent network inspection, firewall options
- User, group, pool, HA and ACME listing; group and ACME management
- **Read-only by default** — write tools are not registered unless you opt in
- **TLS verification on by default**

51 tools total: 28 read-only, 23 write. Run `go run ./cmd/tools_list write` for the full list.

## Installation

```bash
go build -o mcp-proxmox
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

> `.env` is resolved against the working directory of whatever launches the
> server. MCP hosts often set that to something other than the project
> directory, so prefer passing environment variables directly (see below).

### Authentication

Use **either** an API token (recommended — tokens do not expire) or a username
and password.

```env
PROXMOX_API_URL=https://your-server:8006/api2/json

# Option 1: API token (recommended)
PROXMOX_TOKEN_ID=root@pam!mcp
PROXMOX_TOKEN_SECRET=your-token-secret

# Option 2: username and password
# The username must include a realm.
PROXMOX_USERNAME=root@pam
PROXMOX_PASSWORD=your-password
```

Password sessions use a Proxmox ticket that expires after about two hours; API
tokens are the supported mode for a long-running server.

### Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `PROXMOX_API_URL` | `https://localhost:8006/api2/json` | Cluster API endpoint |
| `PROXMOX_TOKEN_ID` / `PROXMOX_TOKEN_SECRET` | — | API token authentication |
| `PROXMOX_USERNAME` / `PROXMOX_PASSWORD` | — | Password authentication |
| `PROXMOX_READ_ONLY` | `true` | `false` registers the write tools |
| `PROXMOX_TLS_INSECURE` | `false` | `true` skips certificate verification |
| `PROXMOX_CA_FILE` | — | PEM bundle to trust, e.g. the node's `pve-root-ca.pem` |
| `PROXMOX_HTTP_TIMEOUT` | `30s` | Per-request timeout (`45s`, `2m`, or bare seconds) |
| `PROXMOX_TASK_TIMEOUT` | `300` | Seconds to wait for an async task |

`PROXMOX_READ_ONLY` and `PROXMOX_TLS_INSECURE` accept `true/1/yes/on/enabled/y/t`
and `false/0/no/off/disabled/n/f`, case-insensitively. **Any unrecognized value
falls back to the safe default and logs a warning** — a typo such as `fasle`
cannot silently enable write operations.

### TLS certificates

Proxmox ships a self-signed certificate, so a default install will fail
verification. Pick one:

```env
# Preferred: trust the cluster's own CA.
# Copy /etc/pve/pve-root-ca.pem from a node.
PROXMOX_CA_FILE=/path/to/pve-root-ca.pem

# Or skip verification entirely (not recommended — the API token is
# exposed to anyone who can intercept the connection).
PROXMOX_TLS_INSECURE=true
```

## Usage with Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "proxmox": {
      "command": "/path/to/mcp-proxmox",
      "env": {
        "PROXMOX_API_URL": "https://your-server:8006/api2/json",
        "PROXMOX_TOKEN_ID": "root@pam!mcp",
        "PROXMOX_TOKEN_SECRET": "your-token-secret",
        "PROXMOX_CA_FILE": "/path/to/pve-root-ca.pem"
      }
    }
  }
}
```

### Enable write mode

Add `"PROXMOX_READ_ONLY": "false"` to register the write tools. In read-only
mode they are never advertised, so the model cannot call them at all.

## Test with MCP Inspector

```bash
npx @modelcontextprotocol/inspector ./mcp-proxmox
```

## Available Tools

```bash
go run ./cmd/tools_list        # read-only tools
go run ./cmd/tools_list write  # all tools, labelled read-only or write
```

The listing is generated from the real registration path, so it cannot drift
from what the server actually serves.

### Read-only (available by default)

`list_nodes`, `get_node_status`, `list_guests`, `get_guest_info`,
`get_guest_config`, `get_guest_status`, `get_guest_by_name`, `list_storage`,
`get_storage_status`, `get_storage_config`, `get_storage_content`,
`list_resources`, `list_snapshots`, `list_pools`, `list_ha_groups`,
`list_metrics_servers`, `list_users`, `list_groups`, `get_group`,
`list_acme_accounts`, `get_acme_account`, `list_acme_plugins`,
`get_disk_bandwidth`, `get_next_vmid`, `get_node_network`,
`get_guest_agent_network`, `get_guest_firewall_options`, `ping_qemu_agent`

### Write mode required

`start_guest`, `stop_guest`, `shutdown_guest`, `reboot_guest`, `pause_guest`,
`resume_guest`, `hibernate_guest`, `delete_guest`, `clone_qemu_vm`,
`clone_lxc_container`, `create_template`, `resize_disk`, `set_disk_bandwidth`,
`clear_disk_bandwidth`, `migrate_guest`, `backup_guest`, `create_group`,
`update_group`, `delete_group`, `delete_acme_account`, `delete_acme_plugin`,
`reboot_node`, `shutdown_node`

## Notes on specific tools

- **`set_disk_bandwidth` merges.** Limits you do not mention are preserved.
  Pass `0` to clear an individual limit, or use `clear_disk_bandwidth` to
  remove them all. Writes carry the config digest, so a concurrent edit fails
  loudly instead of being silently overwritten.
- **`get_next_vmid`** asks the cluster directly when `start_id` is omitted.
  With `start_id` it probes upward at most 64 times before giving up.
- **`reboot_node` / `shutdown_node`** return as soon as the request is
  accepted. A departing node cannot report the completion of its own task.
- **`backup_guest` and `migrate_guest`** wait for the task to finish and can
  take a long time; raise `PROXMOX_TASK_TIMEOUT` for large guests.
- **`hibernate_guest`** suspends to disk; `pause_guest` suspends to RAM.

## Development

```bash
go build ./...
go vet ./...
go test ./... -race -cover
```

## License

MIT
