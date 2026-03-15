package tools

import (
	"strings"
	"testing"
)

// Advanced tests for disk bandwidth functions to increase coverage

func TestParseStorageFromConfig_Advanced(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedStorage string
		expectedVolume  string
	}{
		{
			name:            "storage with special chars in volume",
			diskConfig:      "storage:vm-100-disk-0",
			expectedStorage: "storage",
			expectedVolume:  "vm-100-disk-0",
		},
		{
			name:            "volume with multiple hyphens",
			diskConfig:      "storage:vm-100--disk-0",
			expectedStorage: "storage",
			expectedVolume:  "vm-100--disk-0",
		},
		{
			name:            "volume with underscores",
			diskConfig:      "storage:vm_100_disk_0",
			expectedStorage: "storage",
			expectedVolume:  "vm_100_disk_0",
		},
		{
			name:            "volume with dots",
			diskConfig:      "storage:vm.100.disk.0",
			expectedStorage: "storage",
			expectedVolume:  "vm.100.disk.0",
		},
		{
			name:            "multiple colons in volume",
			diskConfig:      "storage:vol:ume:with:colons",
			expectedStorage: "storage",
			expectedVolume:  "vol:ume:with:colons",
		},
		{
			name:            "only colon after storage",
			diskConfig:      "storage:",
			expectedStorage: "storage",
			expectedVolume:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, volume := parseStorageFromConfig(tt.diskConfig)
			if storage != tt.expectedStorage {
				t.Errorf("parseStorageFromConfig(%q) storage = %q; want %q",
					tt.diskConfig, storage, tt.expectedStorage)
			}
			if volume != tt.expectedVolume {
				t.Errorf("parseStorageFromConfig(%q) volume = %q; want %q",
					tt.diskConfig, volume, tt.expectedVolume)
			}
		})
	}
}

func TestParseDiskBandwidth_Advanced(t *testing.T) {
	tests := []struct {
		name              string
		diskConfig        string
		expectedParams    map[string]int
	}{
		{
			name:       "all mbps parameters",
			diskConfig: "storage:vol,mbps_rd=10,mbps_rd_max=20,mbps_wr=30,mbps_wr_max=40",
			expectedParams: map[string]int{
				"mbps_rd": 10, "mbps_rd_max": 20, "mbps_wr": 30, "mbps_wr_max": 40,
			},
		},
		{
			name:       "all iops parameters",
			diskConfig: "storage:vol,iops_rd=100,iops_rd_max=200,iops_rd_max_length=5,iops_wr=150,iops_wr_max=250,iops_wr_max_length=10",
			expectedParams: map[string]int{
				"iops_rd": 100, "iops_rd_max": 200, "iops_rd_max_length": 5,
				"iops_wr": 150, "iops_wr_max": 250, "iops_wr_max_length": 10,
			},
		},
		{
			name:       "zero values",
			diskConfig: "storage:vol,mbps_rd=0,iops_wr=0",
			expectedParams: map[string]int{
				"mbps_rd": 0, "iops_wr": 0,
			},
		},
		{
			name:           "empty config",
			diskConfig:     "",
			expectedParams: map[string]int{},
		},
		{
			name:           "config with no bandwidth params",
			diskConfig:     "storage:vol,size=32G",
			expectedParams: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDiskBandwidth(tt.diskConfig)
			if len(result) != len(tt.expectedParams) {
				t.Errorf("parseDiskBandwidth(%q) returned %d params; expected %d",
					tt.diskConfig, len(result), len(tt.expectedParams))
			}
			for key, expectedValue := range tt.expectedParams {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("parseDiskBandwidth(%q) missing key %s", tt.diskConfig, key)
				} else if actualValue != expectedValue {
					t.Errorf("parseDiskBandwidth(%q) %s = %v; expected %v",
						tt.diskConfig, key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestBuildDiskConfigString_Advanced(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedContains []string
	}{
		{
			name:            "add to config with existing params",
			originalConfig:  "storage:vol,cache=writeback",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedContains: []string{"storage:vol", "cache=writeback", "mbps_rd=100"},
		},
		{
			name:            "replace existing bandwidth",
			originalConfig:  "storage:vol,mbps_rd=50,ssd=1",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedContains: []string{"storage:vol", "ssd=1", "mbps_rd=100"},
		},
		{
			name:            "remove old bandwidth add new",
			originalConfig:  "storage:vol,mbps_rd=50,mbps_wr=100",
			bandwidthParams: map[string]interface{}{"iops_rd": 200},
			expectedContains: []string{"storage:vol", "iops_rd=200"},
		},
		{
			name:            "multiple new bandwidth params",
			originalConfig:  "storage:vol",
			bandwidthParams: map[string]interface{}{
				"mbps_rd": 100,
				"mbps_wr": 200,
				"iops_rd": 300,
			},
			expectedContains: []string{"storage:vol", "mbps_rd=100", "mbps_wr=200", "iops_rd=300"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("buildDiskConfigString(%q, %v) = %q; expected to contain %q",
						tt.originalConfig, tt.bandwidthParams, result, expected)
				}
			}
		})
	}
}

func TestRemoveBandwidthParams_Advanced(t *testing.T) {
	tests := []struct {
		name           string
		originalConfig string
		expectedConfig  string
	}{
		{
			name:           "config with all bandwidth params",
			originalConfig: "storage:vol,mbps_rd=100,mbps_rd_max=200,mbps_wr=150,mbps_wr_max=250,iops_rd=500,iops_rd_max=1000,iops_rd_max_length=60,iops_wr=600,iops_wr_max=1200,iops_wr_max_length=60",
			expectedConfig:  "storage:vol",
		},
		{
			name:           "config with no bandwidth",
			originalConfig: "storage:vol,size=32G,cache=writeback,ssd=1",
			expectedConfig:  "storage:vol,size=32G,cache=writeback,ssd=1",
		},
		{
			name:           "config with mixed params",
			originalConfig: "storage:vol,mbps_rd=100,size=32G,ssd=1",
			expectedConfig:  "storage:vol,size=32G,ssd=1",
		},
		{
			name:           "empty config",
			originalConfig: "",
			expectedConfig:  "",
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

func TestIsValidDiskID_AdvancedPatterns(t *testing.T) {
	tests := []struct {
		name     string
		diskID   string
		expected bool
	}{
		{"scsi with leading zeros", "scsi00", true},
		{"scsi with two leading zeros", "scsi000", true},
		{"virtio double digit", "virtio15", true},
		{"sata double digit", "sata99", true},
		{"ide single digit", "ide9", true},
		{"ata single digit", "ata9", true},
		{"scsi triple digit", "scsi100", true},
		{"scsi very large", "scsi9999", true},
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
