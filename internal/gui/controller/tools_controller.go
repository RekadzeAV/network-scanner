package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"network-scanner/internal/audit"
	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/gui/errors"
	"network-scanner/internal/nettools"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
	"network-scanner/internal/wol"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ToolsUI предоставляет доступ к виджетам инструментов.
type ToolsUI struct {
	HostEntry           *widget.Entry
	PingCountEnt        *widget.Entry
	TimeoutEnt          *widget.Entry
	TraceHopsEnt        *widget.Entry
	DNSResolverEnt      *widget.Entry
	WOLMacEntry         *widget.Entry
	WOLBcastEntry       *widget.Entry
	WOLIfaceEntry       *widget.Entry
	DeviceTargetEntry   *widget.Entry
	DeviceVendorEntry   *widget.Select
	DeviceUserEntry     *widget.Entry
	DevicePassEntry     *widget.Entry
	AuditMinSeveritySel *widget.Select
	PingBtn             *widget.Button
	TraceBtn            *widget.Button
	DNSBtn              *widget.Button
	WhoisBtn            *widget.Button
	WiFiBtn             *widget.Button
	AuditBtn            *widget.Button
	RiskBtn             *widget.Button
	WOLBtn              *widget.Button
	DeviceStatusBtn     *widget.Button
	DeviceRebootBtn     *widget.Button
	ToolsOutput         *widget.RichText
	OperationsOutput    *widget.RichText
	OperationsSelect    *widget.Select
	OperationsRetryBtn  *widget.Button
	OperationsCancelBtn *widget.Button
	StatusLabel         *widget.Label
}

// ToolsController управляет сетевыми инструментами.
type ToolsController struct {
	app        fyne.App
	ui         *ToolsUI
	errHandler *errors.ErrorHandler
}

// NewToolsController создает контроллер.
func NewToolsController(app fyne.App, ui *ToolsUI) *ToolsController {
	return &ToolsController{
		app:        app,
		ui:         ui,
		errHandler: errors.NewErrorHandler("tools"),
	}
}

// RunPingTool запускает ICMP ping.
func (c *ToolsController) RunPingTool() {
	host, ok := c.withHost()
	if !ok {
		return
	}
	count := parseIntOrDefault(c.ui.PingCountEnt.Text, 4)
	timeout := time.Duration(parseIntOrDefault(c.ui.TimeoutEnt.Text, 60)) * time.Second
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Выполняется ping `%s` (%d пакетов)...", host, count))

	go func() {
		var output string
		var err error

		err = errors.ExecuteWithRetry(context.Background(), func() error {
			var runErr error
			output, runErr = nettools.RunPing(context.Background(), host, count, timeout)
			return runErr
		}, errors.DefaultRetryConfig)

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if err != nil {
				msg := c.errHandler.HandleWithUI(err)
				c.setOutputMarkdown(msg)
				return
			}
			c.setOutputMarkdown(output)
		})
	}()
}

// RunTracerouteTool запускает traceroute.
func (c *ToolsController) RunTracerouteTool() {
	host, ok := c.withHost()
	if !ok {
		return
	}
	maxHops := parseIntOrDefault(c.ui.TraceHopsEnt.Text, 30)
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Выполняется traceroute до `%s`...", host))

	go func() {
		var output string
		var err error

		err = errors.ExecuteWithRetry(context.Background(), func() error {
			var runErr error
			output, runErr = nettools.RunTraceroute(context.Background(), host, time.Duration(maxHops)*time.Second)
			return runErr
		}, errors.DefaultRetryConfig)

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if err != nil {
				msg := c.errHandler.HandleWithUI(err)
				c.setOutputMarkdown(msg)
				return
			}
			c.setOutputMarkdown(output)
		})
	}()
}

// RunDNSTool запускает DNS lookup.
func (c *ToolsController) RunDNSTool() {
	host, ok := c.withHost()
	if !ok {
		return
	}
	resolver := ""
	if c.ui.DNSResolverEnt != nil {
		resolver = strings.TrimSpace(c.ui.DNSResolverEnt.Text)
	}
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Выполняется DNS lookup для `%s`...", host))

	go func() {
		result, err := nettools.LookupDNSWithResolver(context.Background(), host, resolver)
		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if err != nil {
				c.setOutputMarkdown(fmt.Sprintf("Ошибка DNS: %v", err))
				return
			}
			var md strings.Builder
			md.WriteString(fmt.Sprintf("### DNS Lookup: `%s`\n\n", host))
			if len(result.ForwardIPs) > 0 {
				md.WriteString("**Forward:**\n")
				for _, ip := range result.ForwardIPs {
					md.WriteString(fmt.Sprintf("- `%s`\n", ip))
				}
			}
			if len(result.ReverseNames) > 0 {
				md.WriteString("\n**Reverse:**\n")
				for _, name := range result.ReverseNames {
					md.WriteString(fmt.Sprintf("- `%s`\n", name))
				}
			}
			if len(result.ForwardIPs) == 0 && len(result.ReverseNames) == 0 {
				md.WriteString("Нет записей.\n")
			}
			c.setOutputMarkdown(md.String())
		})
	}()
}

