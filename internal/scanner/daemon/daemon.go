package daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"network-scanner/internal/scanner"
)

type EventKind string

const (
	EventProgress EventKind = "progress"
	EventDone     EventKind = "done"
	EventError    EventKind = "error"
	EventStopped  EventKind = "stopped"
)

type Event struct {
	Kind        EventKind
	Stage       string
	Current     int
	Total       int
	Message     string
	Percent     float64
	Results     []scanner.Result
	Diagnostics string
	Err         error
}

type Config struct {
	NetworkCIDR    string
	Timeout        time.Duration
	PortRange      string
	Threads        int
	ShowClosed     bool
	ScanTCPPorts   bool
	ScanUDP        bool
	GrabBanners    bool
	OSDetectActive bool
	VerbosePortLog bool
}

type Runner struct {
	mu      sync.RWMutex
	scanner *scanner.NetworkScanner
	cancel  context.CancelFunc
	running bool
	events  chan Event
	factory func(Config) *scanner.NetworkScanner
}

func NewRunner() *Runner {
	return NewRunnerWithFactory(nil)
}

func NewRunnerWithFactory(factory func(Config) *scanner.NetworkScanner) *Runner {
	if factory == nil {
		factory = func(cfg Config) *scanner.NetworkScanner {
			return scanner.NewNetworkScanner(cfg.NetworkCIDR, cfg.Timeout, cfg.PortRange, cfg.Threads, cfg.ShowClosed)
		}
	}
	return &Runner{
		events:  make(chan Event, 256),
		factory: factory,
	}
}

func (r *Runner) Events() <-chan Event {
	return r.events
}

func (r *Runner) Start(cfg Config) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return errors.New("scan runner is already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	ns := r.factory(cfg)
	if ns == nil {
		r.cancel = nil
		r.running = false
		r.mu.Unlock()
		err := errors.New("scanner factory returned nil scanner")
		r.emit(Event{Kind: EventError, Message: err.Error(), Err: err})
		return err
	}
	r.running = true
	ns.SetScanTCPPorts(cfg.ScanTCPPorts)
	ns.SetScanUDP(cfg.ScanUDP)
	ns.SetGrabBanners(cfg.GrabBanners)
	ns.SetOSDetectActive(cfg.OSDetectActive)
	ns.SetVerbosePortLogs(cfg.VerbosePortLog)
	ns.SetProgressCallback(func(stage string, current int, total int, message string) {
		percent := 0.0
		if total > 0 {
			percent = float64(current) / float64(total)
		}
		r.emit(Event{
			Kind:    EventProgress,
			Stage:   stage,
			Current: current,
			Total:   total,
			Message: message,
			Percent: percent,
		})
	})
	r.scanner = ns
	r.mu.Unlock()

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("scan runner panic: %v", rec)
				r.emit(Event{Kind: EventError, Message: err.Error(), Err: err})
			}
			r.mu.Lock()
			r.running = false
			r.cancel = nil
			r.scanner = nil
			r.mu.Unlock()
		}()

		done := make(chan struct{})
		go func() {
			defer close(done)
			ns.Scan()
		}()

		select {
		case <-ctx.Done():
			ns.Stop()
			r.emit(Event{Kind: EventStopped, Message: "Сканирование остановлено"})
			return
		case <-done:
			r.emit(Event{
				Kind:        EventDone,
				Results:     ns.GetResults(),
				Diagnostics: ns.GetDiagnosticsSummary(),
			})
		}
	}()
	return nil
}

func (r *Runner) Stop() {
	r.mu.RLock()
	cancel := r.cancel
	r.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
}

func (r *Runner) CurrentScanner() *scanner.NetworkScanner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.scanner
}

func (r *Runner) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

func (r *Runner) emit(ev Event) {
	select {
	case r.events <- ev:
	default:
	}
}
