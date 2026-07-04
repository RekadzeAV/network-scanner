package security

import (
	"context"

	"network-scanner/internal/audit"
	"network-scanner/internal/contracts"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

// securityServiceImpl реализация SecurityService
type securityServiceImpl struct{}

// NewService создаёт SecurityService
func NewService() contracts.SecurityService {
	return &securityServiceImpl{}
}

func (s *securityServiceImpl) AnalyzeRun(ctx context.Context, results []contracts.ScanResult) (*contracts.SecurityReport, error) {
	// Конвертация результатов в internal format
	rawResults := make([]scanner.Result, 0, len(results))
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

		rawResults = append(rawResults, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	// Аудит портов
	portFindings := audit.EvaluateOpenPorts(rawResults)
	portAudit := make([]contracts.Finding, 0, len(portFindings))
	for _, f := range portFindings {
		portAudit = append(portAudit, contracts.Finding{
			Severity:       f.Severity,
			Host:           f.Host,
			Title:          f.Title,
			Recommendation: f.Recommendation,
		})
	}

	// Risk signatures
	riskFindings := []risksignature.Finding{}
	if db, err := risksignature.LoadDefault(); err == nil {
		riskFindings = risksignature.Evaluate(rawResults, db)
	}
	riskSig := make([]contracts.Finding, 0, len(riskFindings))
	for _, f := range riskFindings {
		riskSig = append(riskSig, contracts.Finding{
			Severity:       f.Severity,
			Host:           f.HostIP,
			Title:          f.Title,
			Recommendation: f.Recommendation,
		})
	}

	// Расчёт индекса безопасности
	severityCounts := map[string]int{}
	for _, f := range portAudit {
		severityCounts[f.Severity]++
	}
	for _, f := range riskSig {
		severityCounts[f.Severity]++
	}
	score := calculateSecurityIndex(severityCounts)

	return &contracts.SecurityReport{
		PortAudit: portAudit,
		RiskSig:   riskSig,
		Score:     score,
	}, nil
}

func calculateSecurityIndex(severityCounts map[string]int) int {
	score := 100
	score -= severityCounts["critical"] * 30
	score -= severityCounts["high"] * 20
	score -= severityCounts["medium"] * 10
	score -= severityCounts["low"] * 5
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