// RunWOLTool запускает Wake-on-LAN.
func (c *ToolsController) RunWOLTool() {
	mac := ""
	if c.ui.WOLMacEntry != nil {
		mac = strings.TrimSpace(c.ui.WOLMacEntry.Text)
	}
	if mac == "" {
		dialog.ShowInformation("Wake-on-LAN", "Введите MAC-адрес", c.app.(interface{ CurrentWindow() fyne.Window }).CurrentWindow())
		return
	}
	bcast := ""
	if c.ui.WOLBcastEntry != nil {
		bcast = strings.TrimSpace(c.ui.WOLBcastEntry.Text)
	}
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Отправка WoL-пакета на MAC %s...", mac))

	go func() {
		var wolErr error

		wolErr = errors.ExecuteWithRetry(context.Background(), func() error {
			return wol.SendMagicPacket(mac, bcast)
		}, errors.RetryConfig{
			MaxAttempts:   2,
			BaseDelay:     200 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
		})

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if wolErr != nil {
				msg := c.errHandler.HandleWithUI(wolErr)
				c.setOutputMarkdown(msg)
				return
			}
			c.setOutputMarkdown("WoL-пакет успешно отправлен.")
		})
	}()
}

// RunPortAuditTool запускает аудит открытых портов.
func (c *ToolsController) RunPortAuditTool(results []scanner.Result) {
	severity := "low"
	if c.ui.AuditMinSeveritySel != nil {
		severity = c.ui.AuditMinSeveritySel.Selected
	}
	_, ok := audit.NormalizeSeverity(severity)
	if !ok {
		severity = "low"
	}
	c.setButtonsEnabled(false)
	c.setOutputMarkdown("Выполняется аудит открытых портов...")

	go func() {
		findings := audit.EvaluateOpenPorts(results)
		findings = audit.FilterByMinSeverity(findings, severity)
		report := audit.FormatFindings(findings)
		fyne.Do(func() {
			c.setButtonsEnabled(true)
			c.setOutputMarkdown(report)
		})
	}()
}

// RunRiskSignaturesTool запускает проверку сигнатур рисков.
func (c *ToolsController) RunRiskSignaturesTool(results []scanner.Result) {
	c.setButtonsEnabled(false)
	c.setOutputMarkdown("Выполняется проверка сигнатур рисков...")

	go func() {
		db, err := risksignature.LoadDefault()
		if err != nil {
			fyne.Do(func() {
				c.setButtonsEnabled(true)
				c.setOutputMarkdown(fmt.Sprintf("Ошибка загрузки базы сигнатур: %v", err))
			})
			return
		}
		findings := risksignature.Evaluate(results, db)
		var md strings.Builder
		md.WriteString("### Risk Signature Findings\n\n")
		if len(findings) == 0 {
			md.WriteString("Рисков не обнаружено.\n")
		} else {
			md.WriteString(fmt.Sprintf("Найдено рисков: %d\n\n", len(findings)))
			for _, f := range findings {
				md.WriteString(fmt.Sprintf("- **[%s]** %s: %s\n  - Reason: %s\n", f.Severity, f.HostIP, f.Title, f.Reason))
			}
		}
		fyne.Do(func() {
			c.setButtonsEnabled(true)
			c.setOutputMarkdown(md.String())
		})
	}()
}

// RunDeviceControlTool запускает управление устройством.
func (c *ToolsController) RunDeviceControlTool(action string) {
	target := ""
	if c.ui.DeviceTargetEntry != nil {
		target = strings.TrimSpace(c.ui.DeviceTargetEntry.Text)
	}
	if target == "" {
		dialog.ShowInformation("Device Control", "Введите URL устройства", c.app.(interface{ CurrentWindow() fyne.Window }).CurrentWindow())
		return
	}
	vendor := ""
	if c.ui.DeviceVendorEntry != nil {
		vendor = c.ui.DeviceVendorEntry.Selected
	}
	user := ""
	pass := ""
	if c.ui.DeviceUserEntry != nil {
		user = c.ui.DeviceUserEntry.Text
	}
	if c.ui.DevicePassEntry != nil {
		pass = c.ui.DevicePassEntry.Text
	}
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Выполняется действие `%s` для `%s`...", action, target))

	go func() {
		var resp devicecontrol.Response
		var ctrlErr error

		ctrlErr = errors.ExecuteWithRetry(context.Background(), func() error {
			var runErr error
			resp, runErr = devicecontrol.Execute(context.Background(), devicecontrol.Request{
				Action:    action,
				TargetURL: target,
				Vendor:    vendor,
				Username:  user,
				Password:  pass,
				Timeout:   10 * time.Second,
			})
			return runErr
		}, errors.DefaultRetryConfig)

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if ctrlErr != nil {
				msg := c.errHandler.HandleWithUI(ctrlErr)
				c.setOutputMarkdown(msg)
				return
			}
			c.setOutputMarkdown(fmt.Sprintf("Успех: %s\nСообщение: %s", resp.TargetURL, resp.Message))
		})
	}()
}

