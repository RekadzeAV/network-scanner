package builder

import (
	"testing"
)

func TestNewContainer(t *testing.T) {
	cfg := Config{
		LogLevel: "debug",
		DBPath:   "",
	}
	c := NewContainer(cfg)
	if c == nil {
		t.Fatal("expected non-nil Container")
	}
}

func TestContainer_GetScanner(t *testing.T) {
	c := NewContainer(Config{LogLevel: "info"})
	if c.GetScanner() == nil {
		t.Fatal("expected non-nil scanner service")
	}
}

func TestContainer_GetTopology(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetTopology() == nil {
		t.Fatal("expected non-nil topology service")
	}
}

func TestContainer_GetSecurity(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetSecurity() == nil {
		t.Fatal("expected non-nil security service")
	}
}

func TestContainer_GetRemoteExec(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetRemoteExec() == nil {
		t.Fatal("expected non-nil remote exec service")
	}
}

func TestContainer_GetInventory(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetInventory() == nil {
		t.Fatal("expected non-nil inventory service")
	}
}
