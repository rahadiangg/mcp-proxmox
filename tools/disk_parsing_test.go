package tools

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsValidDiskID_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		diskID   string
		expected bool
	}{
		// Valid disk IDs
		{"scsi0 valid", "scsi0", true},
		{"scsi10 valid", "scsi10", true},
		{"scsi99 valid", "scsi99", true},
		{"virtio0 valid", "virtio0", true},
		{"virtio15 valid", "virtio15", true},
		{"sata0 valid", "sata0", true},
		{"sata5 valid", "sata5", true},
		{"ide0 valid", "ide0", true},
		{"ide2 valid", "ide2", true},
		{"ata0 valid", "ata0", true},
		{"ata3 valid", "ata3", true},

		// Invalid disk IDs - wrong prefixes
		{"nvme0 invalid prefix", "nvme0", false},
		{"net0 invalid prefix", "net0", false},
		{"usb0 invalid prefix", "usb0", false},
		{"sd0 invalid prefix", "sd0", false},
		{"hd0 invalid prefix", "hd0", false},
		{"fd0 invalid prefix", "fd0", false},
		{"cdrom0 invalid prefix", "cdrom0", false},

		// Invalid disk IDs - missing number
		{"scsi no number", "scsi", false},
		{"virtio no number", "virtio", false},
		{"sata no number", "sata", false},
		{"ide no number", "ide", false},
		{"ata no number", "ata", false},

		// Invalid disk IDs - special characters
		{"scsi-0 with hyphen", "scsi-0", false},
		{"scsi_0 with underscore", "scsi_0", false},
		{"scsi.0 with dot", "scsi.0", false},
		{"scsi 0 with space", "scsi 0", false},

		// Invalid disk IDs - empty
		{"empty string", "", false},

		// Invalid disk IDs - number only
		{"0 only", "0", false},
		{"123 only", "123", false},

		// Invalid disk IDs - prefix with number but in wrong format
		{"0scsi0 reversed", "0scsi0", false},
		{"scsi010 valid", "scsi010", true}, // valid (disk 10)
		{"scsi00 valid", "scsi00", true}, // valid (disk 0)
		{"scsi01 valid", "scsi01", true}, // valid (disk 1)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDiskID(tt.diskID)
			if result != tt.expected {
				t.Errorf("isValidDiskID(%q) = %v; want %v", tt.diskID, result, tt.expected)
			}
		})
	}
}

func TestIsDiskKey_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		// Valid disk keys
		{"scsi0", "scsi0", true},
		{"virtio1", "virtio1", true},
		{"sata2", "sata2", true},
		{"ide3", "ide3", true},
		{"ata4", "ata4", true},
		{"scsi15", "scsi15", true},
		{"scsi99", "scsi99", true},
		{"scsi00", "scsi00", true},

		// Invalid keys
		{"net0", "net0", false},
		{"net1", "net1", false},
		{"pci0", "pci0", false},
		{"usb0", "usb0", false},
		{"scsi", "scsi", false},
		{"", "", false},
		{"scsi-0", "scsi-0", false},
		{"scsi0net0", "scsi0net0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDiskKey(tt.key)
			if result != tt.expected {
				t.Errorf("isDiskKey(%q) = %v; want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestParseStorageFromConfig_RealWorld(t *testing.T) {
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
	}{
		{
			name:       "local-lvm typical",
			diskConfig: "local-lvm:vm-100-disk-0",
			storage:    "local-lvm",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "local typical",
			diskConfig: "local:vm-100-disk-0",
			storage:    "local",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "nfs storage",
			diskConfig: "nfs:vm-100-disk-0",
			storage:    "nfs",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "ceph storage",
			diskConfig: "ceph:vm-100-disk-0",
			storage:    "ceph",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "zfs storage",
			diskConfig: "zfs:vm-100-disk-0",
			storage:    "zfs",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "with subdirectory",
			diskConfig: "local:subdir/vm-100-disk-0",
			storage:    "local",
			volumePath: "subdir/vm-100-disk-0",
		},
		{
			name:       "iso file",
			diskConfig: "local:iso/debian-12.iso",
			storage:    "local",
			volumePath: "iso/debian-12.iso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, volumePath := parseStorageFromConfig(tt.diskConfig)
			if storage != tt.storage || volumePath != tt.volumePath {
				t.Errorf("parseStorageFromConfig(%q) = (%q, %q); want (%q, %q)",
					tt.diskConfig, storage, volumePath, tt.storage, tt.volumePath)
			}
		})
	}
}

