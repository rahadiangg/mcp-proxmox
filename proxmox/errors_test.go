package proxmox

import (
	"errors"
	"testing"
)

func TestProxmoxError_Error(t *testing.T) {
	originalErr := errors.New("underlying error")
	wrappedErr := &ProxmoxError{
		Operation: "testOperation",
		Err:       originalErr,
	}

	expected := "proxmox testOperation failed: underlying error"
	if result := wrappedErr.Error(); result != expected {
		t.Errorf("Error() = %q; want %q", result, expected)
	}
}

func TestProxmoxError_Unwrap(t *testing.T) {
	originalErr := errors.New("underlying error")
	wrappedErr := &ProxmoxError{
		Operation: "testOperation",
		Err:       originalErr,
	}

	if unwrapped := wrappedErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("Unwrap() = %v; want %v", unwrapped, originalErr)
	}
}

func TestWrapError(t *testing.T) {
	t.Run("wraps non-nil error", func(t *testing.T) {
		originalErr := errors.New("test error")
		wrapped := WrapError("testOp", originalErr)

		if wrapped == nil {
			t.Fatal("WrapError() returned nil")
		}

		proxmoxErr, ok := wrapped.(*ProxmoxError)
		if !ok {
			t.Fatalf("WrapError() returned %T, want *ProxmoxError", wrapped)
		}

		if proxmoxErr.Operation != "testOp" {
			t.Errorf("Operation = %q; want %q", proxmoxErr.Operation, "testOp")
		}

		if proxmoxErr.Err != originalErr {
			t.Errorf("Err = %v; want %v", proxmoxErr.Err, originalErr)
		}
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := WrapError("testOp", nil)
		if wrapped != nil {
			t.Errorf("WrapError(testOp, nil) = %v; want nil", wrapped)
		}
	})

	t.Run("unwrap returns original error", func(t *testing.T) {
		originalErr := errors.New("test error")
		wrapped := WrapError("testOp", originalErr)

		if unwrapped := errors.Unwrap(wrapped); unwrapped != originalErr {
			t.Errorf("errors.Unwrap() = %v; want %v", unwrapped, originalErr)
		}
	})

	t.Run("error chain works with Is", func(t *testing.T) {
		originalErr := errors.New("test error")
		wrapped := WrapError("testOp", originalErr)

		if !errors.Is(wrapped, originalErr) {
			t.Errorf("errors.Is() should return true for original error")
		}
	})
}
