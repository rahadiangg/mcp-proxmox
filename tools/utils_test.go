package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMapBoolToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected int
	}{
		{"true returns 1", true, 1},
		{"false returns 0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapBoolToInt(tt.input)
			if result != tt.expected {
				t.Errorf("mapBoolToInt(%v) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetOptionalStringParam(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"existing": "value",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing key", "existing", "value"},
		{"missing key", "missing", ""},
		{"empty string key", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOptionalStringParam(req, tt.key)
			if result != tt.expected {
				t.Errorf("getOptionalStringParam(req, %q) = %q; want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetOptionalBoolParam(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"trueParam":  true,
				"falseParam": false,
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"true parameter", "trueParam", true},
		{"false parameter", "falseParam", false},
		{"missing parameter", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOptionalBoolParam(req, tt.key)
			if result != tt.expected {
				t.Errorf("getOptionalBoolParam(req, %q) = %v; want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetRequiredIntParam(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"existing": float64(42),
				"zero":     float64(0),
			},
		},
	}

	tests := []struct {
		name        string
		key         string
		expected    int
		expectError bool
	}{
		{"existing key returns value", "existing", 42, false},
		{"missing key returns error", "missing", 0, true},
		{"zero value returns error", "zero", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getRequiredIntParam(req, tt.key)
			if tt.expectError {
				if err == nil {
					t.Errorf("getRequiredIntParam(req, %q) expected error, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("getRequiredIntParam(req, %q) unexpected error: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("getRequiredIntParam(req, %q) = %d; want %d", tt.key, result, tt.expected)
				}
			}
		})
	}
}

func TestGetRequiredIntParam_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		key         string
		expected    int
		expectError bool
	}{
		{
			name:        "negative value",
			arguments:   map[string]interface{}{"value": float64(-1)},
			key:         "value",
			expected:    -1,
			expectError: false,
		},
		{
			name:        "large value",
			arguments:   map[string]interface{}{"value": float64(999999)},
			key:         "value",
			expected:    999999,
			expectError: false,
		},
		{
			name:        "int instead of float",
			arguments:   map[string]interface{}{"value": 42},
			key:         "value",
			expected:    42,
			expectError: false,
		},
		{
			name:        "string number",
			arguments:   map[string]interface{}{"value": "42"},
			key:         "value",
			expected:    42,
			expectError: false,
		},
		{
			name:        "nil arguments",
			arguments:   nil,
			key:         "value",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty arguments map",
			arguments:   map[string]interface{}{},
			key:         "value",
			expected:    0,
			expectError: true,
		},
		{
			name:        "float with decimal",
			arguments:   map[string]interface{}{"value": float64(42.7)},
			key:         "value",
			expected:    42,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}
			result, err := getRequiredIntParam(req, tt.key)
			if tt.expectError {
				if err == nil {
					t.Errorf("getRequiredIntParam(req, %q) expected error, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("getRequiredIntParam(req, %q) unexpected error: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("getRequiredIntParam(req, %q) = %d; want %d", tt.key, result, tt.expected)
				}
			}
		})
	}
}

func TestGetOptionalStringParam_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		arguments map[string]interface{}
		key      string
		expected string
	}{
		{
			name:      "nil arguments",
			arguments: nil,
			key:       "value",
			expected:  "",
		},
		{
			name:      "empty arguments map",
			arguments: map[string]interface{}{},
			key:       "value",
			expected:  "",
		},
		{
			name:      "numeric value not converted to string",
			arguments: map[string]interface{}{"value": float64(42)},
			key:       "value",
			expected:  "",
		},
		{
			name:      "boolean value not converted to string",
			arguments: map[string]interface{}{"value": true},
			key:       "value",
			expected:  "",
		},
		{
			name:      "empty string value",
			arguments: map[string]interface{}{"value": ""},
			key:       "value",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}
			result := getOptionalStringParam(req, tt.key)
			if result != tt.expected {
				t.Errorf("getOptionalStringParam(req, %q) = %q; want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetOptionalBoolParam_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		arguments map[string]interface{}
		key       string
		expected  bool
	}{
		{
			name:      "nil arguments",
			arguments: nil,
			key:       "value",
			expected:  false,
		},
		{
			name:      "empty arguments map",
			arguments: map[string]interface{}{},
			key:       "value",
			expected:  false,
		},
		{
			name:      "numeric 1 returns true",
			arguments: map[string]interface{}{"value": float64(1)},
			key:       "value",
			expected:  true,
		},
		{
			name:      "numeric 0 returns false",
			arguments: map[string]interface{}{"value": float64(0)},
			key:       "value",
			expected:  false,
		},
		{
			name:      "string true returns true",
			arguments: map[string]interface{}{"value": "true"},
			key:       "value",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}
			result := getOptionalBoolParam(req, tt.key)
			if result != tt.expected {
				t.Errorf("getOptionalBoolParam(req, %q) = %v; want %v", tt.key, result, tt.expected)
			}
		})
	}
}
