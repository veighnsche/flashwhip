package net

import "testing"

func TestDefaultHTTPClient(t *testing.T) {
	client1 := DefaultHTTPClient()
	client2 := DefaultHTTPClient()

	if client1 == nil {
		t.Fatalf("DefaultHTTPClient returned nil")
	}

	if client1 != client2 {
		t.Errorf("DefaultHTTPClient is not returning a singleton instance")
	}

	if client1.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0 (no hardcoded HTTP deadline; context cancellation is the only guard)", client1.Timeout)
	}
}
