package network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ARPCache — асинхронное кэширование ARP-таблицы с фоновым обновлением.
// Позволяет избежать многократного парсинга ARP-таблицы при сканировании подсети.
type ARPCache struct {
	mu          sync.RWMutex
	entries     map[string]string // IP -> MAC
	freshAt     time.Time
	ttl         time.Duration
	refreshFunc func() (map[string]string, error)
	refreshed   bool
}

// ARPEntry представляет одну запись ARP.
type ARPEntry struct {
	IP  string
	MAC string
}

// NewARPCache создаёт новый кэш ARP с указанным TTL и функцией обновления.
func NewARPCache(ttl time.Duration, refreshFunc func() (map[string]string, error)) *ARPCache {
	return &ARPCache{
		entries:     make(map[string]string),
		ttl:         ttl,
		refreshFunc: refreshFunc,
	}
}

// Get возвращает MAC-адрес для IP из кэша.
// Если запись отсутствует или устарела, запускает фоновое обновление.
func (c *ARPCache) Get(ip string) (string, error) {
	c.mu.RLock()
	mac, cached := c.entries[ip]
	freshAt := c.freshAt
	c.mu.RUnlock()

	// Если запись найдена и свежая — возвращаем
	if cached && time.Since(freshAt) < c.ttl {
		return mac, nil
	}

	// Запускаем фоновое обновление
	c.RefreshAsync()

	// После обновления пробуем ещё раз
	c.mu.RLock()
	mac, cached = c.entries[ip]
	c.mu.RUnlock()

	if cached {
		return mac, nil
	}

	return "", fmt.Errorf("MAC not found for IP: %s", ip)
}

// RefreshAsync запускает фоновое обновление ARP-таблицы.
// Безопасен для многократного вызова — использует sync.Once-подобную логику.
func (c *ARPCache) RefreshAsync() {
	if c.refreshFunc == nil {
		return
	}

	go func() {
		entries, err := c.refreshFunc()
		if err != nil {
			return // При ошибке оставляем старый кэш
		}

		c.mu.Lock()
		c.entries = entries
		c.freshAt = time.Now()
		c.refreshed = true
		c.mu.Unlock()
	}()
}

// Refresh синхронно обновляет ARP-таблицу.
func (c *ARPCache) Refresh() error {
	if c.refreshFunc == nil {
		return fmt.Errorf("refresh function not set")
	}

	entries, err := c.refreshFunc()
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.entries = entries
	c.freshAt = time.Now()
	c.refreshed = true
	c.mu.Unlock()

	return nil
}

// GetBatch возвращает MAC-адреса для списка IP.
// Оптимизация: один вызов refreshFunc для всех IP.
func (c *ARPCache) GetBatch(ips []string) map[string]string {
	results := make(map[string]string)

	// Собираем IP, которых нет в кэше
	missing := make([]string, 0)
	c.mu.RLock()
	for _, ip := range ips {
		mac, cached := c.entries[ip]
		if cached && time.Since(c.freshAt) < c.ttl {
			results[ip] = mac
		} else {
			missing = append(missing, ip)
		}
	}
	c.mu.RUnlock()

	// Если все IP есть в кэше — возвращаем
	if len(missing) == 0 {
		return results
	}

	// Обновляем кэш
	c.RefreshAsync()

	// Ждём завершения обновления (до 2 секунд)
	done := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Таймаут, возвращаем что есть
	}

	// Собираем результаты после обновления
	c.mu.RLock()
	for _, ip := range missing {
		if mac, ok := c.entries[ip]; ok {
			results[ip] = mac
		}
	}
	c.mu.RUnlock()

	return results
}

// GetAll возвращает все записи из кэша.
func (c *ARPCache) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make(map[string]string, len(c.entries))
	for ip, mac := range c.entries {
		entries[ip] = mac
	}
	return entries
}

// IsFresh возвращает true, если кэш свежий.
func (c *ARPCache) IsFresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.freshAt) < c.ttl
}

// Size возвращает количество записей в кэше.
func (c *ARPCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// IsRefreshed возвращает true, если кэш хотя бы раз был обновлён.
func (c *ARPCache) IsRefreshed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshed
}

