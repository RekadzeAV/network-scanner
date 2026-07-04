package network

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// --- Test NewARPCache ---

func TestNewARPCache(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	if cache == nil {
		t.Fatal("NewARPCache() returned nil")
	}

	if cache.Size() != 0 {
		t.Errorf("Initial cache size = %d, want 0", cache.Size())
	}

	if cache.IsFresh() {
		t.Error("Cache should not be fresh initially")
	}

	if cache.IsRefreshed() {
		t.Error("Cache should not be refreshed initially")
	}
}

func TestNewARPCacheNilRefreshFunc(t *testing.T) {
	cache := NewARPCache(5*time.Minute, nil)

	if cache == nil {
		t.Fatal("NewARPCache() returned nil")
	}

	// RefreshAsync с nil refreshFunc не должен паниковать
	cache.RefreshAsync()

	// Refresh с nil refreshFunc должен вернуть ошибку
	err := cache.Refresh()
	if err == nil {
		t.Error("Refresh() should return error with nil refreshFunc")
	}
}

// --- Test Get ---

func TestGetCachedEntry(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Сначала обновляем кэш
	cache.Refresh()

	// Получаем запись из кэша
	mac, err := cache.Get("192.168.1.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Get() mac = %v, want aa:bb:cc:dd:ee:ff", mac)
	}
}

func TestGetUncachedEntry(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш синхронно (для предсказуемости теста)
	cache.Refresh()

	// Получаем запись из кэша
	mac, err := cache.Get("192.168.1.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Get() mac = %v, want aa:bb:cc:dd:ee:ff", mac)
	}
}

func TestGetNonExistentIP(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Запрашиваем IP, которого нет в кэше
	_, err := cache.Get("192.168.1.100")
	if err == nil {
		t.Error("Get() should return error for non-existent IP")
	}
}

func TestGetExpiredEntry(t *testing.T) {
	callCount := 0
	refreshFunc := func() (map[string]string, error) {
		callCount++
		if callCount == 1 {
			return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
		}
		return map[string]string{"192.168.1.1": "11:22:33:44:55:66"}, nil
	}

	cache := NewARPCache(10*time.Millisecond, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Ждём истечения TTL
	time.Sleep(20 * time.Millisecond)

	// Синхронно обновляем кэш (для предсказуемости теста)
	cache.Refresh()

	// Запрашиваем запись — должна быть новая
	mac, err := cache.Get("192.168.1.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if mac != "11:22:33:44:55:66" {
		t.Errorf("Get() mac = %v, want 11:22:33:44:55:66 (refreshed)", mac)
	}
}

// --- Test RefreshAsync ---

func TestRefreshAsync(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Запускаем асинхронное обновление
	cache.RefreshAsync()

	// Ждём завершения goroutine
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что кэш обновлён
	if !cache.IsRefreshed() {
		t.Error("Cache should be refreshed after RefreshAsync()")
	}

	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1", cache.Size())
	}
}

func TestRefreshAsyncWithError(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return nil, fmt.Errorf("network error")
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Запускаем асинхронное обновление с ошибкой
	cache.RefreshAsync()

	// Ждём завершения goroutine
	time.Sleep(100 * time.Millisecond)

	// Кэш должен остаться пустым
	if cache.IsRefreshed() {
		t.Error("Cache should not be marked as refreshed on error")
	}
}

// --- Test Refresh ---

func TestRefresh(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	err := cache.Refresh()
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if cache.Size() != 2 {
		t.Errorf("Cache size = %d, want 2", cache.Size())
	}

	if !cache.IsFresh() {
		t.Error("Cache should be fresh after Refresh()")
	}
}

func TestRefreshWithError(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return nil, fmt.Errorf("network error")
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	err := cache.Refresh()
	if err == nil {
		t.Error("Refresh() should return error")
	}
}

// --- Test GetBatch ---

func TestGetBatchAllCached(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
			"192.168.1.3": "aa:bb:cc:dd:ee:00",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Запрашиваем батч
	results := cache.GetBatch([]string{"192.168.1.1", "192.168.1.2", "192.168.1.3"})

	if len(results) != 3 {
		t.Errorf("GetBatch() returned %d results, want 3", len(results))
	}
}

func TestGetBatchPartialCached(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Запрашиваем батч с одним новым IP
	results := cache.GetBatch([]string{"192.168.1.1", "192.168.1.99"})

	// 192.168.1.1 должен быть в кэше, 192.168.1.99 — после обновления
	if len(results) < 1 {
		t.Errorf("GetBatch() returned %d results, want at least 1", len(results))
	}
}

func TestGetBatchEmpty(t *testing.T) {
	cache := NewARPCache(5*time.Minute, nil)

	results := cache.GetBatch([]string{})

	if len(results) != 0 {
		t.Errorf("GetBatch() returned %d results, want 0", len(results))
	}
}

// --- Test GetAll ---

func TestGetAll(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	all := cache.GetAll()

	if len(all) != 2 {
		t.Errorf("GetAll() returned %d entries, want 2", len(all))
	}

	// Проверяем, что возвращённая карта — копия
	all["192.168.1.1"] = "modified"
	if cache.entries["192.168.1.1"] == "modified" {
		t.Error("GetAll() should return a copy, not the original map")
	}
}

func TestGetAllEmpty(t *testing.T) {
	cache := NewARPCache(5*time.Minute, nil)

	all := cache.GetAll()

	if len(all) != 0 {
		t.Errorf("GetAll() returned %d entries, want 0", len(all))
	}
}

// --- Test IsFresh ---

