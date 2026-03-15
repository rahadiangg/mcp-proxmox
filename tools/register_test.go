package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rahadiangg/mcp-proxmox/proxmox"
)

// Test that all registration functions can be called without panicking
func TestRegisterFunctionsNoPanic(t *testing.T) {
	client := &proxmox.Client{}
	s := server.NewMCPServer("Test Server", "1.0.0")

	t.Run("RegisterNodeTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterNodeTools panicked: %v", r)
			}
		}()
		RegisterNodeTools(s, client)
	})

	t.Run("RegisterGuestTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterGuestTools panicked: %v", r)
			}
		}()
		RegisterGuestTools(s, client)
	})

	t.Run("RegisterStorageTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterStorageTools panicked: %v", r)
			}
		}()
		RegisterStorageTools(s, client)
	})

	t.Run("RegisterPoolTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterPoolTools panicked: %v", r)
			}
		}()
		RegisterPoolTools(s, client)
	})

	t.Run("RegisterHATools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterHATools panicked: %v", r)
			}
		}()
		RegisterHATools(s, client)
	})

	t.Run("RegisterMetricsTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterMetricsTools panicked: %v", r)
			}
		}()
		RegisterMetricsTools(s, client)
	})

	t.Run("RegisterUserTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterUserTools panicked: %v", r)
			}
		}()
		RegisterUserTools(s, client)
	})

	t.Run("RegisterGroupTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterGroupTools panicked: %v", r)
			}
		}()
		RegisterGroupTools(s, client)
	})

	t.Run("RegisterACMETools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterACMETools panicked: %v", r)
			}
		}()
		RegisterACMETools(s, client)
	})

	t.Run("RegisterResourceTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterResourceTools panicked: %v", r)
			}
		}()
		RegisterResourceTools(s, client)
	})

	t.Run("RegisterStorageContentTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterStorageContentTools panicked: %v", r)
			}
		}()
		RegisterStorageContentTools(s, client)
	})

	t.Run("RegisterSnapshotTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterSnapshotTools panicked: %v", r)
			}
		}()
		RegisterSnapshotTools(s, client)
	})

	t.Run("RegisterQemuAgentTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterQemuAgentTools panicked: %v", r)
			}
		}()
		RegisterQemuAgentTools(s, client)
	})

	t.Run("RegisterNodeNetworkTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterNodeNetworkTools panicked: %v", r)
			}
		}()
		RegisterNodeNetworkTools(s, client)
	})

	t.Run("RegisterNetworkTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterNetworkTools panicked: %v", r)
			}
		}()
		RegisterNetworkTools(s, client)
	})

	t.Run("RegisterFirewallTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterFirewallTools panicked: %v", r)
			}
		}()
		RegisterFirewallTools(s, client)
	})

	t.Run("RegisterCreateTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterCreateTools panicked: %v", r)
			}
		}()
		RegisterCreateTools(s, client)
	})

	t.Run("RegisterDiskBandwidthTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterDiskBandwidthTools panicked: %v", r)
			}
		}()
		RegisterDiskBandwidthTools(s, client)
	})

	// Write tools
	t.Run("RegisterLifecycleTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterLifecycleTools panicked: %v", r)
			}
		}()
		RegisterLifecycleTools(s, client)
	})

	t.Run("RegisterCloneTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterCloneTools panicked: %v", r)
			}
		}()
		RegisterCloneTools(s, client)
	})

	t.Run("RegisterDiskTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterDiskTools panicked: %v", r)
			}
		}()
		RegisterDiskTools(s, client)
	})

	t.Run("RegisterMigrateTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterMigrateTools panicked: %v", r)
			}
		}()
		RegisterMigrateTools(s, client)
	})

	t.Run("RegisterBackupTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterBackupTools panicked: %v", r)
			}
		}()
		RegisterBackupTools(s, client)
	})

	t.Run("RegisterGroupWriteTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterGroupWriteTools panicked: %v", r)
			}
		}()
		RegisterGroupWriteTools(s, client)
	})

	t.Run("RegisterACMEWriteTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterACMEWriteTools panicked: %v", r)
			}
		}()
		RegisterACMEWriteTools(s, client)
	})

	t.Run("RegisterNodeWriteTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterNodeWriteTools panicked: %v", r)
			}
		}()
		RegisterNodeWriteTools(s, client)
	})

	t.Run("RegisterDiskBandwidthWriteTools", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterDiskBandwidthWriteTools panicked: %v", r)
			}
		}()
		RegisterDiskBandwidthWriteTools(s, client)
	})
}

func TestNilClientHandling(t *testing.T) {
	s := server.NewMCPServer("Test Server", "1.0.0")

	// Test that registration functions handle nil client gracefully
	t.Run("RegisterNodeTools with nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterNodeTools with nil client panicked: %v", r)
			}
		}()
		RegisterNodeTools(s, nil)
	})

	t.Run("RegisterGuestTools with nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterGuestTools with nil client panicked: %v", r)
			}
		}()
		RegisterGuestTools(s, nil)
	})

	t.Run("RegisterStorageTools with nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterStorageTools with nil client panicked: %v", r)
			}
		}()
		RegisterStorageTools(s, nil)
	})
}
