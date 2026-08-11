package plugin

import (
	"context"
	"testing"
)

// ============================================================================
// OSFilterPlugin — полное покрытие
// ============================================================================

func TestNewOSFilter(t *testing.T) {
	p := NewOSFilter()
	if p == nil {
		t.Fatal("NewOSFilter returned nil")
	}
	if p.info.Name != "OSFilter" {
		t.Errorf("Name = %q, want OSFilter", p.info.Name)
	}
	if p.info.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", p.info.Version)
	}
	if p.info.Type != TypeFilter {
		t.Errorf("Type = %q, want filter", p.info.Type)
	}
	if p.osFilter != "" {
		t.Error("osFilter should be empty by default")
	}
}

func TestOSFilterPlugin_Info(t *testing.T) {
	p := NewOSFilter()
	info := p.Info()
	if info.Name != "OSFilter" {
		t.Errorf("Name = %q, want OSFilter", info.Name)
	}
	if info.Description != "Фильтрация результатов по операционной системе" {
		t.Errorf("Description mismatch")
	}
	if info.Author != "Network Scanner Team" {
		t.Errorf("Author mismatch")
	}
}

func TestOSFilterPlugin_Init_NoConfig(t *testing.T) {
	p := NewOSFilter()
	err := p.Init(nil)
	if err != nil {
		t.Errorf("Init(nil) should succeed, got %v", err)
	}
}

func TestOSFilterPlugin_Init_EmptyConfig(t *testing.T) {
	p := NewOSFilter()
	err := p.Init(map[string]interface{}{})
	if err != nil {
		t.Errorf("Init(empty) should succeed, got %v", err)
	}
}

func TestOSFilterPlugin_Init_WithOS(t *testing.T) {
	p := NewOSFilter()
	err := p.Init(map[string]interface{}{"os": "Linux"})
	if err != nil {
		t.Errorf("Init with os should succeed, got %v", err)
	}
	if p.osFilter != "Linux" {
		t.Errorf("osFilter = %q, want Linux", p.osFilter)
	}
}

func TestOSFilterPlugin_Init_WrongType(t *testing.T) {
	p := NewOSFilter()
	err := p.Init(map[string]interface{}{"os": 123})
	if err != nil {
		t.Errorf("Init with wrong type should succeed, got %v", err)
	}
	if p.osFilter != "" {
		t.Error("osFilter should remain empty for wrong type")
	}
}

func TestOSFilterPlugin_Run_Empty(t *testing.T) {
	p := NewOSFilter()
	ctx := context.Background()
	var results []interface{}

	out, err := p.Run(ctx, results)
	if err != nil {
		t.Errorf("Run should succeed, got %v", err)
	}
	if out == nil {
		t.Error("Run should return results")
	}
}

func TestOSFilterPlugin_Run_WithResults(t *testing.T) {
	p := NewOSFilter()
	ctx := context.Background()
	results := []interface{}{
		map[string]interface{}{"IP": "192.168.1.1", "GuessOS": "Linux"},
		map[string]interface{}{"IP": "192.168.1.2", "GuessOS": "Windows"},
	}

	out, err := p.Run(ctx, results)
	if err != nil {
		t.Errorf("Run should succeed, got %v", err)
	}
	if out == nil {
		t.Error("Run should return results")
	}
}

func TestOSFilterPlugin_Close(t *testing.T) {
	p := NewOSFilter()
	err := p.Close()
	if err != nil {
		t.Errorf("Close should succeed, got %v", err)
	}
}

// ============================================================================
// CSVExporterPlugin — полное покрытие
// ============================================================================

func TestNewCSVExporter(t *testing.T) {
	p := NewCSVExporter()
	if p == nil {
		t.Fatal("NewCSVExporter returned nil")
	}
	if p.info.Name != "CSVExporter" {
		t.Errorf("Name = %q, want CSVExporter", p.info.Name)
	}
	if p.info.Type != TypeExporter {
		t.Errorf("Type = %q, want exporter", p.info.Type)
	}
}

func TestCSVExporterPlugin_Info(t *testing.T) {
	p := NewCSVExporter()
	info := p.Info()
	if info.Name != "CSVExporter" {
		t.Errorf("Name = %q, want CSVExporter", info.Name)
	}
	if info.Description != "Экспорт результатов в формат CSV" {
		t.Errorf("Description mismatch")
	}
	if info.Author != "Network Scanner Team" {
		t.Errorf("Author mismatch")
	}
}

func TestCSVExporterPlugin_Init_NoConfig(t *testing.T) {
	p := NewCSVExporter()
	err := p.Init(nil)
	if err != nil {
		t.Errorf("Init(nil) should succeed, got %v", err)
	}
}

func TestCSVExporterPlugin_Init_EmptyConfig(t *testing.T) {
	p := NewCSVExporter()
	err := p.Init(map[string]interface{}{})
	if err != nil {
		t.Errorf("Init(empty) should succeed, got %v", err)
	}
}

func TestCSVExporterPlugin_Run_NotImplemented(t *testing.T) {
	p := NewCSVExporter()
	ctx := context.Background()

	_, err := p.Run(ctx, nil)
	if err == nil {
		t.Fatal("Run should return error for not implemented")
	}
}

func TestCSVExporterPlugin_Close(t *testing.T) {
	p := NewCSVExporter()
	err := p.Close()
	if err != nil {
		t.Errorf("Close should succeed, got %v", err)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkNewOSFilter(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOSFilter()
	}
}

func BenchmarkNewCSVExporter(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCSVExporter()
	}
}
