package network

import (
	"fmt"
	"time"
)

// TCPPortScanner scans TCP ports via net dial checks.
type TCPPortScanner struct {
	Timeout time.Duration
}

// ScanPort scans a single TCP port.
func (s TCPPortScanner) ScanPort(ip string, port int, proto string) (bool, error) {
	if proto != "" && proto != "tcp" {
		return false, fmt.Errorf("tcp scanner does not support protocol: %s", proto)
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = time.Second
	}
	return IsPortOpen(ip, port, timeout), nil
}

// ScanPorts scans the provided ports for TCP.
func (s TCPPortScanner) ScanPorts(ip string, ports []int, proto string) ([]int, error) {
	open := make([]int, 0, len(ports))
	for _, port := range ports {
		isOpen, err := s.ScanPort(ip, port, proto)
		if err != nil {
			return nil, err
		}
		if isOpen {
			open = append(open, port)
		}
	}
	return open, nil
}

// UDPPortScanner scans UDP ports via probe checks.
type UDPPortScanner struct {
	Timeout time.Duration
}

// ScanPort scans a single UDP port.
func (s UDPPortScanner) ScanPort(ip string, port int, proto string) (bool, error) {
	if proto != "" && proto != "udp" {
		return false, fmt.Errorf("udp scanner does not support protocol: %s", proto)
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = time.Second
	}
	return IsUDPPortOpen(ip, port, timeout), nil
}

// ScanPorts scans the provided ports for UDP.
func (s UDPPortScanner) ScanPorts(ip string, ports []int, proto string) ([]int, error) {
	open := make([]int, 0, len(ports))
	for _, port := range ports {
		isOpen, err := s.ScanPort(ip, port, proto)
		if err != nil {
			return nil, err
		}
		if isOpen {
			open = append(open, port)
		}
	}
	return open, nil
}
