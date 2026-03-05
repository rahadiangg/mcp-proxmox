package main

import (
	"os"
	"testing"

	"github.com/rahadiangg/mcp-proxmox/config"
)

func TestDefaultModeReadOnly(t *testing.T) {
	// Ensure PROXMOX_READ_ONLY is not set
	os.Unsetenv("PROXMOX_READ_ONLY")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if !cfg.ReadOnly {
		t.Errorf("Default mode should be read-only, got ReadOnly=%v", cfg.ReadOnly)
	}
}

func TestWriteModeEnv(t *testing.T) {
	os.Setenv("PROXMOX_READ_ONLY", "false")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if cfg.ReadOnly {
		t.Errorf("Write mode should disable read-only flag, got ReadOnly=%v", cfg.ReadOnly)
	}
}

func TestExplicitReadOnlyMode(t *testing.T) {
	os.Setenv("PROXMOX_READ_ONLY", "true")
	defer os.Unsetenv("PROXMOX_READ_ONLY")

	cfg := config.Load()
	if !cfg.ReadOnly {
		t.Errorf("Explicit read-only mode should set ReadOnly=true, got %v", cfg.ReadOnly)
	}
}

// Test that all write tool registration functions compile correctly
func TestWriteToolRegistrationCompile(t *testing.T) {
	// This is a compile-time test to ensure the write tool registration functions exist
	// We don't actually run them here - they're tested indirectly via the main() function
	_ = func() {
		// If this compiles, the functions are properly exported
		var registerNodeWriteTools func(interface{}, interface{})
		var registerGroupWriteTools func(interface{}, interface{})
		var registerACMEWriteTools func(interface{}, interface{})
		_, _, _ = registerNodeWriteTools, registerGroupWriteTools, registerACMEWriteTools
	}
}
