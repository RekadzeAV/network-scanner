package cache

import (
	"testing"
	"time"
)

func BenchmarkDNSCache_Set(b *testing.B) {
	cache := NewDNSCache(5*time.Minute, 10000)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cache.Set("example.com", "93.184.216.34")
	}
}

func BenchmarkDNSCache_Get(b *testing.B) {
	cache := NewDNSCache(5*time.Minute, 10000)
	for i := 0; i < 1000; i++ {
		cache.Set("example.com", "93.184.216.34")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("example.com")
	}
}

func BenchmarkMACVendorCache_Set(b *testing.B) {
	cache := NewMACVendorCache()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cache.Set("AA:BB:CC:DD:EE:FF", "Cisco")
	}
}

func BenchmarkMACVendorCache_Get(b *testing.B) {
	cache := NewMACVendorCache()
	cache.Set("AA:BB:CC:DD:EE:FF", "Cisco")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("AA:BB:CC:DD:EE:FF")
	}
}
