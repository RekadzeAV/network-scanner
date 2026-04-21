package wol

import (
	"fmt"
	"net"
	"strings"
)

// SendMagicPacket отправляет стандартный Wake-on-LAN magic packet на UDP порт 9.
func SendMagicPacket(macStr, broadcastAddr string) error {
	mac, err := parseMAC(macStr)
	if err != nil {
		return err
	}
	if broadcastAddr == "" {
		broadcastAddr = "255.255.255.255:9"
	}
	if !strings.Contains(broadcastAddr, ":") {
		broadcastAddr = broadcastAddr + ":9"
	}
	payload := make([]byte, 6+16*6)
	for i := 0; i < 6; i++ {
		payload[i] = 0xff
	}
	for i := 1; i <= 16; i++ {
		copy(payload[i*6:], mac)
	}
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer c.Close()
	_, err = c.Write(payload)
	return err
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
