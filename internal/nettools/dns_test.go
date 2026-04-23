package nettools

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestNormalizeDNSErrorTimeout(t *testing.T) {
	err := normalizeDNSError(context.DeadlineExceeded)
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("expected ToolError, got %T", err)
	}
	if te.Code != ToolErrorTimeout {
		t.Fatalf("expected %q, got %q", ToolErrorTimeout, te.Code)
	}
}

func TestNormalizeDNSErrorNotFound(t *testing.T) {
	err := normalizeDNSError(&net.DNSError{IsNotFound: true})
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("expected ToolError, got %T", err)
	}
	if te.Code != ToolErrorNetwork {
		t.Fatalf("expected %q, got %q", ToolErrorNetwork, te.Code)
	}
}
