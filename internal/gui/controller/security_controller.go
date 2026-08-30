package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"network-scanner/internal/audit"
	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/logger"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
	"network-scanner/internal/wol"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SecurityUI предоставляет доступ к виджетам безопасности и управления устройствами.
type SecurityUI struct {
	AuditMinSeveritySel *widget.Select
	DeviceTargetEntry   *widget.Entry
	DeviceVendorEntry   *widget.Select
	DeviceUserEntry     *widget.Entry
	DevicePassEntry     *widget.Entry
	DeviceStatusBtn     *widget.Button
	DeviceRebootBtn     *widget.Button
	WOLMacEntry         *widget.Entry
	WOLBcastEntry       *widget.Entry
	WOLIfaceEntry       *widget.Entry
	StatusLabel         *widget.Label
	Window              fyne.Window
	AuditResultsView    *widget.Label // Виджет для отображения результатов аудита
}

// SecurityController управляет безопасностью и управлением устройствами.
type SecurityController struct {
	app fyne.App
	ui  *SecurityUI
}

// NewSecurityController создает контроллер безопасности.
func NewSecurityController(app fyne.App, ui *SecurityUI) *SecurityController {
	return &SecurityController{
		app: app,
		ui:  ui,
	}
}

// RunAudit запускает аудит открытых портов с реальными данными.
func (c *SecurityController) RunAudit(scanResults []scanner.Result, minSeverity string) {
	if len(scanResults) == 0 {
		c.setStatus("Нет данных для аудита (выполните сканирование)")
		if c.ui.AuditResultsView != nil {
			c.ui.AuditResultsView.SetText("Нет данных для аудита")
			c.ui.AuditResultsView.Refresh()
		}
		return
	}
	if minSeverity == "" {
		minSeverity = "low"
	}
	_, ok := audit.NormalizeSeverity(minSeverity)
	if !ok {
		minSeverity = "low"
	}
	c.setStatus("Выполняется аудит открытых портов...")
	logger.Log("Аудит портов: severity=%s, хостов=%d", minSeverity, len(scanResults))

	// Реальный вызов audit с передачей реальных результатов
	findings := audit.EvaluateOpenPorts(scanResults)
	findings = audit.FilterByMinSeverity(findings, minSeverity)
	report := audit.FormatFindings(findings)

	c.setStatus(fmt.Sprintf("Аудит завершен. Найдено проблем: %d", len(findings)))
	logger.Log("Аудит завершен. Проблем: %d", len(findings))

	// Выводим отчет в UI
	if c.ui.AuditResultsView != nil {
		c.ui.AuditResultsView.SetText(report)
		c.ui.AuditResultsView.Refresh()
	}
}

// CheckDeviceStatus проверяет статус устройства через HTTP API.
func (c *SecurityController) CheckDeviceStatus() {
	target := ""
	if c.ui.DeviceTargetEntry != nil {
		target = strings.TrimSpace(c.ui.DeviceTargetEntry.Text)
	}
	if target == "" {
		c.setStatus("Введите URL устройства")
		return
	}

	vendor := ""
	if c.ui.DeviceVendorEntry != nil {
		vendor = c.ui.DeviceVendorEntry.Selected
	}

	user := ""
	pass := ""
	if c.ui.DeviceUserEntry != nil {
		user = strings.TrimSpace(c.ui.DeviceUserEntry.Text)
	}
	if c.ui.DevicePassEntry != nil {
		pass = strings.TrimSpace(c.ui.DevicePassEntry.Text)
	}

	c.setStatus("Проверка статуса устройства...")
	logger.Log("Проверка статуса: target=%s, vendor=%s", target, vendor)

	// Реальный вызов devicecontrol.Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := devicecontrol.Request{
		Action:    devicecontrol.ActionStatus,
		TargetURL: target,
		Vendor:    vendor,
		Username:  user,
		Password:  pass,
		Timeout:   10 * time.Second,
	}

	resp, err := devicecontrol.Execute(ctx, req)
	if err != nil {
		c.setStatus(fmt.Sprintf("Ошибка: %v", err))
		logger.LogError(err, "devicecontrol status")
		dialog.ShowError(err, c.ui.Window)
		return
	}

	c.setStatus(fmt.Sprintf("Статус: %s", resp.Message))
	logger.Log("Статус устройства: %s", resp.Message)
}

