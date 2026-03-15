package tools

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsValidDiskID(t *testing.T) {
	tests := []struct {
		name     string
		diskID   string
		expected bool
	}{
		{"valid scsi disk", "scsi0", true},
		{"valid virtio disk", "virtio1", true},
		{"valid sata disk", "sata2", true},
		{"valid ide disk", "ide3", true},
		{"valid ata disk", "ata4", true},
		{"valid scsi high number", "scsi15", true},
		{"invalid - missing number", "scsi", false},
		{"invalid - wrong prefix", "nvme0", false},
		{"invalid - empty", "", false},
		{"invalid - special chars", "scsi-0", false},
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

func TestIsDiskKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"valid scsi key", "scsi0", true},
		{"valid virtio key", "virtio1", true},
		{"valid sata key", "sata2", true},
		{"invalid - missing number", "scsi", false},
		{"invalid - wrong prefix", "net0", false},
		{"invalid empty", "", false},
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

func TestParseStorageFromConfig(t *testing.T) {
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
	}{
		{
			name:       "basic storage and volume",
			diskConfig: "local-lvm:vm-100-disk-0",
			storage:    "local-lvm",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "storage with additional params",
			diskConfig: "local:vm-100-disk-0,size=32G",
			storage:    "local",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "storage only",
			diskConfig: "local-lvm",
			storage:    "local-lvm",
			volumePath: "",
		},
		{
			name:       "empty config",
			diskConfig: "",
			storage:    "",
			volumePath: "",
		},
		{
			name:       "storage with colon in path",
			diskConfig: "storage:path:with:colons",
			storage:    "storage",
			volumePath: "path:with:colons",
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

func TestParseDiskBandwidth(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedParams  map[string]interface{}
	}{
		{
			name:       "no bandwidth params",
			diskConfig: "local-lvm:vm-100-disk-0",
			expectedParams: map[string]interface{}{},
		},
		{
			name:       "single mbps read limit",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
			},
		},
		{
			name:       "multiple bandwidth params",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_wr=200,iops_rd=500",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
				"mbps_wr": 200,
				"iops_rd": 500,
			},
		},
		{
			name:       "all bandwidth params",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_rd_max=150,mbps_wr=200,mbps_wr_max=250,iops_rd=500,iops_rd_max=600,iops_rd_max_length=60,iops_wr=800,iops_wr_max=900,iops_wr_max_length=120",
			expectedParams: map[string]interface{}{
				"mbps_rd":            100,
				"mbps_rd_max":        150,
				"mbps_wr":            200,
				"mbps_wr_max":        250,
				"iops_rd":            500,
				"iops_rd_max":        600,
				"iops_rd_max_length": 60,
				"iops_wr":            800,
				"iops_wr_max":        900,
				"iops_wr_max_length": 120,
			},
		},
		{
			name:       "mixed with non-bandwidth params",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,size=32G",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
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

func TestBuildDiskConfigString(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedConfig   string
	}{
		{
			name:            "add bandwidth to simple config",
			originalConfig:  "local-lvm:vm-100-disk-0",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig:  "local-lvm:vm-100-disk-0,mbps_rd=100",
		},
		{
			name:           "preserve existing params",
			originalConfig: "local-lvm:vm-100-disk-0,size=32G,ssd=1",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig:  "local-lvm:vm-100-disk-0,size=32G,ssd=1,mbps_rd=100",
		},
		{
			name:            "empty bandwidth params",
			originalConfig:  "local-lvm:vm-100-disk-0,size=32G",
			bandwidthParams: map[string]interface{}{},
			expectedConfig:  "local-lvm:vm-100-disk-0,size=32G",
		},
		{
			name:           "replace existing bandwidth params",
			originalConfig: "local-lvm:vm-100-disk-0,mbps_rd=50,size=32G",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig:  "local-lvm:vm-100-disk-0,size=32G,mbps_rd=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			// For tests with multiple parameters, order may vary - use map comparison
			if strings.Contains(tt.expectedConfig, ",") && strings.Contains(result, ",") {
				// Split both results into parts and compare as sets
				resultParts := strings.Split(result, ",")
				expectedParts := strings.Split(tt.expectedConfig, ",")
				resultMap := make(map[string]bool)
				expectedMap := make(map[string]bool)
				for _, p := range resultParts {
					resultMap[p] = true
				}
				for _, p := range expectedParts {
					expectedMap[p] = true
				}
				// Check same number of parts
				if len(resultMap) != len(expectedMap) {
					t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q (different number of parts: %d vs %d)",
						tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig, len(resultMap), len(expectedMap))
					return
				}
				// Check all parts present
				for part := range expectedMap {
					if !resultMap[part] {
						t.Errorf("buildDiskConfigString(%q, %v) = %q; missing expected part %q",
							tt.originalConfig, tt.bandwidthParams, result, part)
					}
				}
				for part := range resultMap {
					if !expectedMap[part] {
						t.Errorf("buildDiskConfigString(%q, %v) = %q; unexpected part %q",
							tt.originalConfig, tt.bandwidthParams, result, part)
					}
				}
			} else if result != tt.expectedConfig {
				t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
					tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
			}
		})
	}
}

