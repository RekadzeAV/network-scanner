package network

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkNewARPCache
func BenchmarkNewARPCache(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewARPCache(5*time.Minute, refreshFunc)
	}
}

// BenchmarkARPCacheGet
func BenchmarkARPCacheGet(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1":  "aa:bb:cc:dd:ee:ff",
			"192.168.1.2":  "11:22:33:44:55:66",
			"192.168.1.3":  "aa:bb:cc:dd:ee:00",
			"192.168.1.10": "bb:cc:dd:ee:ff:00",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)
	cache.Refresh()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("192.168.1.1")
	}
}

// BenchmarkARPCacheGetBatch
func BenchmarkARPCacheGetBatch(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		entries := make(map[string]string)
		for i := 1; i <= 100; i++ {
			entries[fmt.Sprintf("192.168.1.%d", i)] = "aa:bb:cc:dd:ee:ff"
		}
		return entries, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)
	cache.Refresh()

	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d", i+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.GetBatch(ips)
	}
}

// BenchmarkARPCacheRefresh
func BenchmarkARPCacheRefresh(b *testing.B) {
	callCount := 0
	refreshFunc := func() (map[string]string, error) {
		callCount++
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Refresh()
	}
}

// BenchmarkARPCacheGetAll
func BenchmarkARPCacheGetAll(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		entries := make(map[string]string)
		for i := 1; i <= 100; i++ {
			entries[fmt.Sprintf("192.168.1.%d", i)] = "aa:bb:cc:dd:ee:ff"
		}
		return entries, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)
	cache.Refresh()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.GetAll()
	}
}

// BenchmarkParseWindowsARP
func BenchmarkParseWindowsARP(b *testing.B) {
	output := `
Интерфейс: 192.168.1.1 --- 0xb
  Address          Type         Interface
  192.168.1.1      0a-1b-2c-3d-4e-5f   dynamic      192.168.1.1
  192.168.1.10     11-22-33-44-55-66   dynamic      192.168.1.1
  192.168.1.20     aa:bb:cc:dd:ee:ff   static       192.168.1.1
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseWindowsARP(output)
	}
}

// BenchmarkParseLinuxARP
func BenchmarkParseLinuxARP(b *testing.B) {
	output := `
192.168.1.1 dev eth0 lladdr 0a:1b:2c:3d:4e:5f REACHABLE
192.168.1.10 dev eth0 lladdr 11:22:33:44:55:66 STALE
192.168.1.20 dev eth0 lladdr aa:bb:cc:dd:ee:ff COMPLETE
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseLinuxARP(output)
	}
}

// BenchmarkARPCacheConcurrentGet
func BenchmarkARPCacheConcurrentGet(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)
	cache.Refresh()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("192.168.1.1")
		}
	})
}

// BenchmarkResolveMACBatch
func BenchmarkResolveMACBatch(b *testing.B) {
	refreshFunc := func() (map[string]string, error) {
		entries := make(map[string]string)
		for i := 1; i <= 100; i++ {
			entries[fmt.Sprintf("192.168.1.%d", i)] = "aa:bb:cc:dd:ee:ff"
		}
		return entries, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)
	cache.Refresh()

	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d", i+1)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ResolveMACBatch(ctx, ips, cache)
	}
}