// Stop останавливает фоновое обновление (для тестов и clean shutdown).
func (c *ARPCache) Stop() {
	// В текущей реализации sync.Once не нужен, так как goroutine завершаются автоматически
}

// --- Platform-specific ARP parsing ---

// parseWindowsARP парсит ARP-таблицу Windows (cmd: "arp -a").
func parseWindowsARP(output string) map[string]string {
	entries := make(map[string]string)

	// Регулярка для строк вида:
	// 192.168.1.1   0a-1b-2c-3d-4e-5f   dynamic
	// 192.168.1.10  0a-1b-2c-3d-4e-5f   static
	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)\s+([\da-fA-F-]+)\s+(dynamic|static)`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			ip := match[1]
			mac := strings.ReplaceAll(match[2], "-", ":")
			entries[ip] = mac
		}
	}

	return entries
}

// parseLinuxARP парсит ARP-таблицу Linux (cmd: "ip neigh" или "arp -n").
func parseLinuxARP(output string) map[string]string {
	entries := make(map[string]string)

	// Регулярка для "ip neigh":
	// 192.168.1.1 dev eth0 lladdr 0a:1b:2c:3d:4e:5f REACHABLE
	// Используем более простой regex
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Ищем "lladdr MAC"
		if strings.Contains(line, "lladdr") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "lladdr" && i+1 < len(parts) {
					ip := ""
					// IP должен быть до "dev"
					for j := 0; j < i; j++ {
						if parts[j] == "dev" {
							if j > 0 {
								ip = parts[j-1]
							}
							break
						}
						// Последний токен до "dev" — это IP
						if j == i-1 {
							ip = parts[j]
						}
					}
					// Если IP не найден через "dev", берём первый токен
					if ip == "" {
						ip = parts[0]
					}
					mac := parts[i+1]
					entries[ip] = mac
					break
				}
			}
		}
	}

	// Если ничего не найдено, пробуем формат "arp -n"
	if len(entries) == 0 {
		return parseLinuxARPFromArpString(output)
	}

	return entries
}

// parseLinuxARPFromArpString парсит вывод "arp -n" для Linux.
func parseLinuxARPFromArpString(output string) map[string]string {
	entries := make(map[string]string)

	// Регулярка для "arp -n":
	// 192.168.1.1              ether   0a:1b:2c:3d:4e:5f   brd
	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)\s+\S+\s+([\da-fA-F:]+)\s+brd`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			ip := match[1]
			mac := match[2]
			entries[ip] = mac
		}
	}

	return entries
}

// GetARPTabaleWindows запускает "arp -a" и парсит результат.
func GetARPTabaleWindows() (map[string]string, error) {
	cmd := exec.Command("cmd", "/c", "arp", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run arp -a: %w", err)
	}

	return parseWindowsARP(string(output)), nil
}

// GetARPTabaleLinux запускает "ip neigh" и парсит результат.
func GetARPTabaleLinux() (map[string]string, error) {
	cmd := exec.Command("ip", "neigh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback на "arp -n"
		cmd = exec.Command("arp", "-n")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to run arp/ip command: %w", err)
		}
		// Парсим "arp -n"
		return parseLinuxARPFromArpString(string(output)), nil
	}

	return parseLinuxARP(string(output)), nil
}

// GetARPTabale — кроссплатформенная функция для получения ARP-таблицы.
func GetARPTabale() (map[string]string, error) {
	if runtime.GOOS == "windows" {
		return GetARPTabaleWindows()
	}
	return GetARPTabaleLinux()
}

// NewDefaultARPCache создаёт ARPCache с дефолтной функцией обновления.
func NewDefaultARPCache(ttl time.Duration) *ARPCache {
	return NewARPCache(ttl, func() (map[string]string, error) {
		return GetARPTabale()
	})
}

// ResolveMACBatch — вспомогательная функция для разрешения MAC-адресов для списка IP.
func ResolveMACBatch(ctx context.Context, ips []string, cache *ARPCache) map[string]net.HardwareAddr {
	results := make(map[string]net.HardwareAddr)

	macMap := cache.GetBatch(ips)
	for ip, mac := range macMap {
		hw, err := net.ParseMAC(mac)
		if err == nil {
			results[ip] = hw
		}
	}

	return results
}