func TestRemoveBandwidthParams(t *testing.T) {
	tests := []struct {
		name           string
		originalConfig string
		expectedConfig string
	}{
		{
			name:           "simple config no params",
			originalConfig: "local-lvm:vm-100-disk-0",
			expectedConfig: "local-lvm:vm-100-disk-0",
		},
		{
			name:           "remove bandwidth params",
			originalConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_wr=200",
			expectedConfig: "local-lvm:vm-100-disk-0",
		},
		{
			name:           "keep non-bandwidth params",
			originalConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,size=32G,ssd=1",
			expectedConfig: "local-lvm:vm-100-disk-0,size=32G,ssd=1",
		},
		{
			name:           "all bandwidth params",
			originalConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_rd_max=150,mbps_wr=200,mbps_wr_max=250,iops_rd=500,iops_rd_max=600,iops_rd_max_length=60,iops_wr=800,iops_wr_max=900,iops_wr_max_length=120",
			expectedConfig: "local-lvm:vm-100-disk-0",
		},
		{
			name:           "mixed bandwidth and other params",
			originalConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,size=32G,iops_wr=500,ssd=1",
			expectedConfig: "local-lvm:vm-100-disk-0,size=32G,ssd=1",
		},
		{
			name:           "empty config",
			originalConfig: "",
			expectedConfig: "",
		},
		{
			name:           "only bandwidth params",
			originalConfig: "storage:volume,mbps_rd=100,iops_wr=500",
			expectedConfig: "storage:volume",
		},
		{
			name:           "params with spaces",
			originalConfig: "local-lvm:vm-100-disk-0, mbps_rd=100 , size=32G",
			expectedConfig: "local-lvm:vm-100-disk-0,size=32G",
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

func TestBuildDiskConfigString_MoreEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedConfig   string
	}{
		{
			name:            "empty config with bandwidth",
			originalConfig:  "",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig:  ",mbps_rd=100",
		},
		{
			name:           "config with empty values",
			originalConfig: "storage:vol,size=32G",
			bandwidthParams: map[string]interface{}{
				"mbps_rd": 0,
				"iops_rd": 0,
			},
			expectedConfig: "storage:vol,size=32G,mbps_rd=0,iops_rd=0",
		},
		{
			name:           "existing bandwidth replaced",
			originalConfig: "storage:vol,mbps_rd=50,iops_rd=100",
			bandwidthParams: map[string]interface{}{
				"mbps_rd": 100,
				"iops_rd": 200,
			},
			expectedConfig: "storage:vol,mbps_rd=100,iops_rd=200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			// For non-empty configs, check that all expected parts are present (order may vary due to map iteration)
			if tt.originalConfig != "" && result != tt.expectedConfig {
				expectedParts := strings.Split(tt.expectedConfig, ",")
				resultParts := strings.Split(result, ",")
				if len(expectedParts) != len(resultParts) {
					t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
						tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
					return
				}
				// Create maps to compare (order-insensitive)
				expectedMap := make(map[string]bool)
				resultMap := make(map[string]bool)
				for _, p := range expectedParts {
					expectedMap[p] = true
				}
				for _, p := range resultParts {
					resultMap[p] = true
				}
				for k := range expectedMap {
					if !resultMap[k] {
						t.Errorf("buildDiskConfigString(%q, %v) = %q; missing part %q",
							tt.originalConfig, tt.bandwidthParams, result, k)
					}
				}
				for k := range resultMap {
					if !expectedMap[k] {
						t.Errorf("buildDiskConfigString(%q, %v) = %q; unexpected part %q",
							tt.originalConfig, tt.bandwidthParams, result, k)
					}
				}
			} else if tt.originalConfig == "" && result != tt.expectedConfig {
				t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
					tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
			}
		})
	}
}

