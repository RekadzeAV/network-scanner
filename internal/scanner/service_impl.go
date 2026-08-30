package scanner

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"network-scanner/internal/contracts"
)

// scannerServiceImpl реализация ScannerService
type scannerServiceImpl struct {
	logLevel   string
	mu         sync.RWMutex
	activeScan *activeScan
	isScanning atomic.Bool
	StopScan   atomic.Value // bool
}

type activeScan struct {
	scanner *NetworkScanner
	cancel  context.CancelFunc
	ctx     context.Context
	done    chan struct{}
}

// NewService создаёт ScannerService
func NewService(logLevel string) contracts.ScannerService {
	return &scannerServiceImpl{
		logLevel: logLevel,
	}
}

func (s *scannerServiceImpl) Scan(ctx context.Context, cfg contracts.ScanConfig, onProgress contracts.ProgressHandler) ([]contracts.ScanResult, error) {
	// Проверяем, не запущено ли уже сканирование
	if s.isScanning.Load() {
		return nil, fmt.Errorf("сканирование уже запущено, вызовите Stop() перед новым сканированием")
	}

	// Если контекст nil, создаём фоновый
	if ctx == nil {
		ctx = context.Background()
	}

	// Создаём cancellable контекст для остановки
	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	// Создаём канал для отслеживания завершения
	done := make(chan struct{})

	// Сохраняем activeScan
	s.mu.Lock()
	s.activeScan = &activeScan{
		scanner: ns,
		cancel:  cancel,
		ctx:     scanCtx,
		done:    done,
	}
	s.mu.Unlock()

	// Устанавливаем флаг сканирования
	s.isScanning.Store(true)

	// Запускаем сканирование в отдельной горутине
	go func() {
		defer close(done)
		ns.Scan()
	}()

	// Ждём завершения или отмены контекста
	select {
	case <-scanCtx.Done():
		// Отмена сканирования
		<-done
		s.isScanning.Store(false)
		s.mu.Lock()
		s.activeScan = nil
		s.mu.Unlock()
		return nil, fmt.Errorf("сканирование отменено: %w", scanCtx.Err())
	case <-done:
	}

	// Сбрасываем флаг сканирования
	s.isScanning.Store(false)
	s.mu.Lock()
	s.activeScan = nil
	s.mu.Unlock()

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
	s.mu.RLock()
	active := s.activeScan
	s.mu.RUnlock()

	if active == nil {
		return
	}

	// Отменяем контекст — это прервёт активное сканирование
	active.cancel()

	// Ждём завершения горутины сканирования
	select {
	case <-active.done:
		// Сканирование завершено
	case <-active.ctx.Done():
		// Контекст отменён
	}

	// Сбрасываем флаг сканирования
	s.isScanning.Store(false)

	s.mu.Lock()
	s.activeScan = nil
	s.mu.Unlock()
}
