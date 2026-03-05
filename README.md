# Proxmox MCP Server

A Model Context Protocol (MCP) server for Proxmox VE, enabling AI assistants like Claude to interact with your Proxmox cluster.

## Features

- Node, VM, and container management
- Storage and network operations
- High availability and backup management
- User and group management
- **Read-only mode by default for security**

## Installation

```bash
go build -o mcp-proxmox
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

```env
# Proxmox API Configuration
PROXMOX_API_URL=https://your-server:8006/api2/json
PROXMOX_TOKEN_ID=your-token-id
PROXMOX_TOKEN_SECRET=your-token-secret

# Optional: Enable write operations (default: read-only)
# PROXMOX_READ_ONLY=false
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
        "PROXMOX_TOKEN_ID": "your-token-id",
        "PROXMOX_TOKEN_SECRET": "your-token-secret"
      }
    }
  }
}
```

### Enable Write Mode (Optional)

Add `"PROXMOX_READ_ONLY": "false"` to enable write operations like starting/stopping VMs.

## Test with MCP Inspector

```bash
npx @modelcontextprotocol/inspector ./mcp-proxmox
```

## Available Tools

### Read-Only (Available by Default)
- `list_nodes` - List cluster nodes
- `list_guests` - List VMs and containers
- `get_guest_info` - Get guest details
- `list_storage` - List storage
- `list_pools` - List resource pools
- `list_users` - List users
- `list_snapshots` - List snapshots
- And more...

### Write Mode Required
- `start_guest` / `stop_guest` / `reboot_guest`
- `clone_qemu_vm` / `clone_lxc_container`
- `delete_guest` / `create_snapshot`
- `reboot_node` / `shutdown_node`
- And more...

## Development

```bash
go test ./...
go build
./mcp-proxmox
```

## License

MIT
