package network

import (
	"net"
	"strings"
)

// IsIPv4 проверяет, является ли IP-адрес IPv4.
func IsIPv4(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() != nil
}

// IsIPv6 проверяет, является ли IP-адрес IPv6.
func IsIPv6(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() == nil && ip.To16() != nil
}

// IsIP проверяет, является ли строка валидным IP-адресом (IPv4 или IPv6).
func IsIP(ipStr string) bool {
	return net.ParseIP(strings.TrimSpace(ipStr)) != nil
}

// IsCIDR проверяет, является ли строка валидным CIDR (IPv4 или IPv6).
func IsCIDR(cidrStr string) bool {
	_, _, err := net.ParseCIDR(strings.TrimSpace(cidrStr))
	return err == nil
}

// DetectLocalNetworks определяет все локальные сети (IPv4 и IPv6).
// Возвращает списки IPv4 и IPv6 CIDR.
func DetectLocalNetworks() (ipv4CIDRs []string, ipv6CIDRs []string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP
				if ip.IsLoopback() {
					continue
				}
				ones, bits := ipnet.Mask.Size()
				if bits == 32 && ones > 0 {
					ipv4CIDRs = append(ipv4CIDRs, ipnet.String())
				} else if bits == 128 && ones > 0 {
					ipv6CIDRs = append(ipv6CIDRs, ipnet.String())
				}
			}
		}
	}

	return ipv4CIDRs, ipv6CIDRs
}

// PreferredNetwork возвращает предпочитаемую сеть: IPv4 если есть, иначе IPv6.
func PreferredNetwork(ipv4CIDRs, ipv6CIDRs []string) string {
	if len(ipv4CIDRs) > 0 {
		return ipv4CIDRs[0]
	}
	if len(ipv6CIDRs) > 0 {
		return ipv6CIDRs[0]
	}
	return ""
}

// FormatIPForDisplay форматирует IP-адрес для отображения в GUI.
// Добавляет скобки для IPv6 (например, [::1]).
func FormatIPForDisplay(ipStr string) string {
	if IsIPv6(ipStr) {
		return "[" + ipStr + "]"
	}
	return ipStr
}

// ProtocolForHost возвращает "tcp4", "tcp6", "udp4" или "udp6" для dial-операций.
func ProtocolForHost(host string, useTCP bool, preferUDP bool) string {
	prefix := "tcp"
	if preferUDP {
		prefix = "udp"
	}
	if IsIPv6(host) {
		return prefix + "6"
	}
	return prefix + "4"
}

// DialAddress формирует адрес для dial с учётом IPv6.
func DialAddress(host string, port int) string {
	return net.JoinHostPort(host, string(rune('0'+port%10)))
}
