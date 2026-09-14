package plugin

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
)

// OSFilterPlugin фильтрует результаты по операционной системе
type OSFilterPlugin struct {
	info     Info
	osFilter string
}

// NewOSFilter создает новый плагин фильтрации по ОС
func NewOSFilter() *OSFilterPlugin {
	return &OSFilterPlugin{
		info: Info{
			Name:        "OSFilter",
			Version:     "1.0.0",
			Description: "Фильтрация результатов по операционной системе",
			Author:      "Network Scanner Team",
			Type:        TypeFilter,
		},
		osFilter: "",
	}
}

// Info реализует Plugin
func (p *OSFilterPlugin) Info() Info {
	return p.info
}

// Init реализует Plugin
func (p *OSFilterPlugin) Init(cfg map[string]interface{}) error {
	if osFilter, ok := cfg["os"].(string); ok {
		p.osFilter = osFilter
	}
	return nil
}

// Run реализует Plugin — фильтрация результатов по ОС
func (p *OSFilterPlugin) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	filtered := make([]contracts.ScanResult, 0)

	for _, r := range results {
		if p.osFilter == "" || strings.Contains(strings.ToLower(r.GuessOS), strings.ToLower(p.osFilter)) {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

// Close реализует Plugin
func (p *OSFilterPlugin) Close() error {
	return nil
}

// CSVExporterPlugin экспортирует результаты в CSV
type CSVExporterPlugin struct {
	info Info
}

// NewCSVExporter создает новый плагин экспорта в CSV
func NewCSVExporter() *CSVExporterPlugin {
	return &CSVExporterPlugin{
		info: Info{
			Name:        "CSVExporter",
			Version:     "1.0.0",
			Description: "Экспорт результатов в формат CSV",
			Author:      "Network Scanner Team",
			Type:        TypeExporter,
		},
	}
}

// Info реализует Plugin
func (p *CSVExporterPlugin) Info() Info {
	return p.info
}

// Init реализует Plugin
func (p *CSVExporterPlugin) Init(cfg map[string]interface{}) error {
	return nil
}

// Run реализует Plugin — экспорт в CSV
func (p *CSVExporterPlugin) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	type csvResult struct {
		filename string
		data     string
	}

	var sb strings.Builder
	sb.WriteString("IP,Hostname,MAC,DeviceType,DeviceVendor,GuessOS,OpenPorts\n")

	for _, r := range results {
		ports := make([]int, 0, len(r.Ports))
		for _, port := range r.Ports {
			if port.State == "open" {
				ports = append(ports, port.Port)
			}
		}

		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%v\n",
			r.IP,
			r.Hostname,
			r.MAC,
			r.DeviceType,
			r.DeviceVendor,
			r.GuessOS,
			ports,
		))
	}

	return csvResult{
		filename: "export.csv",
		data:     sb.String(),
	}, nil
}

// Close реализует Plugin
func (p *CSVExporterPlugin) Close() error {
	return nil
}

// ExportCSVToPath экспортирует результаты в файл CSV
func ExportCSVToPath(results []scanner.Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Заголовок
	if err := w.Write([]string{"IP", "Hostname", "MAC", "DeviceType", "DeviceVendor", "GuessOS", "OpenPorts"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	// Данные
	for _, host := range results {
		ports := make([]int, 0, len(host.Ports))
		for _, p := range host.Ports {
			if p.State == "open" {
				ports = append(ports, p.Port)
			}
		}

		record := []string{
			host.IP,
			host.Hostname,
			host.MAC,
			host.DeviceType,
			host.DeviceVendor,
			host.GuessOS,
			fmt.Sprintf("%v", ports),
		}
		if err := w.Write(record); err != nil {
			return fmt.Errorf("write csv record: %w", err)
		}
	}

	return nil
}
