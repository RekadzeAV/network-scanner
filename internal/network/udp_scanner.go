package network

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// UDPScanState описывает состояние UDP порта после сканирования.
type UDPScanState string

const (
	// UDPOpen — порт ответил на UDP запрос (открыт).
	UDPOpen UDPScanState = "open"
	// UDPClosed — порт закрыт (получен ICMP Destination Unreachable / connection refused).
	UDPClosed UDPScanState = "closed"
	// UDPFiltered — порт фильтруется (таймаут, нет ответа).
	UDPFiltered UDPScanState = "filtered"
	// UDPOpenFiltered — порт может быть открытым или фильтруемым (неопределённо).
	UDPOpenFiltered UDPScanState = "open|filtered"
)

// UDPPortResult описывает результат сканирования одного UDP порта.
type UDPPortResult struct {
	Port   int
	State  UDPScanState
	Detail string
}

// UDPAdvancedScanner — расширенный UDP-сканер с поддержкой параллелизма и
// классификацией состояний портов (open/closed/filtered/open|filtered).
//
// Используется для продвинутого UDP-сканирования. Для базовой проверки
// открытости порта используйте UDPPortScanner из port_scanner.go.
type UDPAdvancedScanner struct {
	Timeout time.Duration
	Threads int
}

// ScanPort сканирует один UDP port и возвращает состояние.
func (s UDPAdvancedScanner) ScanPort(ip string, port int) UDPPortResult {
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 1 * time.Second
	}

	address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("udp", address)
	if err != nil {
		// Ошибка соединения для UDP часто означает ICMP Destination Unreachable → порт закрыт.
		return UDPPortResult{Port: port, State: UDPClosed, Detail: err.Error()}
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	// Отправляем UDP пакет с минимальными данными (не пустой, для сервисов, которые ждут payload).
	_, err = conn.Write([]byte{0x00})
	if err != nil {
		return UDPPortResult{Port: port, State: UDPClosed, Detail: err.Error()}
	}

	// Ждём ответ.
	buffer := make([]byte, 2048)
	_, err = conn.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Таймаут — порт открыт/фильтрован (UDP не гарантирует ответ).
			return UDPPortResult{Port: port, State: UDPOpenFiltered, Detail: "timeout (open|filtered)"}
		}
		return UDPPortResult{Port: port, State: UDPClosed, Detail: err.Error()}
	}

	// Получили ответ — порт открыт.
	return UDPPortResult{Port: port, State: UDPOpen, Detail: "received response"}
}

// ScanPorts сканирует список UDP портов последовательно.
func (s UDPAdvancedScanner) ScanPorts(ip string, ports []int) []UDPPortResult {
	results := make([]UDPPortResult, 0, len(ports))
	for _, port := range ports {
		results = append(results, s.ScanPort(ip, port))
	}
	return results
}

// ScanPortsParallel сканирует UDP порты параллельно с ограничением через семафор.
func (s UDPAdvancedScanner) ScanPortsParallel(ip string, ports []int) []UDPPortResult {
	if s.Threads <= 0 {
		s.Threads = 50
	}

	results := make([]UDPPortResult, len(ports))
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.Threads)

	for i, port := range ports {
		wg.Add(1)
		go func(idx int, p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = s.ScanPort(ip, p)
		}(i, port)
	}

	wg.Wait()
	return results
}

// IsUDPPortOpenAdvanced проверяет, открыт ли UDP порт (расширенная версия
// с классификацией состояний). Для базовой проверки используйте IsUDPPortOpen
// из network.go.
func IsUDPPortOpenAdvanced(host string, port int, timeout time.Duration) bool {
	s := UDPAdvancedScanner{Timeout: timeout}
	res := s.ScanPort(host, port)
	return res.State == UDPOpen || res.State == UDPOpenFiltered
}
