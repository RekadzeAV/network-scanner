package network

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// DetectLocalNetwork определяет локальную сеть автоматически
func DetectLocalNetwork() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы и loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
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
	_, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}

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

// IsPortOpen проверяет, открыт ли порт
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetServiceName возвращает название сервиса по порту
func GetServiceName(port int) string {
	services := map[int]string{
		20:   "FTP-Data",
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		80:   "HTTP",
		110:  "POP3",
		143:  "IMAP",
		443:  "HTTPS",
		445:  "SMB",
		3306: "MySQL",
		3389: "RDP",
		5432: "PostgreSQL",
		5900: "VNC",
		8080: "HTTP-Proxy",
		8443: "HTTPS-Alt",
	}

	if name, ok := services[port]; ok {
		return name
	}
	return "Unknown"
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