func TestParseDiskBandwidth_RealWorld(t *testing.T) {
	tests := []struct {
		name           string
		diskConfig     string
		expectedParams map[string]interface{}
	}{
		{
			name:       "no bandwidth - simple config",
			diskConfig: "local-lvm:vm-100-disk-0",
			expectedParams: map[string]interface{}{},
		},
		{
			name:       "read limit only",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100",
			expectedParams: map[string]interface{}{"mbps_rd": 100},
		},
		{
			name:       "write limit only",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_wr=200",
			expectedParams: map[string]interface{}{"mbps_wr": 200},
		},
		{
			name:       "read/write limits",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_wr=200",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
				"mbps_wr": 200,
			},
		},
		{
			name:       "with burst limits",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_rd_max=150",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
				"mbps_rd_max": 150,
			},
		},
		{
			name:       "iops limits",
			diskConfig: "local-lvm:vm-100-disk-0,iops_rd=1000,iops_wr=2000",
			expectedParams: map[string]interface{}{
				"iops_rd": 1000,
				"iops_wr": 2000,
			},
		},
		{
			name:       "mixed with other params",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,size=32G,ssd=1",
			expectedParams: map[string]interface{}{"mbps_rd": 100},
		},
		{
			name:       "zero values (unlimited)",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=0,mbps_wr=0",
			expectedParams: map[string]interface{}{
				"mbps_rd": 0,
				"mbps_wr": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDiskBandwidth(tt.diskConfig)
			if !reflect.DeepEqual(result, tt.expectedParams) {
				t.Errorf("parseDiskBandwidth(%q) = %v; want %v", tt.diskConfig, result, tt.expectedParams)
			}
		})
	}
}

func TestRemoveBandwidthParams_RealWorld(t *testing.T) {
	tests := []struct {
		name           string
		originalConfig string
		expectedConfig string
	}{
		{
			name:           "simple with bandwidth",
			originalConfig: "local:vm-100-disk-0,mbps_rd=100",
			expectedConfig: "local:vm-100-disk-0",
		},
		{
			name:           "with other params",
			originalConfig: "local:vm-100-disk-0,mbps_rd=100,size=32G",
			expectedConfig: "local:vm-100-disk-0,size=32G",
		},
		{
			name:           "all bandwidth types",
			originalConfig: "local:vm-100-disk-0,mbps_rd=100,mbps_wr=200,iops_rd=500,iops_wr=800",
			expectedConfig: "local:vm-100-disk-0",
		},
		{
			name:           "with cache and size",
			originalConfig: "local:vm-100-disk-0,mbps_rd=100,cache=writeback,size=32G",
			expectedConfig: "local:vm-100-disk-0,cache=writeback,size=32G",
		},
		{
			name:           "only non-bandwidth params",
			originalConfig: "local:vm-100-disk-0,size=32G,ssd=1",
			expectedConfig: "local:vm-100-disk-0,size=32G,ssd=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeBandwidthParams(tt.originalConfig)
			if result != tt.expectedConfig {
				t.Errorf("removeBandwidthParams(%q) = %q; want %q",
					tt.originalConfig, result, tt.expectedConfig)
			}
		})
	}
}

