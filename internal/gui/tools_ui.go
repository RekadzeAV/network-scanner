package gui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/nettools"
)

// runDeviceControlTool запускает управление устройством.
//
// P0-3: reboot — необратимое действие, поэтому вынесено за модальное
// подтверждение пользователя. Без окна подтверждения (headless) действие не
// выполняется.
func (a *App) runDeviceControlTool(action string) {
	if a == nil || a.toolsDeviceTargetEntry == nil {
		return
	}
	target := strings.TrimSpace(a.toolsDeviceTargetEntry.Text)
	if target == "" {
		dialog.ShowInformation("Device Control", "Введите Device target URL", a.myWindow)
		return
	}
	vendor := devicecontrol.VendorGenericHTTP
	if a.toolsDeviceVendorEntry != nil && strings.TrimSpace(a.toolsDeviceVendorEntry.Selected) != "" {
		vendor = strings.TrimSpace(a.toolsDeviceVendorEntry.Selected)
	}
	username := ""
	if a.toolsDeviceUserEntry != nil {
		username = strings.TrimSpace(a.toolsDeviceUserEntry.Text)
	}
	password := ""
	if a.toolsDevicePassEntry != nil {
		password = strings.TrimSpace(a.toolsDevicePassEntry.Text)
	}
	timeoutSec := 10
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}

	run := func() {
		a.runDeviceControlRequest(action, target, vendor, username, password, timeoutSec)
	}
	if action == devicecontrol.ActionReboot {
		confirmDeviceReboot(a.myWindow, target, run)
		return
	}
	run()
}

// confirmDeviceReboot показывает диалог подтверждения перезагрузки устройства.
func confirmDeviceReboot(win fyne.Window, target string, onConfirm func()) {
	if win == nil || onConfirm == nil {
		return
	}
	message := "Перезагрузить устройство " + target +
		"?\n\nДействие необратимо: сессия устройства и подключённые клиенты будут прерваны."
	dialog.ShowConfirm("Подтверждение перезагрузки", message, func(ok bool) {
		if ok {
			onConfirm()
		}
	}, win)
}

// runDeviceControlRequest выполняет запрос device-control и пишет audit-запись.
func (a *App) runDeviceControlRequest(action, target, vendor, username, password string, timeoutSec int) {
	consent := ""
	if action == devicecontrol.ActionReboot {
		consent = devicecontrol.ConsentToken
	}
	a.runToolOperation("Device Control", fmt.Sprintf("Выполняется device action `%s`...", action), func(ctx context.Context) (string, error) {
		req := devicecontrol.Request{
			Action:    action,
			TargetURL: target,
			Vendor:    vendor,
			Username:  username,
			Password:  password,
			Timeout:   time.Duration(timeoutSec) * time.Second,
			Consent:   consent,
		}
		res, err := devicecontrol.Execute(ctx, req)
		entry := devicecontrol.AuditEntry{
			Action:    req.Action,
			TargetURL: req.TargetURL,
			Vendor:    req.Vendor,
			Success:   err == nil && res.Success,
			Message:   strings.TrimSpace(res.Message),
		}
		if err != nil && strings.TrimSpace(entry.Message) == "" {
			entry.Message = err.Error()
		}
		auditPath := deviceAuditLogPath()
		_ = devicecontrol.AppendAudit(auditPath, entry)
		if err != nil {
			return fmt.Sprintf("### Device Control\n\nОшибка: `%v`\n\n- audit: `%s`", err, auditPath), err
		}
		var sb strings.Builder
		sb.WriteString("### Device Control\n\n")
		sb.WriteString(fmt.Sprintf("- action: `%s`\n", strings.TrimSpace(res.Action)))
		sb.WriteString(fmt.Sprintf("- target: `%s`\n", strings.TrimSpace(res.TargetURL)))
		sb.WriteString(fmt.Sprintf("- status_code: `%d`\n", res.StatusCode))
		sb.WriteString(fmt.Sprintf("- result: `%s`\n", strings.TrimSpace(res.Message)))
		sb.WriteString(fmt.Sprintf("- audit: `%s`\n", auditPath))
		return sb.String(), nil
	})
}

// --- Инструменты: Ping & Traceroute & DNS ---

