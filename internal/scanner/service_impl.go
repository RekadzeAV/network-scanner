package scanner

import (
	"context"

	"network-scanner/internal/contracts"
)

// scannerServiceImpl реализация ScannerService
type scannerServiceImpl struct {
	logLevel string
}

// NewService создаёт ScannerService
func NewService(logLevel string) contracts.ScannerService {
	return &scannerServiceImpl{logLevel: logLevel}
}

func (s *scannerServiceImpl) Scan(ctx context.Context, cfg contracts.ScanConfig, onProgress contracts.ProgressHandler) ([]contracts.ScanResult, error) {
	// Если контекст nil, создаём фоновый
	if ctx == nil {
		ctx = context.Background()
	}

	// Создаём NetworkScanner с параметрами из ScanConfig
	ns := NewNetworkScanner(
		cfg.NetworkCIDR,
		cfg.Timeout,
		cfg.PortRange,
		cfg.Threads,
		false, // showClosed
	)

	ns.SetScanUDP(cfg.ScanUDP)
	ns.SetGrabBanners(cfg.GrabBanners)
	ns.SetOSDetectActive(cfg.OSActive)
	ns.SetVerbosePortLogs(cfg.VerboseLogs)

	// Обёртка для ProgressHandler
	if onProgress != nil {
		ns.SetProgressCallback(func(stage string, current, total int, message string) {
			onProgress(stage, current, total, message)
		})
	}

	// Запускаем сканирование в отдельной горутине с контекстом
	done := make(chan struct{})
	go func() {
		ns.Scan()
		close(done)
	}()

	// Ждём завершения или отмены контекста
	select {
	case <-ctx.Done():
		// Отмена сканирования
		<-done
		return nil, ctx.Err()
	case <-done:
	}

	// Конвертируем результаты
	rawResults := ns.GetResults()
	results := make([]contracts.ScanResult, 0, len(rawResults))
	for _, r := range rawResults {
		ports := make([]contracts.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, contracts.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}

		results = append(results, contracts.ScanResult{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	return results, nil
}

func (s *scannerServiceImpl) Stop() {
	// Реализация остановки (может потребовать глобального состояния)
}
