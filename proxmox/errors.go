package proxmox

import (
	"fmt"
)

// ProxmoxError wraps errors from Proxmox API calls
type ProxmoxError struct {
	Operation string
	Err       error
}

func (e *ProxmoxError) Error() string {
	return fmt.Sprintf("proxmox %s failed: %v", e.Operation, e.Err)
}

func (e *ProxmoxError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with operation context
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &ProxmoxError{Operation: operation, Err: err}
}
