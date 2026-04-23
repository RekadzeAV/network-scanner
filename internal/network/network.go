package network

import (
	"fmt"
	"net"
	"math"
	"strings"
	"time"

	portdb "network-scanner/internal/ports"
)

const maxEnumeratedHosts = 65536

// EstimateHostCount оценивает количество адресов-хостов в CIDR диапазоне без полной генерации списка IP.
func EstimateHostCount(cidr string) (int, error) {
	_, ipnet, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return 0, err
	}
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones
	if hostBits <= 0 {
		return 1, nil
	}
	if bits == 32 {
		// Для IPv4 исключаем network/broadcast только когда это применимо.
		if hostBits == 1 {
			return 2, nil // /31: оба адреса используются
		}
		if hostBits >= 2 {
			return int(math.Pow(2, float64(hostBits))) - 2, nil
		}
	}
	if hostBits > 30 {
		return 0, fmt.Errorf("слишком большой диапазон для оценки: %s", cidr)
	}
	return 1 << hostBits, nil
}

// DetectLocalNetwork определяет локальную сеть автоматически
func DetectLocalNetwork() (string, error) {
	// Получаем интерфейсы с таймаутом (избегаем зависания в Windows)
	interfacesChan := make(chan []net.Interface, 1)
	errChan := make(chan error, 1)
	go func() {
		interfaces, err := net.Interfaces()
		if err != nil {
			errChan <- err
			return
		}
		interfacesChan <- interfaces
	}()

	var interfaces []net.Interface
	select {
	case interfaces = <-interfacesChan:
		// Успешно получили интерфейсы
	case err := <-errChan:
		return "", err
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("таймаут получения сетевых интерфейсов")
	}

	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы и loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем адреса интерфейса с таймаутом (избегаем зависания в Windows)
		addrsChan := make(chan []net.Addr, 1)
		addrErrChan := make(chan error, 1)
		go func() {
			addrs, err := iface.Addrs()
			if err != nil {
				addrErrChan <- err
				return
			}
			addrsChan <- addrs
		}()

		var addrs []net.Addr
		select {
		case addrs = <-addrsChan:
			// Успешно получили адреса
		case <-addrErrChan:
			continue
		case <-time.After(2 * time.Second):
			// Таймаут для получения адресов интерфейса, пропускаем этот интерфейс
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Пропускаем IPv6 и loopback
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			// Определяем маску подсети
			if ipnet, ok := addr.(*net.IPNet); ok {
				mask := ipnet.Mask
				ones, bits := mask.Size()
				if ones > 0 && bits == 32 {
					network := fmt.Sprintf("%s/%d", ipnet.IP.Mask(mask).String(), ones)
					return network, nil
				}
			}
		}
	}

	return "", fmt.Errorf("не найдена активная сеть")
}

// ParseNetworkRange парсит диапазон сети (например, 192.168.1.0/24)
func ParseNetworkRange(network string) ([]net.IP, error) {
	baseIP, ipnet, err := net.ParseCIDR(strings.TrimSpace(network))
	if err != nil {
		return nil, err
	}

	if baseIP.To4() != nil {
		return parseIPv4NetworkRange(ipnet), nil
	}

	return parseIPv6NetworkRange(ipnet)
}

func parseIPv4NetworkRange(ipnet *net.IPNet) []net.IP {
	var ips []net.IP
	networkIP := ipnet.IP.Mask(ipnet.Mask)

	// Получаем broadcast адрес
	broadcast := make(net.IP, len(networkIP))
	copy(broadcast, networkIP)
	for i := range broadcast {
		broadcast[i] |= ^ipnet.Mask[i]
	}

	// Генерируем все IP адреса в подсети, исключая сетевой и broadcast адреса
	ip := make(net.IP, len(networkIP))
	copy(ip, networkIP)
	inc(ip) // Пропускаем сетевой адрес

	for ipnet.Contains(ip) {
		// Пропускаем broadcast адрес
		if ip.Equal(broadcast) {
			break
		}

		// Создаем копию IP для добавления в список
		ipCopy := make(net.IP, 4)
		copy(ipCopy, ip.To4())
		ips = append(ips, ipCopy)

		inc(ip)
	}

	return ips
}