func TestParseDiskBandwidth_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedParams  map[string]interface{}
	}{
		{
			name:           "empty config",
			diskConfig:     "",
			expectedParams: map[string]interface{}{},
		},
		{
			name:           "only bandwidth params",
			diskConfig:     "mbps_rd=100,mbps_wr=200",
			expectedParams: map[string]interface{}{"mbps_rd": 100, "mbps_wr": 200},
		},
		{
			name:       "invalid numeric values ignored",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=invalid,mbps_wr=200",
			expectedParams: map[string]interface{}{
				"mbps_wr": 200,
			},
		},
		{
			name:           "duplicate params keeps last",
			diskConfig:     "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_rd=200",
			expectedParams: map[string]interface{}{"mbps_rd": 200},
		},
		{
			name:       "partial param names ignored",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,mbps=50",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
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

func TestParseStorageFromConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		diskConfig string
		storage    string
		volumePath string
	}{
		{
			name:       "multiple colons",
			diskConfig: "storage:path:with:many:colons",
			storage:    "storage",
			volumePath: "path:with:many:colons",
		},
		{
			name:       "colon at end",
			diskConfig: "storage:",
			storage:    "storage",
			volumePath: "",
		},
		{
			name:       "colon at start",
			diskConfig: ":volume",
			storage:    "",
			volumePath: "volume",
		},
		{
			name:       "only colon",
			diskConfig: ":",
			storage:    "",
			volumePath: "",
		},
		{
			name:       "no colon only storage",
			diskConfig: "storage-name",
			storage:    "storage-name",
			volumePath: "",
		},
		{
			name:       "storage with dash and underscore",
			diskConfig: "local-lvm_storage:vm-100-disk-0",
			storage:    "local-lvm_storage",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "storage with numbers",
			diskConfig: "storage123:vm-100-disk-0",
			storage:    "storage123",
			volumePath: "vm-100-disk-0",
		},
		{
			name:       "volume with special chars",
			diskConfig: "storage:vm-100-disk-0.raw",
			storage:    "storage",
			volumePath: "vm-100-disk-0.raw",
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

func TestDiskBandwidthParsing_ComplexCases(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedParams  map[string]interface{}
	}{
		{
			name:           "bandwidth with zero values",
			diskConfig:     "local-lvm:vm-100-disk-0,mbps_rd=0,iops_rd=0",
			expectedParams: map[string]interface{}{"mbps_rd": 0, "iops_rd": 0},
		},
		{
			name:       "mixed bandwidth with other params",
			diskConfig: "local-lvm:vm-100-disk-0,mbps_rd=100,size=32G,cache=writeback",
			expectedParams: map[string]interface{}{
				"mbps_rd": 100,
			},
		},
		{
			name:           "all iops params",
			diskConfig:     "local-lvm:vm-100-disk-0,iops_rd=500,iops_rd_max=600,iops_rd_max_length=60",
			expectedParams: map[string]interface{}{"iops_rd": 500, "iops_rd_max": 600, "iops_rd_max_length": 60},
		},
		{
			name:           "all mbps params",
			diskConfig:     "local-lvm:vm-100-disk-0,mbps_rd=100,mbps_rd_max=150,mbps_wr=200,mbps_wr_max=250",
			expectedParams: map[string]interface{}{"mbps_rd": 100, "mbps_rd_max": 150, "mbps_wr": 200, "mbps_wr_max": 250},
		},
		{
			name:       "all write iops params",
			diskConfig: "local-lvm:vm-100-disk-0,iops_wr=800,iops_wr_max=900,iops_wr_max_length=120",
			expectedParams: map[string]interface{}{"iops_wr": 800, "iops_wr_max": 900, "iops_wr_max_length": 120},
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

func TestRemoveBandwidthParams_MoreCases(t *testing.T) {
	tests := []struct {
		name           string
		originalConfig string
		expectedConfig string
	}{
		{
			name:           "config with only bandwidth",
			originalConfig: "storage:vol,mbps_rd=100,mbps_wr=200",
			expectedConfig: "storage:vol",
		},
		{
			name:           "config with bandwidth and cache",
			originalConfig: "storage:vol,mbps_rd=100,cache=writeback",
			expectedConfig: "storage:vol,cache=writeback",
		},
		{
			name:           "config with all params mixed",
			originalConfig: "storage:vol,mbps_rd=100,size=32G,iops_rd=500,ssd=1",
			expectedConfig: "storage:vol,size=32G,ssd=1",
		},
		{
			name:           "config without bandwidth",
			originalConfig: "storage:vol,size=32G,ssd=1",
			expectedConfig: "storage:vol,size=32G,ssd=1",
		},
		{
			name:           "config with duplicate bandwidth param",
			originalConfig: "storage:vol,mbps_rd=100,mbps_rd=200",
			expectedConfig: "storage:vol",
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

func TestBuildDiskConfigString_MoreCases(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedConfig   string
	}{
		{
			name:           "add to config with cache",
			originalConfig: "storage:vol,cache=writeback",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig: "storage:vol,cache=writeback,mbps_rd=100",
		},
		{
			name:            "empty bandwidth params",
			originalConfig:  "storage:vol,size=32G",
			bandwidthParams: map[string]interface{}{},
			expectedConfig:  "storage:vol,size=32G",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			if result != tt.expectedConfig {
				t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
					tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
			}
		})
	}
}

func TestBuildDiskConfigString_SingleBandwidthParam(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedConfig   string
	}{
		{
			name:           "add mbps_rd only",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedConfig: "storage:vol,mbps_rd=100",
		},
		{
			name:           "add mbps_wr only",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"mbps_wr": 200},
			expectedConfig: "storage:vol,mbps_wr=200",
		},
		{
			name:           "add iops_rd only",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"iops_rd": 500},
			expectedConfig: "storage:vol,iops_rd=500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			if result != tt.expectedConfig {
				t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
					tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
			}
		})
	}
}

func TestBuildDiskConfigString_ExtremeValues(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedConfig   string
	}{
		{
			name:           "very large mbps value",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"mbps_rd": 999999},
			expectedConfig: "storage:vol,mbps_rd=999999",
		},
		{
			name:           "zero bandwidth values",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"mbps_rd": 0, "mbps_wr": 0},
			expectedConfig: "storage:vol,mbps_rd=0,mbps_wr=0",
		},
		{
			name:           "max burst duration",
			originalConfig: "storage:vol",
			bandwidthParams: map[string]interface{}{"iops_rd_max_length": 1000},
			expectedConfig: "storage:vol,iops_rd_max_length=1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDiskConfigString(tt.originalConfig, tt.bandwidthParams)
			// For zero bandwidth values, order may vary
			if tt.name == "zero bandwidth values" {
				// Check both possible orderings
				if result != "storage:vol,mbps_rd=0,mbps_wr=0" && result != "storage:vol,mbps_wr=0,mbps_rd=0" {
					t.Errorf("buildDiskConfigString(%q, %v) = %q; want one of [storage:vol,mbps_rd=0,mbps_wr=0, storage:vol,mbps_wr=0,mbps_rd=0]",
						tt.originalConfig, tt.bandwidthParams, result)
				}
			} else {
				if result != tt.expectedConfig {
					t.Errorf("buildDiskConfigString(%q, %v) = %q; want %q",
						tt.originalConfig, tt.bandwidthParams, result, tt.expectedConfig)
				}
			}
		})
	}
}

