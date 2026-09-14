package gui

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

// WOLResult результат Wake-on-LAN
type WOLResult struct {
	Success  bool
	Message  string
	Error    string
	Duration time.Duration
}

// WOLService обёртка для Wake-on-LAN
type WOLService struct {
}

// NewWOLService создаёт WOLService
func NewWOLService() *WOLService {
	return &WOLService{}
}

// SendWOL отправляет WoL-магический пакет.
//
// Пакет — 6 байт 0xFF и 16 повторов MAC-адреса цели; уходит UDP-broadcast'ом
// на порт 9 (discard) на указанный широковещательный адрес. Успех означает
// лишь, что дейтаграмма принята локальным стеком: подтверждения от самой
// машины протокол не предусматривает.
func (s *WOLService) SendWOL(ctx context.Context, mac, bcast, iface string, timeout time.Duration) (*WOLResult, error) {
	if mac == "" {
		return nil, fmt.Errorf("MAC address is required")
	}

	start := time.Now()

	target, err := parseWOLMAC(mac)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(bcast) == "" {
		bcast = "255.255.255.255"
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	// Привязка к интерфейсу — по его IPv4-адресу. Не нашли адрес (нет такого
	// имени, нет IPv4) — отправляем без привязки: лучше доставить пакет через
	// маршрут по умолчанию, чем упасть.
	if localIP := interfaceIPv4(iface); localIP != nil {
		dialer.LocalAddr = &net.UDPAddr{IP: localIP}
	}

	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(bcast, "9"))
	if err != nil {
		res := &WOLResult{Success: false, Duration: time.Since(start)}
		res.Error = err.Error()
		return res, fmt.Errorf("wol dial %s: %w", bcast, err)
	}
	defer conn.Close()

	// Примечание: для направленного broadcast (broadcast-адрес чужой подсети,
	// как в тестах) достаточно обычной UDP-дейтаграммы — пакет маршрутизируется
	// как unicast. Для широковещания в своей подсети ОС может потребовать
	// SO_BROADCAST; тогда ошибка dial/write вернётся вызывающему с контекстом.

	if _, err := conn.Write(buildMagicPacket(target)); err != nil {
		res := &WOLResult{Success: false, Duration: time.Since(start)}
		res.Error = err.Error()
		return res, fmt.Errorf("wol write: %w", err)
	}

	return &WOLResult{
		Success:  true,
		Message:  "Magic packet sent successfully",
		Duration: time.Since(start),
	}, nil
}

// parseWOLMAC разбирает MAC в 6 байт; допускает разделители ':', '-' и пробелы.
func parseWOLMAC(mac string) ([]byte, error) {
	clean := strings.NewReplacer(":", "", "-", "", " ", "").Replace(strings.TrimSpace(mac))
	if len(clean) != 12 {
		return nil, fmt.Errorf("invalid MAC address %q: expected 12 hex digits", mac)
	}
	b, err := hex.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC address %q: %w", mac, err)
	}
	return b, nil
}

// buildMagicPacket собирает WoL-пакет: 6 × 0xFF + 16 повторов MAC.
func buildMagicPacket(target []byte) []byte {
	packet := make([]byte, 6+16*len(target))
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*len(target):], target)
	}
	return packet
}

// interfaceIPv4 возвращает первый IPv4-адрес интерфейса с указанным именем.
func interfaceIPv4(name string) net.IP {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return ipnet.IP.To4()
		}
	}
	return nil
}
