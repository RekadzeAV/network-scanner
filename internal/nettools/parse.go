package nettools

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reWindowsLoss = regexp.MustCompile(`Lost\s*=\s*(\d+)\s*\((\d+)%\s*loss\)`)
	reWindowsRTT  = regexp.MustCompile(`Minimum\s*=\s*(\d+)ms,\s*Maximum\s*=\s*(\d+)ms,\s*Average\s*=\s*(\d+)ms`)
	reWindowsLossRU = regexp.MustCompile(`(?i)потеряно\s*=\s*(\d+)\s*\((\d+)%\s*потерь\)`)
	reWindowsRTTRU  = regexp.MustCompile(`Минимальное\s*=\s*(\d+)мсек,\s*Максимальное\s*=\s*(\d+)мсек,\s*Среднее\s*=\s*(\d+)мсек`)
	reUnixLoss    = regexp.MustCompile(`(\d+)\s+packets transmitted,\s+(\d+)\s+(?:packets )?received,\s+([0-9.]+)%\s+packet loss`)
	reUnixRTT     = regexp.MustCompile(`(?:round-trip|rtt)\s+min/avg/max(?:/[a-z]+)?\s*=\s*([0-9.]+)/([0-9.]+)/([0-9.]+)`)
	reHopPrefix   = regexp.MustCompile(`^\s*(\d+)\s+`)
	reIPv4        = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	reMSFloat     = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*ms`)
)

// PingStats содержит нормализованные метрики ping.
type PingStats struct {
	Sent       int
	Received   int
	PacketLoss float64
	RTTMin     time.Duration
	RTTAvg     time.Duration
	RTTMax     time.Duration
}

// PingResult объединяет сырой вывод и распарсенные метрики.
type PingResult struct {
	RawOutput string
	Stats     PingStats
}

// TracerouteHop содержит данные одного hop.
type TracerouteHop struct {
	Index       int
	Address     string
	RTTMin      time.Duration
	RTTAvg      time.Duration
	RTTMax      time.Duration
	IsTimeout   bool
	RawLine     string
	Measurements int
}

// TracerouteResult объединяет сырой вывод и распарсенные hop-строки.
type TracerouteResult struct {
	RawOutput string
	Hops      []TracerouteHop
}

func parsePingStats(raw string, count int) PingStats {
	stats := PingStats{}
	if count > 0 {
		stats.Sent = count
	}
	for _, line := range strings.Split(raw, "\n") {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		if m := reWindowsLoss.FindStringSubmatch(s); len(m) == 3 {
			lost, _ := strconv.Atoi(m[1])
			lossPct, _ := strconv.ParseFloat(m[2], 64)
			stats.PacketLoss = lossPct
			if stats.Sent == 0 {
				stats.Sent = lost
			}
			if stats.Sent > 0 {
				stats.Received = stats.Sent - lost
			}
			continue
		}
		if m := reWindowsLossRU.FindStringSubmatch(s); len(m) == 3 {
			lost, _ := strconv.Atoi(m[1])
			lossPct, _ := strconv.ParseFloat(m[2], 64)
			stats.PacketLoss = lossPct
			if stats.Sent == 0 {
				stats.Sent = lost
			}
			if stats.Sent > 0 {
				stats.Received = stats.Sent - lost
			}
			continue
		}
		if m := reWindowsRTT.FindStringSubmatch(s); len(m) == 4 {
			minMs, _ := strconv.Atoi(m[1])
			maxMs, _ := strconv.Atoi(m[2])
			avgMs, _ := strconv.Atoi(m[3])
			stats.RTTMin = time.Duration(minMs) * time.Millisecond
			stats.RTTMax = time.Duration(maxMs) * time.Millisecond
			stats.RTTAvg = time.Duration(avgMs) * time.Millisecond
			continue
		}
		if m := reWindowsRTTRU.FindStringSubmatch(s); len(m) == 4 {
			minMs, _ := strconv.Atoi(m[1])
			maxMs, _ := strconv.Atoi(m[2])
			avgMs, _ := strconv.Atoi(m[3])
			stats.RTTMin = time.Duration(minMs) * time.Millisecond
			stats.RTTMax = time.Duration(maxMs) * time.Millisecond
			stats.RTTAvg = time.Duration(avgMs) * time.Millisecond
			continue
		}
		if m := reUnixLoss.FindStringSubmatch(s); len(m) == 4 {
			sent, _ := strconv.Atoi(m[1])
			recv, _ := strconv.Atoi(m[2])
			lossPct, _ := strconv.ParseFloat(m[3], 64)
			stats.Sent = sent
			stats.Received = recv
			stats.PacketLoss = lossPct
			continue
		}
		if m := reUnixRTT.FindStringSubmatch(s); len(m) >= 4 {
			stats.RTTMin = parseFloatMs(m[1])
			stats.RTTAvg = parseFloatMs(m[2])
			stats.RTTMax = parseFloatMs(m[3])
		}
	}
	return stats
}

func parseTraceroute(raw string) []TracerouteHop {
	hops := make([]TracerouteHop, 0)
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		m := reHopPrefix.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		idx, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		hop := TracerouteHop{
			Index:   idx,
			RawLine: trimmed,
		}
		if strings.Contains(trimmed, "*") {
			hop.IsTimeout = true
		}
		if ip := reIPv4.FindString(trimmed); ip != "" {
			hop.Address = ip
		}
		msMatches := reMSFloat.FindAllStringSubmatch(trimmed, -1)
		var sum time.Duration
		for _, mm := range msMatches {
			if len(mm) < 2 {
				continue
			}
			d := parseFloatMs(mm[1])
			if d <= 0 {
				continue
			}
			if hop.Measurements == 0 || d < hop.RTTMin {
				hop.RTTMin = d
			}
			if d > hop.RTTMax {
				hop.RTTMax = d
			}
			hop.Measurements++
			sum += d
		}
		if hop.Measurements > 0 {
			hop.RTTAvg = time.Duration(int64(sum) / int64(hop.Measurements))
		}
		hops = append(hops, hop)
	}
	return hops
}

func parseFloatMs(raw string) time.Duration {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v <= 0 {
		return 0
	}
	return time.Duration(v * float64(time.Millisecond))
}