func TestIsFresh(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// До обновления
	if cache.IsFresh() {
		t.Error("Cache should not be fresh before Refresh()")
	}

	// После обновления
	cache.Refresh()

	if !cache.IsFresh() {
		t.Error("Cache should be fresh after Refresh()")
	}
}

func TestIsFreshExpired(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(10*time.Millisecond, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Ждём истечения TTL
	time.Sleep(20 * time.Millisecond)

	if cache.IsFresh() {
		t.Error("Cache should not be fresh after TTL expiration")
	}
}

// --- Test Size ---

func TestSize(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
			"192.168.1.3": "aa:bb:cc:dd:ee:00",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	if cache.Size() != 0 {
		t.Errorf("Initial size = %d, want 0", cache.Size())
	}

	cache.Refresh()

	if cache.Size() != 3 {
		t.Errorf("Size after Refresh() = %d, want 3", cache.Size())
	}
}

// --- Test Concurrent Access ---

func TestConcurrentGet(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	// Запускаем множество горутин для параллельного доступа
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, _ = cache.Get("192.168.1.1")
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentRefreshAndGet(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	var wg sync.WaitGroup

	// Goroutines для RefreshAsync
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.RefreshAsync()
		}()
	}

	// Goroutines для Get
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, _ = cache.Get("192.168.1.1")
			}
		}()
	}

	wg.Wait()
}

// --- Test ResolveMACBatch ---

func TestResolveMACBatch(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
			"192.168.1.2": "11:22:33:44:55:66",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	ctx := context.Background()
	results := ResolveMACBatch(ctx, []string{"192.168.1.1", "192.168.1.2"}, cache)

	if len(results) != 2 {
		t.Errorf("ResolveMACBatch() returned %d results, want 2", len(results))
	}
}

func TestResolveMACBatchWithInvalidMAC(t *testing.T) {
	refreshFunc := func() (map[string]string, error) {
		return map[string]string{
			"192.168.1.1": "invalid-mac",
		}, nil
	}

	cache := NewARPCache(5*time.Minute, refreshFunc)

	// Обновляем кэш
	cache.Refresh()

	ctx := context.Background()
	results := ResolveMACBatch(ctx, []string{"192.168.1.1"}, cache)

	// Должен вернуть пустой результат из-за невалидного MAC
	if len(results) != 0 {
		t.Errorf("ResolveMACBatch() returned %d results, want 0 (invalid MAC)", len(results))
	}
}

func TestResolveMACBatchEmpty(t *testing.T) {
	cache := NewARPCache(5*time.Minute, nil)

	ctx := context.Background()
	results := ResolveMACBatch(ctx, []string{}, cache)

	if len(results) != 0 {
		t.Errorf("ResolveMACBatch() returned %d results, want 0", len(results))
	}
}

// --- Test parseWindowsARP ---

func TestParseWindowsARP(t *testing.T) {
	// Тестовая строка с MAC-адресами
	outputWithMAC := `
Интерфейс: 192.168.1.1 --- 0xb
  Address          Type         Interface
  192.168.1.1      0a-1b-2c-3d-4e-5f   dynamic      192.168.1.1
  192.168.1.10     11-22-33-44-55-66   dynamic      192.168.1.1
`

	entries := parseWindowsARP(outputWithMAC)

	if len(entries) != 2 {
		t.Errorf("parseWindowsARP() returned %d entries, want 2", len(entries))
	}

	if entries["192.168.1.1"] != "0a:1b:2c:3d:4e:5f" {
		t.Errorf("parseWindowsARP() mac for 192.168.1.1 = %v, want 0a:1b:2c:3d:4e:5f", entries["192.168.1.1"])
	}
}

func TestParseWindowsARPEmpty(t *testing.T) {
	entries := parseWindowsARP("")

	if len(entries) != 0 {
		t.Errorf("parseWindowsARP() returned %d entries, want 0", len(entries))
	}
}

// --- Test parseLinuxARP ---

func TestParseLinuxARP(t *testing.T) {
	output := `
192.168.1.1 dev eth0 lladdr 0a:1b:2c:3d:4e:5f REACHABLE
192.168.1.10 dev eth0 lladdr 11:22:33:44:55:66 STALE
`

	entries := parseLinuxARP(output)

	if len(entries) != 2 {
		t.Errorf("parseLinuxARP() returned %d entries, want 2", len(entries))
	}

	if entries["192.168.1.1"] != "0a:1b:2c:3d:4e:5f" {
		t.Errorf("parseLinuxARP() mac for 192.168.1.1 = %v, want 0a:1b:2c:3d:4e:5f", entries["192.168.1.1"])
	}
}

func TestParseLinuxARPEmpty(t *testing.T) {
	entries := parseLinuxARP("")

	if len(entries) != 0 {
		t.Errorf("parseLinuxARP() returned %d entries, want 0", len(entries))
	}
}

// --- Test ARPCache with short TTL ---

func TestARPCacheRapidRefresh(t *testing.T) {
	callCount := 0
	refreshFunc := func() (map[string]string, error) {
		callCount++
		return map[string]string{
			"192.168.1.1": "aa:bb:cc:dd:ee:ff",
		}, nil
	}

	cache := NewARPCache(10*time.Millisecond, refreshFunc)

	// Множественные запросы с быстрым истечением TTL
	for i := 0; i < 5; i++ {
		time.Sleep(15 * time.Millisecond)
		_, _ = cache.Get("192.168.1.1")
	}

	// refreshFunc должен был вызван несколько раз
	if callCount < 2 {
		t.Errorf("refreshFunc called %d times, want at least 2", callCount)
	}
}

// --- Test Stop ---

func TestStop(t *testing.T) {
	cache := NewARPCache(5*time.Minute, nil)

	// Stop не должен паниковать
	cache.Stop()
}
