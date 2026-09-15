package gui

import (
	"context"
	"fmt"
	"time"

	"network-scanner/internal/audit"
	"network-scanner/internal/contracts"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

// AuditResult результат аудита безопасности
type AuditResult struct {
	Entries  []string
	Total    int
	Duration time.Duration
	Findings []contracts.Finding
	Score    int
}

// AuditService обёртка для аудита
type AuditService struct {
}

// NewAuditService создаёт AuditService
func NewAuditService() *AuditService {
	return &AuditService{}
}

// RunAudit запускает аудит безопасности с реальными данными
func (s *AuditService) RunAudit(ctx context.Context, results []contracts.ScanResult, minSeverity string, timeout time.Duration) (*AuditResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no scan results to audit")
	}

	start := time.Now()

	// Конвертируем contracts.ScanResult в scanner.Result
	scanResults := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}

		scanResults = append(scanResults, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	// Реальный вызов audit
	findings := audit.EvaluateOpenPorts(scanResults)

	// Фильтрация по severity
	if minSeverity != "" {
		findings = audit.FilterByMinSeverity(findings, minSeverity)
	}

	// Конвертируем findings в entries и findings для UI
	entries := make([]string, 0, len(findings))
	guiFindings := make([]contracts.Finding, 0, len(findings))

	for _, f := range findings {
		entries = append(entries, fmt.Sprintf("[%s] %s %d/%s: %s. Рекомендация: %s",
			f.Severity, f.Host, f.Port, f.Protocol, f.Title, f.Recommendation))

		guiFindings = append(guiFindings, contracts.Finding{
			Severity:       f.Severity,
			Title:          f.Title,
			Recommendation: f.Recommendation,
		})
	}

	// Вычисляем score
	summary := audit.BuildSummary(findings)

	return &AuditResult{
		Entries:  entries,
		Total:    len(entries),
		Duration: time.Since(start),
		Findings: guiFindings,
		Score:    summary.OverallRiskScore,
	}, nil
}

// RiskSignatureResult результат risk signature анализа
type RiskSignatureResult struct {
	Entries  []string
	Total    int
	Duration time.Duration
}

// RiskSignatureService обёртка для risk signature
type RiskSignatureService struct {
}

// NewRiskSignatureService создаёт RiskSignatureService
func NewRiskSignatureService() *RiskSignatureService {
	return &RiskSignatureService{}
}

// RunRiskSignatures запускает анализ risk signatures с реальными данными
func (s *RiskSignatureService) RunRiskSignatures(ctx context.Context, results []contracts.ScanResult, timeout time.Duration) (*RiskSignatureResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no scan results to analyze")
	}

	start := time.Now()

	// Конвертируем contracts.ScanResult в scanner.Result
	scanResults := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}

		scanResults = append(scanResults, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	// Реальный вызов risksignature
	db, err := risksignature.LoadDefault()
	if err != nil {
		return &RiskSignatureResult{
			Entries:  []string{fmt.Sprintf("Ошибка загрузки сигнатур: %v", err)},
			Total:    1,
			Duration: time.Since(start),
		}, nil
	}

	sigs := risksignature.Evaluate(scanResults, db)

	entries := make([]string, 0, len(sigs))
	for _, sig := range sigs {
		entries = append(entries, fmt.Sprintf("[%s] %s: %s", sig.Severity, sig.HostIP, sig.Title))
	}

	return &RiskSignatureResult{
		Entries:  entries,
		Total:    len(entries),
		Duration: time.Since(start),
	}, nil
}