func TestParseStorageFromConfig_MoreEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedStorage string
		expectedVolume  string
	}{
		{
			name:            "multiple colons - use first",
			diskConfig:      "storage:vol:ume:path",
			expectedStorage: "storage",
			expectedVolume:  "vol:ume:path",
		},
		{
			name:            "trailing comma without additional params",
			diskConfig:      "storage:vol,",
			expectedStorage: "storage",
			expectedVolume:  "vol",
		},
		{
			name:            "only storage no volume",
			diskConfig:      "storage",
			expectedStorage: "storage",
			expectedVolume:  "",
		},
		{
			name:            "empty string",
			diskConfig:      "",
			expectedStorage: "",
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

func TestParseDiskBandwidth_InvalidValues(t *testing.T) {
	tests := []struct {
		name           string
		diskConfig     string
		expectedParams int
	}{
		{
			name:           "non-numeric bandwidth value ignored",
			diskConfig:     "storage:vol,mbps_rd=invalid",
			expectedParams: 0,
		},
		{
			name:           "float value ignored",
			diskConfig:     "storage:vol,mbps_rd=10.5",
			expectedParams: 0,
		},
		{
			name:           "negative value parsed successfully",
			diskConfig:     "storage:vol,mbps_rd=-100",
			expectedParams: 1,
		},
		{
			name:           "mixed valid and invalid - only valid counted",
			diskConfig:     "storage:vol,mbps_rd=100,mbps_wr=invalid,iops_rd=500",
			expectedParams: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDiskBandwidth(tt.diskConfig)
			if len(result) != tt.expectedParams {
				t.Errorf("parseDiskBandwidth(%q) returned %d params; expected %d",
					tt.diskConfig, len(result), tt.expectedParams)
			}
		})
	}
}

func TestIsValidDiskID_NumbersBeyondRange(t *testing.T) {
	tests := []struct {
		name     string
		diskID   string
		expected bool
	}{
		{"three digit number", "scsi100", true},
		{"very large number", "virtio9999", true},
		{"number in middle", "sata12", true},
		{"double digit", "ide99", true},
		{"zero prefix", "scsi01", true},
		{"multiple zeros", "ata001", true},
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

func TestParseStorageFromConfig_TrickyInputs(t *testing.T) {
	tests := []struct {
		name            string
		diskConfig      string
		expectedStorage string
		expectedVolume  string
	}{
		{
			name:            "only colon",
			diskConfig:      ":",
			expectedStorage: "",
			expectedVolume:  "",
		},
		{
			name:            "colon at end",
			diskConfig:      "storage:",
			expectedStorage: "storage",
			expectedVolume:  "",
		},
		{
			name:            "comma at start",
			diskConfig:      ",storage:vol",
			expectedStorage: "",
			expectedVolume:  "",
		},
		{
			name:            "storage with multiple commas",
			diskConfig:      "storage:vol,size=32G,ssd=1",
			expectedStorage: "storage",
			expectedVolume:  "vol",
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

func TestBuildDiskConfigString_AdvancedEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		originalConfig   string
		bandwidthParams  map[string]interface{}
		expectedContains []string // check that result contains these parts
	}{
		{
			name:            "empty config with bandwidth",
			originalConfig:  "",
			bandwidthParams: map[string]interface{}{"mbps_rd": 100},
			expectedContains: []string{"mbps_rd=100"},
		},
		{
			name:            "config with spaces around commas",
			originalConfig:  "storage:vol , size=32G , ssd=1",
			bandwidthParams: map[string]interface{}{"mbps_wr": 200},
			expectedContains: []string{"mbps_wr=200", "size=32G", "ssd=1"},
		},
		{
			name:            "all bandwidth params",
			originalConfig:  "storage:vol",
			bandwidthParams: map[string]interface{}{
				"mbps_rd": 100,
				"mbps_rd_max": 200,
				"mbps_wr": 150,
				"mbps_wr_max": 250,
				"iops_rd": 500,
				"iops_rd_max": 1000,
				"iops_rd_max_length": 60,
				"iops_wr": 600,
				"iops_wr_max": 1200,
				"iops_wr_max_length": 60,
			},
			expectedContains: []string{
				"mbps_rd=100", "mbps_rd_max=200", "mbps_wr=150", "mbps_wr_max=250",
				"iops_rd=500", "iops_rd_max=1000", "iops_rd_max_length=60",
				"iops_wr=600", "iops_wr_max=1200", "iops_wr_max_length=60",
			},
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

func TestRemoveBandwidthParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		originalConfig string
		expectedConfig string
	}{
		{
			name:           "empty string",
			originalConfig: "",
			expectedConfig: "",
		},
		{
			name:           "only bandwidth params",
			originalConfig: "storage:vol,mbps_rd=100,mbps_wr=200",
			expectedConfig: "storage:vol",
		},
		{
			name:           "params without equals",
			originalConfig: "storage:vol,ssd,readonly",
			expectedConfig: "storage:vol,ssd,readonly",
		},
		{
			name:           "mixed bandwidth and non-bandwidth",
			originalConfig: "storage:vol,mbps_rd=100,ssd=1,iops_wr=500",
			expectedConfig: "storage:vol,ssd=1",
		},
		{
			name:           "params with extra commas",
			originalConfig: "storage:vol,size=32G,,ssd=1",
			expectedConfig: "storage:vol,size=32G,ssd=1",
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
