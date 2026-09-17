package controller

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/network"
)

// ApplyPreset применяет пресет сканирования.
func (c *ScanController) ApplyPreset(mode string) {
	switch mode {
	case "quick":
		c.ui.PortRangeEntry.SetText("22,80,443,445,3389")
		c.ui.TimeoutEntry.SetText("1")
		c.ui.ThreadsEntry.SetText("120")
		c.ui.ScanUDPCheck.SetChecked(false)
		c.ui.ScanBannersCheck.SetChecked(false)
		c.ui.ScanOSActiveCheck.SetChecked(false)
		c.ui.StatusLabel.SetText("Пресет: Быстро (обзор)")
	case "deep":
		c.ui.PortRangeEntry.SetText("1-2000")
		c.ui.TimeoutEntry.SetText("3")
		c.ui.ThreadsEntry.SetText("40")
		c.ui.ScanUDPCheck.SetChecked(true)
		c.ui.ScanBannersCheck.SetChecked(true)
		c.ui.ScanOSActiveCheck.SetChecked(true)
		c.ui.StatusLabel.SetText("Пресет: Глубоко (детальный анализ)")
	default:
		c.ui.PortRangeEntry.SetText("1-1000")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("50")
		c.ui.ScanUDPCheck.SetChecked(false)
		c.ui.ScanBannersCheck.SetChecked(false)
		c.ui.ScanOSActiveCheck.SetChecked(false)
		c.ui.StatusLabel.SetText("Пресет: Баланс")
	}
	if c.app != nil {
		c.app.Preferences().SetString("scan.preset", mode)
	}
	c.RefreshPresetUI()
}

// ApplyRecommendedProfile применяет рекомендованный профиль.
func (c *ScanController) ApplyRecommendedProfile(networkStr string) {
	hosts := 0
	if networkStr != "" {
		if h, err := network.EstimateHostCount(networkStr); err == nil && h > 0 {
			hosts = h
		}
	}
	var profileName string
	switch {
	case hosts >= 2048:
		c.ui.PortRangeEntry.SetText("22,80,443,445,3389")
		c.ui.TimeoutEntry.SetText("1")
		c.ui.ThreadsEntry.SetText("40")
		profileName = "бережный для очень крупной подсети"
	case hosts >= 1024:
		c.ui.PortRangeEntry.SetText("1-1024")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("60")
		profileName = "бережный для крупной подсети"
	case hosts >= 512:
		c.ui.PortRangeEntry.SetText("1-1024,3389")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("80")
		profileName = "сбалансированный для средней подсети"
	default:
		c.ui.PortRangeEntry.SetText("1-2048,3389")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("100")
		profileName = "углубленный для небольшой подсети"
	}
	c.ui.ScanUDPCheck.SetChecked(false)
	c.ui.ScanBannersCheck.SetChecked(false)
	c.ui.ScanOSActiveCheck.SetChecked(false)
	if c.ui.ScanVerboseLogsCheck != nil {
		c.ui.ScanVerboseLogsCheck.SetChecked(false)
	}
	if c.ui.AutoProfileCheck != nil {
		c.ui.AutoProfileCheck.SetChecked(true)
	}
	if c.app != nil {
		c.app.Preferences().SetString("scan.preset", "recommended")
	}
	if hosts > 0 {
		c.ui.StatusLabel.SetText(fmt.Sprintf("Применен рекомендованный профиль (%s), оценка подсети: ~%d хостов", profileName, hosts))
	} else {
		c.ui.StatusLabel.SetText(fmt.Sprintf("Применен рекомендованный профиль (%s)", profileName))
	}
	badgeClass := c.recommendedBadgeClassForHosts(hosts)
	if c.ui.RecommendedBadge != nil {
		c.ui.RecommendedBadge.Text = c.recommendedBadgeText(profileName, badgeClass)
		c.ui.RecommendedBadge.Color = &color.RGBA{R: 55, G: 130, B: 200, A: 255}
		c.ui.RecommendedBadge.Refresh()
	}
	c.RefreshPresetUI()
}

// recommendedBadgeClassForHosts возвращает класс бейджа.
func (c *ScanController) recommendedBadgeClassForHosts(hosts int) string {
	switch {
	case hosts >= 2048:
		return "red"
	case hosts >= 1024:
		return "orange"
	case hosts >= 512:
		return "yellow"
	default:
		return "green"
	}
}

