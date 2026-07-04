package cache

import (
	"testing"
	"time"
)

func TestDNSCache_SetAndGet(t *testing.T) {
	cache := NewDNSCache(5*time.Minute, 100)

	cache.Set("example.com", "93.184.216.34")

	ip, ok := cache.Get("example.com")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if ip != "93.184.216.34" {
		t.Fatalf("expected 93.184.216.34, got %s", ip)
	}
}

func TestDNSCache_Expired(t *testing.T) {
	cache := NewDNSCache(100*time.Millisecond, 100)

	cache.Set("example.com", "93.184.216.34")

	time.Sleep(150 * time.Millisecond)

	_, ok := cache.Get("example.com")
	if ok {
		t.Fatal("expected key to be expired")
	}
}

func TestDNSCache_MaxSize(t *testing.T) {
	cache := NewDNSCache(5*time.Minute, 3)

	cache.Set("a.com", "1.1.1.1")
	cache.Set("b.com", "2.2.2.2")
	cache.Set("c.com", "3.3.3.3")

	if cache.Size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.Size())
	}

	// После добавления 4-го элемента кэш должен остаться 3 (любой элемент может быть удалён)
	cache.Set("d.com", "4.4.4.4")
	if cache.Size() != 3 {
		t.Fatalf("expected size 3 after overflow, got %d", cache.Size())
	}
}

func TestDNSCache_Clear(t *testing.T) {
	cache := NewDNSCache(5*time.Minute, 100)

	cache.Set("example.com", "93.184.216.34")
	cache.Clear()

	_, ok := cache.Get("example.com")
	if ok {
		t.Fatal("expected key to be cleared")
	}
}

func TestMACVendorCache_SetAndGet(t *testing.T) {
	cache := NewMACVendorCache()

	cache.Set("AA:BB:CC:DD:EE:FF", "Cisco")

	vendor, ok := cache.Get("AA:BB:CC:DD:EE:FF")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if vendor != "Cisco" {
		t.Fatalf("expected Cisco, got %s", vendor)
	}
}

func TestMACVendorCache_NotFound(t *testing.T) {
	cache := NewMACVendorCache()

	_, ok := cache.Get("AA:BB:CC:DD:EE:FF")
	if ok {
		t.Fatal("expected key not to exist")
	}
}

func TestMACVendorCache_Size(t *testing.T) {
	cache := NewMACVendorCache()

	cache.Set("AA:BB:CC:DD:EE:FF", "Cisco")
	cache.Set("11:22:33:44:55:66", "Dell")

	if cache.Size() != 2 {
		t.Fatalf("expected size 2, got %d", cache.Size())
	}
}
