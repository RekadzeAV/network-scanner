package cache

import (
	"sync"
	"time"
)

// DNSCache кэширует результаты DNS-запросов
type DNSCache struct {
	mu      sync.RWMutex
	entries map[string]dnsEntry
	ttl     time.Duration
	maxSize int
}

type dnsEntry struct {
	IP       string
	Hostname string
	expires  time.Time
}

// NewDNSCache создаёт новый DNS кэш
func NewDNSCache(ttl time.Duration, maxSize int) *DNSCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &DNSCache{
		entries: make(map[string]dnsEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get возвращает кэшированное значение
func (c *DNSCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expires) {
		return "", false
	}
	return entry.IP, true
}

// Set сохраняет значение
func (c *DNSCache) Set(key, ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Удаляем старые записи, если достигнут лимит и ключ ещё не существует
	if len(c.entries) >= c.maxSize {
		c.evictExpired()
	}

	// Если всё ещё превышен лимит, удаляем первую попавшуюся
	if len(c.entries) >= c.maxSize {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[key] = dnsEntry{
		IP:      ip,
		expires: time.Now().Add(c.ttl),
	}
}

// evictExpired удаляет истёкшие записи
func (c *DNSCache) evictExpired() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expires) {
			delete(c.entries, key)
		}
	}
}

// Clear очищает весь кэш
func (c *DNSCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]dnsEntry)
}

// Size возвращает количество записей
func (c *DNSCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// MACVendorCache кэширует маппинг MAC -> вендор
type MACVendorCache struct {
	mu      sync.RWMutex
	entries map[string]string
}

// NewMACVendorCache создаёт новый кэш вендоров
func NewMACVendorCache() *MACVendorCache {
	return &MACVendorCache{
		entries: make(map[string]string),
	}
}

// Get возвращает вендора по MAC
func (c *MACVendorCache) Get(mac string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vendor, ok := c.entries[mac]
	return vendor, ok
}

// Set сохраняет вендора
func (c *MACVendorCache) Set(mac, vendor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[mac] = vendor
}

// Size возвращает количество записей
func (c *MACVendorCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