func TestBandwidthConfigString_Parts(t *testing.T) {
	// Test that buildDiskConfigString correctly handles the parts
	originalConfig := "storage:volume,size=32G"
	bandwidthParams := map[string]interface{}{"mbps_rd": 100}

	result := buildDiskConfigString(originalConfig, bandwidthParams)

	// Check that all expected parts are present
	if !strings.Contains(result, "storage:volume") {
		t.Error("Result should contain base storage config")
	}
	if !strings.Contains(result, "size=32G") {
		t.Error("Result should contain size param")
	}
	if !strings.Contains(result, "mbps_rd=100") {
		t.Error("Result should contain bandwidth param")
	}

	// Check that result has exactly 3 parts (base + size + bandwidth)
	parts := strings.Split(result, ",")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d: %v", len(parts), parts)
	}
}

func TestDiskConfigParsing_SpecialCases(t *testing.T) {
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
	}{
		{
			name:       "storage name with dash",
			diskConfig: "local-lvm:vm-100-disk-0",
			storage:    "local-lvm",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "storage name with underscore",
			diskConfig: "my_storage:volume",
			storage:    "my_storage",
			volumePath: "volume",
		},
		{
			name:       "storage name with numbers",
			diskConfig: "storage123:volume",
			storage:    "storage123",
			volumePath: "volume",
		},
		{
			name:       "volume with vmid",
			diskConfig: "local:vm-500-disk-1",
			storage:    "local",
			volumePath: "vm-500-disk-1",
		},
		{
			name:       "volume with extension",
			diskConfig: "local:vm-100-disk-0.raw",
			storage:    "local",
			volumePath: "vm-100-disk-0.raw",
		},
		{
			name:       "volume with qcow2 extension",
			diskConfig: "local:vm-100-disk-0.qcow2",
			storage:    "local",
			volumePath: "vm-100-disk-0.qcow2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, volumePath := parseStorageFromConfig(tt.diskConfig)
			if storage != tt.storage || volumePath != tt.volumePath {
				t.Errorf("parseStorageFromConfig(%q) = (%q, %q); want (%q, %q)",
					tt.diskConfig, storage, volumePath, tt.storage, tt.volumePath)
			}
		})
	}
}

func TestDiskConfig_EdgeCaseParsing(t *testing.T) {
	// Test edge cases in disk config parsing
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
		bandwidth  map[string]bool
	}{
		{
			name:       "config with all bandwidth types",
			diskConfig: "local:vol,mbps_rd=100,mbps_rd_max=150,mbps_wr=200,mbps_wr_max=250,iops_rd=500,iops_rd_max=600,iops_rd_max_length=60,iops_wr=800,iops_wr_max=900,iops_wr_max_length=120",
			storage:    "local",
			volumePath: "vol",
			bandwidth: map[string]bool{
				"mbps_rd": true, "mbps_rd_max": true, "mbps_wr": true, "mbps_wr_max": true,
				"iops_rd": true, "iops_rd_max": true, "iops_rd_max_length": true,
				"iops_wr": true, "iops_wr_max": true, "iops_wr_max_length": true,
			},
		},
		{
			name:       "config with duplicate bandwidth param",
			diskConfig: "local:vol,mbps_rd=100,mbps_rd=200",
			storage:    "local",
			volumePath: "vol",
			bandwidth: map[string]bool{
				"mbps_rd": true, // Should have the last value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, volumePath := parseStorageFromConfig(tt.diskConfig)
			if storage != tt.storage || volumePath != tt.volumePath {
				t.Errorf("parseStorageFromConfig(%q) = (%q, %q); want (%q, %q)",
					tt.diskConfig, storage, volumePath, tt.storage, tt.volumePath)
			}

			bw := parseDiskBandwidth(tt.diskConfig)
			for param := range tt.bandwidth {
				if _, exists := bw[param]; !exists {
					t.Errorf("Expected bandwidth param %s to be present in %v", param, bw)
				}
			}
		})
	}
}

