package alerting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"network-scanner/internal/comparator"
	"network-scanner/internal/scanner"
)

// Severity уровень алерта
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// Rule тип правила алертинга
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        RuleType  `json:"type"`
	Severity    Severity  `json:"severity"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
}

// RuleType тип правила
type RuleType string

const (
	RuleTypeNewHost       RuleType = "new_host"
	RuleTypeNewPort       RuleType = "new_port"
	RuleTypePortClosed    RuleType = "port_closed"
	RuleTypeDeviceRemoved RuleType = "device_removed"
	RuleTypeOSChanged     RuleType = "os_changed"
	RuleTypeHostnameChanged RuleType = "hostname_changed"
)

// Alert предупреждение
type Alert struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Severity  Severity  `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Data      string    `json:"data,omitempty"`
	Host      string    `json:"host,omitempty"`
	Port      int       `json:"port,omitempty"`
}

// Engine движок алертинга
type Engine struct {
	mu       sync.RWMutex
	rules    []Rule
	alerts   []Alert
	logFile  string
	handlers []AlertHandler
}

// AlertHandler обработчик алертов
type AlertHandler interface {
	OnAlert(alert *Alert) error
}

// FileHandler сохраняет алерты в файл
type FileHandler struct {
	Path string
	mu   sync.Mutex
}

// ConsoleHandler выводит алерты в консоль
type ConsoleHandler struct{}

// NewEngine создаёт новый движок алертинга
func NewEngine(logFile string) *Engine {
	return &Engine{
		rules:    defaultRules(),
		alerts:   make([]Alert, 0),
		logFile:  logFile,
		handlers: []AlertHandler{
			&FileHandler{Path: logFile},
			&ConsoleHandler{},
		},
	}
}

// defaultRules возвращает правила по умолчанию
func defaultRules() []Rule {
	return []Rule{
		{
			ID:          "rule-001",
			Name:        "New Host Detected",
			Type:        RuleTypeNewHost,
			Severity:    SeverityMedium,
			Enabled:     true,
			Description: "Alert when a new host is detected in the network",
		},
		{
			ID:          "rule-002",
			Name:        "New Port Opened",
			Type:        RuleTypeNewPort,
			Severity:    SeverityHigh,
			Enabled:     true,
			Description: "Alert when a new open port is detected",
		},
		{
			ID:          "rule-003",
			Name:        "Device Removed",
			Type:        RuleTypeDeviceRemoved,
			Severity:    SeverityHigh,
			Enabled:     true,
			Description: "Alert when a device is no longer present in the network",
		},
		{
			ID:          "rule-004",
			Name:        "Port Closed",
			Type:        RuleTypePortClosed,
			Severity:    SeverityMedium,
			Enabled:     true,
			Description: "Alert when an open port becomes closed",
		},
		{
			ID:          "rule-005",
			Name:        "OS Changed",
			Type:        RuleTypeOSChanged,
			Severity:    SeverityLow,
			Enabled:     true,
			Description: "Alert when device OS fingerprint changes",
		},
		{
			ID:          "rule-006",
			Name:        "Hostname Changed",
			Type:        RuleTypeHostnameChanged,
			Severity:    SeverityLow,
			Enabled:     true,
			Description: "Alert when device hostname changes",
		},
	}
}