// recommendedBadgeText возвращает текст бейджа.
func (c *ScanController) recommendedBadgeText(profileName, badgeClass string) string {
	return fmt.Sprintf("%s (%s)", profileName, badgeClass)
}

// RefreshPresetUI обновляет состояние виджетов после применения пресета.
func (c *ScanController) RefreshPresetUI() {
	if c.ui == nil {
		return
	}
	c.ui.PortRangeEntry.Refresh()
	c.ui.TimeoutEntry.Refresh()
	c.ui.ThreadsEntry.Refresh()
	c.ui.ScanUDPCheck.Refresh()
	c.ui.StatusLabel.Refresh()
}

// autoScanProfile применяет автоматический профиль сканирования
func autoScanProfile(networkStr string, portRange string, threads int) (string, int, string) {
	portRange = strings.TrimSpace(portRange)
	if threads < 1 {
		threads = 1
	}
	hosts, err := network.EstimateHostCount(strings.TrimSpace(networkStr))
	if err != nil || hosts < 256 {
		return portRange, threads, ""
	}

	portCount := 0
	if portRange != "" {
		if ports, perr := network.ParsePortRange(portRange); perr == nil {
			portCount = len(ports)
		}
	}

	newPortRange := portRange
	newThreads := threads
	msg := ""

	switch {
	case hosts >= 2048:
		if portCount > 512 {
			newPortRange = "1-512"
		}
		if newThreads > 24 {
			newThreads = 24
		}
	case hosts >= 1024:
		if portCount > 1024 {
			newPortRange = "1-1024"
		}
		if newThreads > 40 {
			newThreads = 40
		}
	case hosts >= 512:
		if portCount > 2000 {
			newPortRange = "1-2000"
		}
		if newThreads > 64 {
			newThreads = 64
		}
	default:
		if portCount > 10000 {
			newPortRange = "1-4000"
		}
		if newThreads > 96 {
			newThreads = 96
		}
	}

	if newPortRange != portRange || newThreads != threads {
		parts := make([]string, 0, 2)
		if newPortRange != portRange {
			parts = append(parts, fmt.Sprintf("ports: %s -> %s", portRange, newPortRange))
		}
		if newThreads != threads {
			parts = append(parts, fmt.Sprintf("threads: %d -> %d", threads, newThreads))
		}
		msg = fmt.Sprintf("Автопрофиль: подсеть ~%d хостов, %s", hosts, strings.Join(parts, ", "))
	}
	return newPortRange, newThreads, msg
}

// estimateScanUITimeout оценивает таймаут сканирования
func estimateScanUITimeout(networkStr, portRange, timeoutText, threadsText string, scanTCP, scanUDP bool) time.Duration {
	base := 300 * time.Second

	timeoutSec := 2
	if v, err := strconv.Atoi(strings.TrimSpace(timeoutText)); err == nil && v > 0 {
		timeoutSec = v
	}
	threads := 50
	if v, err := strconv.Atoi(strings.TrimSpace(threadsText)); err == nil && v > 0 {
		threads = v
	}
	if threads < 1 {
		threads = 1
	}

	hosts := 256
	if h, err := network.EstimateHostCount(strings.TrimSpace(networkStr)); err == nil && h > 0 {
		hosts = h
	}

	ports := 0
	if scanTCP {
		effectiveRange := strings.TrimSpace(portRange)
		if effectiveRange == "" {
			effectiveRange = "1-65535"
		}
		if parsed, err := network.ParsePortRange(effectiveRange); err == nil {
			ports = len(parsed)
		}
	}
	if scanUDP {
		ports += 9
	}
	if ports == 0 {
		ports = 1
	}

	workUnits := hosts * ports
	estimatedSec := (workUnits * timeoutSec) / threads
	estimated := time.Duration(estimatedSec) * time.Second

	estimated = estimated / 4
	estimated += 90 * time.Second

	if estimated < base {
		return base
	}
	maxTimeout := 45 * time.Minute
	if estimated > maxTimeout {
		return maxTimeout
	}
	return estimated
}

// formatDurationMMSS форматирует длительность в MM:SS
func formatDurationMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Round(time.Second).Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}
