package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// ParseTargetsFromFile reads a file containing target IPs/CIDRs and returns a list of IP addresses.
// Supported formats:
//   - One IP per line: 192.168.1.1
//   - CIDR notation: 192.168.1.0/24
//   - IP ranges: 192.168.1.1-10
//   - Comments starting with #
//   - Empty lines are ignored
func ParseTargetsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open targets file: %w", err)
	}
	defer file.Close()

	ips := make([]string, 0)
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try to parse as CIDR first
		if strings.Contains(line, "/") {
			cidrIPs, err := ParseNetworkRange(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid CIDR %q: %w", lineNum, line, err)
			}
			for _, ip := range cidrIPs {
				ips = append(ips, ip.String())
			}
			continue
		}

		// Try to parse as IP range (e.g., 192.168.1.1-10)
		if strings.Contains(line, "-") {
			rangeIPs, err := parseIPRange(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid IP range %q: %w", lineNum, line, err)
			}
			ips = append(ips, rangeIPs...)
			continue
		}

		// Try to parse as single IP
		ip := net.ParseIP(line)
		if ip == nil {
			return nil, fmt.Errorf("line %d: invalid IP address %q", lineNum, line)
		}
		ips = append(ips, ip.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading targets file: %w", err)
	}

	return ips, nil
}

// parseIPRange parses an IP range like "192.168.1.1-10" and returns a list of IPs.
// The range is inclusive: "1-3" returns 1,2,3 (3 addresses).
func parseIPRange(rangeStr string) ([]string, error) {
	parts := strings.SplitN(rangeStr, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid IP range format: %s", rangeStr)
	}

	baseIP := net.ParseIP(parts[0])
	if baseIP == nil {
		return nil, fmt.Errorf("invalid base IP in range: %s", parts[0])
	}

	endStr := strings.TrimSpace(parts[1])
	end, err := strconv.Atoi(endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end IP in range: %s", endStr)
	}

	if end < 0 {
		return nil, fmt.Errorf("invalid end IP in range (must be >= 0): %s", endStr)
	}

	// Determine the IP version from baseIP
	isIPv4 := baseIP.To4() != nil

	ips := make([]string, 0, end+1)

	if isIPv4 {
		// For IPv4, work with the last 4 bytes
		baseBytes := baseIP.To4()
		baseInt := (int(baseBytes[0]) << 24) | (int(baseBytes[1]) << 16) | (int(baseBytes[2]) << 8) | int(baseBytes[3])

		for i := 0; i < end; i++ {
			ipInt := baseInt + i
			ip := net.IPv4(
				byte((ipInt>>24)&0xFF),
				byte((ipInt>>16)&0xFF),
				byte((ipInt>>8)&0xFF),
				byte(ipInt&0xFF),
			).String()
			ips = append(ips, ip)
		}
	} else {
		// For IPv6, increment the last 2 bytes (little-endian)
		baseBytes := baseIP.To16()
		for i := 0; i < end; i++ {
			ipBytes := make([]byte, 16)
			copy(ipBytes, baseBytes)

			offset := i
			ipBytes[14] = byte(offset & 0xFF)
			ipBytes[15] = byte((offset >> 8) & 0xFF)

			ip := net.IP(ipBytes)
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}
