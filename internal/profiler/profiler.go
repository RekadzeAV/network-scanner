package profiler

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler утилита для CPU и memory profiling
type Profiler struct {
	cpuFile    *os.File
	memFile    *os.File
	startTime  time.Time
	profileDir string
}

// NewProfiler создаёт новый Profiler
func NewProfiler(profileDir string) (*Profiler, error) {
	if profileDir == "" {
		profileDir = "profile"
	}
	
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profile directory: %w", err)
	}
	
	cpuPath := fmt.Sprintf("%s/cpu.profile", profileDir)
	memPath := fmt.Sprintf("%s/memory.profile", profileDir)
	
	cpuFile, err := os.Create(cpuPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	
	memFile, err := os.Create(memPath)
	if err != nil {
		cpuFile.Close()
		return nil, fmt.Errorf("failed to create memory profile file: %w", err)
	}
	
	return &Profiler{
		cpuFile:    cpuFile,
		memFile:    memFile,
		startTime:  time.Now(),
		profileDir: profileDir,
	}, nil
}

// Start запускает profiling
func (p *Profiler) Start() error {
	if err := pprof.StartCPUProfile(p.cpuFile); err != nil {
		p.cpuFile.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	return nil
}

// Stop останавливает profiling и сохраняет данные
func (p *Profiler) Stop() error {
	duration := time.Since(p.startTime)
	
	pprof.StopCPUProfile()
	
	// Memory profile
	runtime.GC()
	if err := pprof.WriteHeapProfile(p.memFile); err != nil {
		p.memFile.Close()
		return fmt.Errorf("failed to write memory profile: %w", err)
	}
	
	p.cpuFile.Close()
	p.memFile.Close()
	
	fmt.Printf("Profile saved to %s/\n", p.profileDir)
	fmt.Printf("  - CPU: %s/cpu.profile\n", p.profileDir)
	fmt.Printf("  - Memory: %s/memory.profile\n", p.profileDir)
	fmt.Printf("Duration: %v\n", duration)
	
	return nil
}

// QuickProfile создаёт и запускает profiling для быстрого анализа
func QuickProfile(profileDir string) (*Profiler, func(), error) {
	profiler, err := NewProfiler(profileDir)
	if err != nil {
		return nil, nil, err
	}
	
	if err := profiler.Start(); err != nil {
		return nil, nil, err
	}
	
	stop := func() {
		profiler.Stop()
	}
	
	return profiler, stop, nil
}
