package orasfetch

import (
	"os"
	"testing"
)

func TestPullBinary(t *testing.T) {
	ref := "ghcr.io/katasec/dstream-ingester-time:v0.0.1"

	binPath, err := PullBinary(ref)
	if err != nil {
		t.Fatalf("Failed to pull binary: %v", err)
	}

	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("Binary not found at %s: %v", binPath, err)
	}

	t.Logf("✅ Pulled binary to: %s", binPath)
}