func parseIPv6NetworkRange(ipnet *net.IPNet) ([]net.IP, error) {
	ones, bits := ipnet.Mask.Size()
	if bits != 128 {
		return nil, fmt.Errorf("неверная маска IPv6 сети")
	}
	hostBits := bits - ones
	if hostBits > 16 {
		return nil, fmt.Errorf("слишком большой диапазон IPv6 (%d бит хоста): ограничение /112 или уже", hostBits)
	}
	hostCount := 1 << hostBits
	if hostCount > maxEnumeratedHosts {
		return nil, fmt.Errorf("слишком большой диапазон IPv6: %d адресов (максимум %d)", hostCount, maxEnumeratedHosts)
	}

	base := ipnet.IP.Mask(ipnet.Mask).To16()
	if base == nil {
		return nil, fmt.Errorf("не удалось нормализовать IPv6 адрес")
	}

	ips := make([]net.IP, 0, hostCount)
	curr := make(net.IP, net.IPv6len)
	copy(curr, base)

	for i := 0; i < hostCount; i++ {
		ipCopy := make(net.IP, net.IPv6len)
		copy(ipCopy, curr)
		ips = append(ips, ipCopy)
		inc(curr)
	}

	return ips, nil
}

// inc увеличивает IP адрес на 1 (используется для генерации диапазона адресов)
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// IsPortOpen проверяет, открыт ли TCP порт
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Используем Dialer с явным таймаутом для лучшей работы в Windows
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return false
	}

	// Убеждаемся, что соединение закрыто немедленно
	if conn != nil {
		conn.Close()
	}
	return true
}

// IsUDPPortOpen проверяет, открыт ли UDP порт
// UDP сканирование сложнее TCP, так как UDP не устанавливает соединение
// Метод: отправляем UDP пакет и проверяем ответ (ICMP порт недоступен = закрыт, ответ = открыт)
func IsUDPPortOpen(host string, port int, timeout time.Duration) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Используем Dialer с явным таймаутом
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Пытаемся отправить UDP пакет
	conn, err := dialer.Dial("udp", address)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Устанавливаем таймаут для чтения
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Отправляем пустой пакет (для некоторых сервисов это может вызвать ответ)
	_, err = conn.Write([]byte{})
	if err != nil {
		// Если не можем отправить, порт скорее всего закрыт
		return false
	}

	// Пытаемся прочитать ответ
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, err = conn.Read(buffer)

	// Если получили ответ, порт открыт
	if err == nil {
		return true
	}

	// Если ошибка таймаута, порт может быть открыт (фильтруется) или закрыт
	// Для UDP сложно определить точно без ICMP, но если нет ошибки соединения - считаем открытым/фильтрованным
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		// Таймаут может означать, что порт фильтруется или открыт, но не отвечает
		// В контексте сканирования считаем это потенциально открытым
		return true
	}

	return false
}

// GetServiceName возвращает название сервиса по порту (IANA + локальные подписи).
func GetServiceName(port int) string {
	return portdb.LookupServiceName(port)
}

// ParsePortRange парсит диапазон портов
func ParsePortRange(portRange string) ([]int, error) {
	var ports []int

	// Поддержка форматов: "1-1000", "80,443,8080", "80,443-445"
	parts := strings.Split(portRange, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Диапазон портов
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неверный формат диапазона портов: %s", part)
			}
			start, err := parseInt(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end, err := parseInt(rangeParts[1])
			if err != nil {
				return nil, err
			}
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			// Один порт
			port, err := parseInt(part)
			if err != nil {
				return nil, err
			}
			ports = append(ports, port)
		}
	}

	return ports, nil
}

// parseInt парсит строку в целое число
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