// RebootDevice перезагружает устройство через HTTP API.
func (c *SecurityController) RebootDevice() {
	target := ""
	if c.ui.DeviceTargetEntry != nil {
		target = strings.TrimSpace(c.ui.DeviceTargetEntry.Text)
	}
	if target == "" {
		c.setStatus("Введите URL устройства")
		return
	}

	vendor := ""
	if c.ui.DeviceVendorEntry != nil {
		vendor = c.ui.DeviceVendorEntry.Selected
	}

	user := ""
	pass := ""
	if c.ui.DeviceUserEntry != nil {
		user = strings.TrimSpace(c.ui.DeviceUserEntry.Text)
	}
	if c.ui.DevicePassEntry != nil {
		pass = strings.TrimSpace(c.ui.DevicePassEntry.Text)
	}

	c.setStatus("Перезагрузка устройства...")
	logger.Log("Перезагрузка: target=%s, vendor=%s", target, vendor)

	// Реальный вызов devicecontrol.Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := devicecontrol.Request{
		Action:    devicecontrol.ActionReboot,
		TargetURL: target,
		Vendor:    vendor,
		Username:  user,
		Password:  pass,
		Timeout:   10 * time.Second,
	}

	resp, err := devicecontrol.Execute(ctx, req)
	if err != nil {
		c.setStatus(fmt.Sprintf("Ошибка: %v", err))
		logger.LogError(err, "devicecontrol reboot")
		dialog.ShowError(err, c.ui.Window)
		return
	}

	c.setStatus(fmt.Sprintf("Перезагрузка: %s", resp.Message))
	logger.Log("Перезагрузка устройства: %s", resp.Message)
}

// WakeOnLAN отправляет Wake-on-LAN пакет.
func (c *SecurityController) WakeOnLAN() {
	mac := ""
	if c.ui.WOLMacEntry != nil {
		mac = strings.TrimSpace(c.ui.WOLMacEntry.Text)
	}
	if mac == "" {
		c.setStatus("Введите MAC-адрес для WoL")
		return
	}

	bcast := ""
	if c.ui.WOLBcastEntry != nil {
		bcast = strings.TrimSpace(c.ui.WOLBcastEntry.Text)
	}

	iface := ""
	if c.ui.WOLIfaceEntry != nil {
		iface = strings.TrimSpace(c.ui.WOLIfaceEntry.Text)
	}

	c.setStatus("Отправка WoL-пакета...")
	logger.Log("WoL: MAC=%s, bcast=%s, iface=%s", mac, bcast, iface)

	if err := wol.SendMagicPacket(mac, bcast); err != nil {
		c.setStatus(fmt.Sprintf("Ошибка WoL: %v", err))
		logger.LogError(err, "Wake-on-LAN")
		dialog.ShowError(err, c.ui.Window)
		return
	}

	c.setStatus("WoL-пакет успешно отправлен на MAC: " + mac)
	logger.Log("WoL отправлен на MAC: %s", mac)
}

// RunRiskSignatures запускает проверку сигнатур уязвимостей с реальными данными.
func (c *SecurityController) RunRiskSignatures(scanResults []scanner.Result) {
	if len(scanResults) == 0 {
		c.setStatus("Нет данных для проверки (выполните сканирование)")
		return
	}

	c.setStatus("Проверка сигнатур уязвимостей...")
	logger.Log("Запуск проверки сигнатур на %d хостах", len(scanResults))

	// Реальный вызов risksignature
	db, err := risksignature.LoadDefault()
	if err != nil {
		c.setStatus(fmt.Sprintf("Ошибка загрузки базы сигнатур: %v", err))
		logger.LogError(err, "risksignature load")
		return
	}

	findings := risksignature.Evaluate(scanResults, db)
	c.setStatus(fmt.Sprintf("Найдено проблем: %d", len(findings)))
	logger.Log("Проверка сигнатур завершена. Проблем: %d", len(findings))
}

// setStatus устанавливает статус в UI.
func (c *SecurityController) setStatus(msg string) {
	if c.ui != nil && c.ui.StatusLabel != nil {
		c.ui.StatusLabel.SetText(msg)
		c.ui.StatusLabel.Refresh()
	}
}