// runPingTool запускает ping.
func (a *App) runPingTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	count := 4
	if a.toolsPingCountEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsPingCountEnt.Text)); err == nil && v > 0 {
			count = v
		}
	}
	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("Ping", "Выполняется `ping`...", func(ctx context.Context) (string, error) {
		res, err := nettools.RunPingStructured(ctx, host, count, timeout)
		if err != nil {
			return fmt.Sprintf("### Ping\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("### Ping: `%s`\n\n", host))
		sb.WriteString(fmt.Sprintf("- Count: `%d`\n", count))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		sb.WriteString(fmt.Sprintf("- Sent: `%d`\n", res.Stats.Sent))
		sb.WriteString(fmt.Sprintf("- Received: `%d`\n", res.Stats.Received))
		sb.WriteString(fmt.Sprintf("- Loss: `%.1f%%`\n", res.Stats.PacketLoss))
		if res.Stats.RTTAvg > 0 {
			sb.WriteString(fmt.Sprintf("- RTT min/avg/max: `%s / %s / %s`\n", res.Stats.RTTMin, res.Stats.RTTAvg, res.Stats.RTTMax))
		}
		sb.WriteString("\n#### Raw output\n\n```\n")
		sb.WriteString(res.RawOutput)
		sb.WriteString("\n```")
		return sb.String(), nil
	})
}

// runTracerouteTool запускает traceroute.
func (a *App) runTracerouteTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	maxHops := 30
	if a.toolsTraceHopsEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTraceHopsEnt.Text)); err == nil && v > 0 {
			maxHops = v
		}
	}
	if maxHops <= 0 {
		maxHops = 30
	}
	if maxHops > 64 {
		maxHops = 64
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("Traceroute", "Выполняется `traceroute`...", func(ctx context.Context) (string, error) {
		res, err := nettools.RunTracerouteStructuredWithMaxHops(ctx, host, timeout, maxHops)
		if err != nil {
			return fmt.Sprintf("### Traceroute\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("### Traceroute: `%s`\n\n", host))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		sb.WriteString(fmt.Sprintf("- Max hops: `%d`\n", maxHops))
		if len(res.Hops) == 0 {
			sb.WriteString("- Hop-данные не распознаны.\n")
		}
		for _, hop := range res.Hops {
			addr := strings.TrimSpace(hop.Address)
			if addr == "" {
				addr = "*"
			}
			if hop.Measurements > 0 {
				sb.WriteString(fmt.Sprintf("- hop `%d`: `%s` (min/avg/max `%s/%s/%s`)\n", hop.Index, addr, hop.RTTMin, hop.RTTAvg, hop.RTTMax))
			} else {
				sb.WriteString(fmt.Sprintf("- hop `%d`: `%s`\n", hop.Index, addr))
			}
		}
		sb.WriteString("\n#### Raw output\n\n```\n")
		sb.WriteString(res.RawOutput)
		sb.WriteString("\n```")
		return sb.String(), nil
	})
}

// runDNSTool запускает DNS lookup.
func (a *App) runDNSTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	resolver := ""
	if a.toolsDNSResolverEnt != nil {
		resolver = strings.TrimSpace(a.toolsDNSResolverEnt.Text)
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("DNS", "Выполняется DNS lookup...", func(ctx context.Context) (string, error) {
		lookupCtx, cancel := context.WithTimeout(ctx, timeout)
		res, err := nettools.LookupDNSWithResolver(lookupCtx, host, resolver)
		cancel()
		if err != nil {
			return fmt.Sprintf("### DNS\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString("### DNS lookup\n\n")
		sb.WriteString(fmt.Sprintf("- Запрос: `%s`\n", host))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		if resolver != "" {
			sb.WriteString(fmt.Sprintf("- Resolver: `%s`\n", resolver))
		}
		if len(res.ForwardIPs) > 0 {
			sb.WriteString("- A/AAAA:\n")
			for _, ip := range res.ForwardIPs {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", strings.TrimSpace(ip)))
			}
		}
		if len(res.ReverseNames) > 0 {
			sb.WriteString("- PTR:\n")
			for _, name := range res.ReverseNames {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", strings.TrimSpace(name)))
			}
		}
		if len(res.ForwardIPs) == 0 && len(res.ReverseNames) == 0 {
			sb.WriteString("- Ответ пустой.\n")
		}
		return sb.String(), nil
	})
}
