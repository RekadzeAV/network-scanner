//go:build integration

package nettools

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIntegrationPingLocalhost(t *testing.T) {
	t.Parallel()
	res, err := RunPingStructured(context.Background(), "127.0.0.1", 1, 10*time.Second)
	if err != nil {
		var te *ToolError
		if errors.As(err, &te) && (te.Code == ToolErrorNotInstalled || te.Code == ToolErrorPermissionDenied) {
			t.Skipf("integration env is not ready for ping: %v", err)
		}
		t.Fatalf("RunPingStructured failed: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil ping result")
	}
	if res.Stats.Sent <= 0 {
		t.Fatalf("expected sent packets > 0, got %d", res.Stats.Sent)
	}
}

func TestIntegrationDNSLocalhost(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := LookupDNSWithResolver(ctx, "localhost", "")
	if err != nil {
		var te *ToolError
		if errors.As(err, &te) && te.Code == ToolErrorTimeout {
			t.Skipf("integration env DNS timeout: %v", err)
		}
		t.Fatalf("LookupDNSWithResolver failed: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil dns result")
	}
}
