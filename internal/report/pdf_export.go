package report

import (
	"fmt"

	"network-scanner/internal/contracts"
	"github.com/jung-kurt/gofpdf/v2"
)

// PDFReport генерирует PDF отчёт со сканированием
type PDFReport struct {
	pdf      *gofpdf.Fpdf
	title    string
	metadata map[string]string
}

// NewPDFReport создаёт новый PDF отчёт
func NewPDFReport(title string) *PDFReport {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	return &PDFReport{
		pdf:      pdf,
		title:    title,
		metadata: make(map[string]string),
	}
}

// AddMetadata добавляет метаданные в отчёт
func (r *PDFReport) AddMetadata(key, value string) {
	r.metadata[key] = value
}

// AddScanResults добавляет результаты сканирования в PDF
func (r *PDFReport) AddScanResults(results []contracts.ScanResult) {
	r.pdf.SetFont("Arial", "B", 14)
	r.pdf.Cell(0, 10, "Scan Results")
	r.pdf.Ln(5)

	r.pdf.SetFont("Arial", "", 10)
	for _, res := range results {
		r.pdf.Cell(30, 8, res.IP)
		if res.Hostname != "" {
			r.pdf.Cell(40, 8, res.Hostname)
		} else {
			r.pdf.Cell(40, 8, "-")
		}
		r.pdf.Cell(20, 8, fmt.Sprintf("%d ports", len(res.Ports)))
		if res.GuessOS != "" {
			r.pdf.Cell(30, 8, res.GuessOS)
		} else {
			r.pdf.Cell(30, 8, "-")
		}
		r.pdf.Ln(8)
	}
}

// AddSecurityFindings добавляет находки безопасности в PDF
func (r *PDFReport) AddSecurityFindings(findings []contracts.Finding) {
	r.pdf.AddPage()
	r.pdf.SetFont("Arial", "B", 14)
	r.pdf.Cell(0, 10, "Security Findings")
	r.pdf.Ln(5)

	r.pdf.SetFont("Arial", "", 10)
	for _, finding := range findings {
		r.pdf.SetFont("Arial", "B", 10)
		r.pdf.Cell(0, 8, fmt.Sprintf("[%s] %s", finding.Severity, finding.Title))
		r.pdf.Ln(8)
		r.pdf.SetFont("Arial", "", 9)
		r.pdf.MultiCell(0, 5, finding.Recommendation, "P", "L", false)
		r.pdf.Ln(3)
	}
}

// AddTopology добавляет топологию сети в PDF
func (r *PDFReport) AddTopology(topology *contracts.Topology) {
	r.pdf.AddPage()
	r.pdf.SetFont("Arial", "B", 14)
	r.pdf.Cell(0, 10, "Network Topology")
	r.pdf.Ln(5)

	r.pdf.SetFont("Arial", "", 10)
	r.pdf.Cell(0, 8, fmt.Sprintf("Devices: %d, Links: %d", len(topology.Devices), len(topology.Links)))
	r.pdf.Ln(8)

	for _, device := range topology.Devices {
		r.pdf.Cell(30, 8, device.IP)
		if device.Hostname != "" {
			r.pdf.Cell(40, 8, device.Hostname)
		} else {
			r.pdf.Cell(40, 8, "-")
		}
		if device.Type != "" {
			r.pdf.Cell(30, 8, device.Type)
		} else {
			r.pdf.Cell(30, 8, "-")
		}
		r.pdf.Ln(8)
	}
}

// Save сохраняет PDF в файл
func (r *PDFReport) Save(path string) error {
	return r.pdf.OutputFileAndClose(path)
}

// Bytes возвращает PDF в виде байтов
func (r *PDFReport) Bytes() ([]byte, error) {
	var buf []byte
	bufWriter := &byteWriter{buf: &buf}
	err := r.pdf.Output(bufWriter)
	return buf, err
}

// byteWriter помогает записать PDF в байты
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// HTMLReportOptions опции для HTML отчёта
type HTMLReportOptions struct {
	IncludeScanResults bool
	IncludeSecurity    bool
	IncludeTopology    bool
	RedactSensitive    bool
}

// DefaultHTMLReportOptions возвращает опции по умолчанию
func DefaultHTMLReportOptions() HTMLReportOptions {
	return HTMLReportOptions{
		IncludeScanResults: true,
		IncludeSecurity:    true,
		IncludeTopology:    true,
		RedactSensitive:    false,
	}
}


