package tools

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
)

// getRequiredIntParam extracts a required integer parameter from the tool request
func getRequiredIntParam(req mcp.CallToolRequest, key string) (int, error) {
	val := req.GetFloat(key, 0)
	if val == 0 {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	return int(val), nil
}

// getOptionalStringParam extracts an optional string parameter from the tool request
func getOptionalStringParam(req mcp.CallToolRequest, key string) string {
	return req.GetString(key, "")
}

// getOptionalBoolParam extracts an optional boolean parameter from the tool request
func getOptionalBoolParam(req mcp.CallToolRequest, key string) bool {
	return req.GetBool(key, false)
}

// mapBoolToInt converts a boolean to int (1 for true, 0 for false)
func mapBoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
