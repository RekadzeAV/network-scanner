package wol

import (
	"fmt"
	"net"
	"strings"
)

// SendMagicPacket отправляет стандартный Wake-on-LAN magic packet на UDP порт 9.
func SendMagicPacket(macStr, broadcastAddr string) error {
	_, err := SendMagicPacketWithInterface(macStr, broadcastAddr, "")
	return err
}

// SendMagicPacketWithInterface отправляет WOL packet с опциональным выбором интерфейса.
// Возвращает фактический адрес отправки (broadcast:port).
func SendMagicPacketWithInterface(macStr, broadcastAddr, ifaceName string) (string, error) {
	mac, err := parseMAC(macStr)
	if err != nil {
		return "", err
	}

	targetAddr, err := resolveBroadcastAddr(broadcastAddr, ifaceName)
	if err != nil {
		return "", err
	}

	payload := make([]byte, 6+16*6)
	for i := 0; i < 6; i++ {
		payload[i] = 0xff
	}
	for i := 1; i <= 16; i++ {
		copy(payload[i*6:], mac)
	}
	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return "", err
	}
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", err
	}
	defer c.Close()
	_, err = c.Write(payload)
	if err != nil {
		return "", err
	}
	return targetAddr, nil
}

func parseMAC(s string) ([]byte, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "-", ":"))
	hw, err := net.ParseMAC(s)
	if err != nil {
		return nil, fmt.Errorf("некорректный MAC: %w", err)
	}
	if len(hw) != 6 {
		return nil, fmt.Errorf("ожидается MAC-48")
	}
	return hw, nil
}

func resolveBroadcastAddr(broadcastAddr, ifaceName string) (string, error) {
	broadcastAddr = strings.TrimSpace(broadcastAddr)
	if broadcastAddr != "" {
		if !strings.Contains(broadcastAddr, ":") {
			broadcastAddr += ":9"
		}
		return broadcastAddr, nil
	}

	if strings.TrimSpace(ifaceName) == "" {
		return "255.255.255.255:9", nil
	}

	bcast, err := broadcastFromInterface(ifaceName)
	if err != nil {
		return "", err
	}
	return bcast + ":9", nil
}

func broadcastFromInterface(ifaceName string) (string, error) {
	ifaceName = strings.TrimSpace(ifaceName)
	if ifaceName == "" {
		return "", fmt.Errorf("интерфейс не задан")
	}
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("интерфейс %q не найден: %w", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("не удалось получить адреса интерфейса %q: %w", ifaceName, err)
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP == nil || ipnet.Mask == nil {
			continue
		}
		ip := ipnet.IP.To4()
		if ip == nil {
			continue
		}
		mask := ipnet.Mask
		if len(mask) < 4 {
			continue
		}
		bcast := net.IPv4(
			ip[0]|^mask[0],
			ip[1]|^mask[1],
			ip[2]|^mask[2],
			ip[3]|^mask[3],
		)
		return bcast.String(), nil
	}
	return "", fmt.Errorf("на интерфейсе %q не найден IPv4 адрес", ifaceName)
}