// RunWhoisTool запускает whois lookup.
func (c *ToolsController) RunWhoisTool() {
	host, ok := c.withHost()
	if !ok {
		return
	}
	timeoutSec := parseIntOrDefault(c.ui.TimeoutEnt.Text, 60)
	timeout := time.Duration(timeoutSec) * time.Second
	c.setButtonsEnabled(false)
	c.setOutputMarkdown(fmt.Sprintf("Выполняется `whois` для `%s`...", host))

	go func() {
		var output string
		var err error

		err = errors.ExecuteWithRetry(context.Background(), func() error {
			var runErr error
			output, runErr = nettools.RunWhois(context.Background(), host, timeout)
			return runErr
		}, errors.DefaultRetryConfig)

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if err != nil {
				c.setOutputMarkdown(fmt.Sprintf("### Whois: `%s`\n\nОшибка: `%s`", host, nettools.HumanizeToolError(err)))
				return
			}
			var md strings.Builder
			md.WriteString(fmt.Sprintf("### Whois: `%s`\n\n", host))
			md.WriteString(fmt.Sprintf("- Timeout: `%ds`\n\n", timeoutSec))
			md.WriteString("```text\n")
			md.WriteString(output)
			md.WriteString("\n```")
			c.setOutputMarkdown(md.String())
		})
	}()
}

// RunWiFiTool запускает получение Wi-Fi информации.
func (c *ToolsController) RunWiFiTool() {
	timeoutSec := parseIntOrDefault(c.ui.TimeoutEnt.Text, 30)
	timeout := time.Duration(timeoutSec) * time.Second
	c.setButtonsEnabled(false)
	c.setOutputMarkdown("Чтение Wi-Fi информации...")

	go func() {
		var output string
		var err error

		err = errors.ExecuteWithRetry(context.Background(), func() error {
			var runErr error
			output, runErr = nettools.GetWiFiInfo(context.Background(), timeout)
			return runErr
		}, errors.DefaultRetryConfig)

		fyne.Do(func() {
			c.setButtonsEnabled(true)
			if err != nil {
				c.setOutputMarkdown(fmt.Sprintf("### Wi-Fi\n\nОшибка: `%s`", nettools.HumanizeToolError(err)))
				return
			}
			var md strings.Builder
			md.WriteString("### Wi-Fi\n\n")
			md.WriteString(fmt.Sprintf("- Timeout: `%ds`\n\n", timeoutSec))
			md.WriteString("```text\n")
			md.WriteString(output)
			md.WriteString("\n```")
			c.setOutputMarkdown(md.String())
		})
	}()
}

// withHost проверяет, введен ли хост.
func (c *ToolsController) withHost() (string, bool) {
	if c.ui == nil || c.ui.HostEntry == nil {
		return "", false
	}
	host := strings.TrimSpace(c.ui.HostEntry.Text)
	if host == "" {
		dialog.ShowInformation("Инструменты", "Введите хост или IP", c.app.(interface{ CurrentWindow() fyne.Window }).CurrentWindow())
		return "", false
	}
	return host, true
}

// setButtonsEnabled включает/выключает кнопки.
func (c *ToolsController) setButtonsEnabled(enabled bool) {
	if c.ui == nil {
		return
	}
	for _, b := range []*widget.Button{
		c.ui.PingBtn, c.ui.TraceBtn, c.ui.DNSBtn, c.ui.WhoisBtn,
		c.ui.WiFiBtn, c.ui.WOLBtn, c.ui.AuditBtn, c.ui.RiskBtn,
		c.ui.DeviceStatusBtn, c.ui.DeviceRebootBtn,
	} {
		if b == nil {
			continue
		}
		if enabled {
			b.Enable()
		} else {
			b.Disable()
		}
	}
}

// setOutputMarkdown устанавливает вывод.
func (c *ToolsController) setOutputMarkdown(md string) {
	if c.ui == nil || c.ui.ToolsOutput == nil {
		return
	}
	c.ui.ToolsOutput.ParseMarkdown(md)
	c.ui.ToolsOutput.Refresh()
}

func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var v int
	fmt.Sscanf(s, "%d", &v)
	if v <= 0 {
		return def
	}
	return v
}
