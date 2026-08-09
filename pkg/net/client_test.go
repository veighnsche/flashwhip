package net

import (
	"testing"
	"time"
)

func TestDefaultHTTPClient(t *testing.T) {
	client1 := DefaultHTTPClient()
	client2 := DefaultHTTPClient()

	if client1 == nil {
		t.Fatalf("DefaultHTTPClient returned nil")
	}

	if client1 != client2 {
		t.Errorf("DefaultHTTPClient is not returning a singleton instance")
	}

	if client1.Timeout != 15*time.Second {
		t.Errorf("Timeout = %v, want 15s", client1.Timeout)
	}
}