// CheckAlerts проверяет изменения и создаёт алерты
func (e *Engine) CheckAlerts(oldHosts, newHosts []scanner.Result) []Alert {
	e.mu.Lock()
	defer e.mu.Unlock()

	comparison := comparator.CompareSnapshots("", "", oldHosts, newHosts)
	alerts := make([]Alert, 0)

	// Проверка новых хостов
	if e.isRuleEnabled(RuleTypeNewHost) {
		for _, host := range comparison.NewHosts {
			alert := e.createAlert(
				"rule-001",
				"New Host Detected",
				SeverityMedium,
				fmt.Sprintf("New host detected: %s (%s)", host.IP, host.Hostname),
				host.IP,
				0,
			)
			alerts = append(alerts, alert)
		}
	}

	// Проверка удалённых хостов
	if e.isRuleEnabled(RuleTypeDeviceRemoved) {
		for _, host := range comparison.RemovedHosts {
			alert := e.createAlert(
				"rule-003",
				"Device Removed",
				SeverityHigh,
				fmt.Sprintf("Device removed: %s (%s)", host.IP, host.Hostname),
				host.IP,
				0,
			)
			alerts = append(alerts, alert)
		}
	}

	// Проверка изменений портов
	if e.isRuleEnabled(RuleTypeNewPort) || e.isRuleEnabled(RuleTypePortClosed) {
		for _, portChange := range comparison.PortChanges {
			var ruleID, ruleName string
			var severity Severity

			if portChange.ChangedFrom == "closed" {
				ruleID = "rule-002"
				ruleName = "New Port Opened"
				severity = SeverityHigh
			} else {
				ruleID = "rule-004"
				ruleName = "Port Closed"
				severity = SeverityMedium
			}

			alert := e.createAlert(
				ruleID,
				ruleName,
				severity,
				fmt.Sprintf("Port %d/%s changed: %s -> %s on %s",
					portChange.Port, portChange.Protocol,
					portChange.ChangedFrom, portChange.ChangedTo,
					portChange.HostIP),
				portChange.HostIP,
				portChange.Port,
			)
			alerts = append(alerts, alert)
		}
	}

	// Проверка изменений хостов
	if e.isRuleEnabled(RuleTypeOSChanged) || e.isRuleEnabled(RuleTypeHostnameChanged) {
		for _, changed := range comparison.ChangedHosts {
			for _, field := range changed.ChangedIn {
				var ruleID, ruleName string
				var severity Severity

				if field == "os" {
					ruleID = "rule-005"
					ruleName = "OS Changed"
					severity = SeverityLow
				} else if field == "hostname" {
					ruleID = "rule-006"
					ruleName = "Hostname Changed"
					severity = SeverityLow
				}

				alert := e.createAlert(
					ruleID,
					ruleName,
					severity,
					fmt.Sprintf("%s on %s: %s", ruleName, changed.IP, field),
					changed.IP,
					0,
				)
				alerts = append(alerts, alert)
			}
		}
	}

	// Сохранение алертов
	e.alerts = append(e.alerts, alerts...)

	// Вызов обработчиков
	for _, handler := range e.handlers {
		for _, alert := range alerts {
			if err := handler.OnAlert(&alert); err != nil {
				fmt.Fprintf(os.Stderr, "Alert handler error: %v\n", err)
			}
		}
	}

	return alerts
}

// createAlert создаёт новое предупреждение
func (e *Engine) createAlert(ruleID, ruleName string, severity Severity, message, host string, port int) Alert {
	return Alert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		RuleID:    ruleID,
		RuleName:  ruleName,
		Severity:  severity,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Host:      host,
		Port:      port,
	}
}

// isRuleEnabled проверяет, включено ли правило
func (e *Engine) isRuleEnabled(ruleType RuleType) bool {
	for _, rule := range e.rules {
		if rule.Type == ruleType && rule.Enabled {
			return true
		}
	}
	return false
}

// GetAlerts возвращает все алерты
func (e *Engine) GetAlerts() []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.alerts
}

// GetAlertsBySeverity возвращает алерты по уровню
func (e *Engine) GetAlertsBySeverity(severity Severity) []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Alert, 0)
	for _, alert := range e.alerts {
		if alert.Severity == severity {
			result = append(result, alert)
		}
	}
	return result
}

// ClearAlerts очищает историю алертов
func (e *Engine) ClearAlerts() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alerts = make([]Alert, 0)
}

// OnAlert реализует AlertHandler для FileHandler
func (h *FileHandler) OnAlert(alert *Alert) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(h.Path), 0o755); err != nil {
		return fmt.Errorf("create alert dir: %w", err)
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("marshal alert: %w", err)
	}

	f, err := os.OpenFile(h.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open alert file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write alert: %w", err)
	}

	return nil
}

// OnAlert реализует AlertHandler для ConsoleHandler
func (h *ConsoleHandler) OnAlert(alert *Alert) error {
	fmt.Printf("[ALERT] [%s] %s: %s", alert.Severity, alert.RuleName, alert.Message)
	if alert.Host != "" {
		fmt.Printf(" (Host: %s", alert.Host)
		if alert.Port != 0 {
			fmt.Printf(":%d", alert.Port)
		}
		fmt.Println(")")
	} else {
		fmt.Println()
	}
	return nil
}
