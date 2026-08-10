package nettools

import (
	"testing"
)

func TestParsePingStatsUnix(t *testing.T) {
	raw := `PING google.com (142.250.74.14): 56 data bytes
64 bytes from 142.250.74.14: icmp_seq=0 ttl=117 time=10.232 ms
64 bytes from 142.250.74.14: icmp_seq=1 ttl=117 time=11.105 ms

--- google.com ping statistics ---
2 packets transmitted, 2 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 10.232/10.668/11.105/0.436 ms`
	stats := parsePingStats(raw, 2)
	if stats.Sent != 2 || stats.Received != 2 {
		t.Fatalf("unexpected send/recv: %+v", stats)
	}
	if stats.PacketLoss != 0 {
		t.Fatalf("unexpected loss: %.2f", stats.PacketLoss)
	}
	if stats.RTTAvg <= 0 {
		t.Fatalf("expected avg RTT > 0, got %v", stats.RTTAvg)
	}
}

func TestParsePingStatsWindows(t *testing.T) {
	raw := `Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=20ms TTL=117
Reply from 8.8.8.8: bytes=32 time=19ms TTL=117

Ping statistics for 8.8.8.8:
    Packets: Sent = 2, Received = 2, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 19ms, Maximum = 20ms, Average = 19ms`
	stats := parsePingStats(raw, 2)
	if stats.Received != 2 || stats.PacketLoss != 0 {
		t.Fatalf("unexpected windows ping stats: %+v", stats)
	}
	if stats.RTTMin <= 0 || stats.RTTMax <= 0 {
		t.Fatalf("expected RTTs parsed, got %+v", stats)
	}
}

func TestParsePingStatsWindowsRU(t *testing.T) {
	raw := `Обмен пакетами с 8.8.8.8 по с 32 байтами данных:
Ответ от 8.8.8.8: число байт=32 время=20мс TTL=117
Ответ от 8.8.8.8: число байт=32 время=19мс TTL=117

Статистика Ping для 8.8.8.8:
    Пакетов: отправлено = 2, получено = 2, потеряно = 0 (0% потерь),
Приблизительное время приема-передачи в мс:
    Минимальное = 19мсек, Максимальное = 20мсек, Среднее = 19мсек`
	stats := parsePingStats(raw, 2)
	if stats.Received != 2 || stats.PacketLoss != 0 {
		t.Fatalf("unexpected windows ru ping stats: %+v", stats)
	}
	if stats.RTTMin <= 0 || stats.RTTMax <= 0 || stats.RTTAvg <= 0 {
		t.Fatalf("expected RU RTTs parsed, got %+v", stats)
	}
}

func TestParseTraceroute(t *testing.T) {
	raw := `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max
 1  192.168.1.1  1.123 ms  0.901 ms  0.843 ms
 2  * * *
 3  10.0.0.1  5.241 ms  5.402 ms  5.390 ms`
	hops := parseTraceroute(raw)
	if len(hops) != 3 {
		t.Fatalf("expected 3 hops, got %d", len(hops))
	}
	if hops[0].Index != 1 || hops[0].Address != "192.168.1.1" || hops[0].Measurements != 3 {
		t.Fatalf("unexpected hop1: %+v", hops[0])
	}
	if !hops[1].IsTimeout {
		t.Fatalf("expected timeout hop on hop2: %+v", hops[1])
	}
	if hops[2].Address != "10.0.0.1" {
		t.Fatalf("unexpected hop3 address: %+v", hops[2])
	}
}

func TestParseTracerouteNoHops(t *testing.T) {
	raw := `traceroute: unknown host example.invalid`
	hops := parseTraceroute(raw)
	if len(hops) != 0 {
		t.Fatalf("expected 0 hops, got %d", len(hops))
	}
}

func TestParsePingStats_Empty(t *testing.T) {
	stats := parsePingStats("", 0)
	if stats.Sent != 0 || stats.Received != 0 {
		t.Fatalf("expected zero stats, got %+v", stats)
	}
}

func TestParsePingStats_CountOverride(t *testing.T) {
	stats := parsePingStats("", 10)
	if stats.Sent != 10 {
		t.Fatalf("expected sent=10, got %d", stats.Sent)
	}
}

func TestParsePingStats_UnixLoss(t *testing.T) {
	raw := `10 packets transmitted, 5 received, 50.0% packet loss`
	stats := parsePingStats(raw, 0)
	if stats.Sent != 10 || stats.Received != 5 || stats.PacketLoss != 50.0 {
		t.Fatalf("unexpected unix loss stats: %+v", stats)
	}
}

func TestParsePingStats_WindowsLossOnly(t *testing.T) {
	raw := `Lost = 3 (30% loss)`
	stats := parsePingStats(raw, 0)
	if stats.Sent != 3 || stats.Received != 0 || stats.PacketLoss != 30 {
		t.Fatalf("unexpected windows loss stats: %+v", stats)
	}
}

func TestParseTraceroute_MixedTimeoutAndRTT(t *testing.T) {
	raw := ` 1  192.168.1.1  1.000 ms
 2  * * *
 3  10.0.0.1  5.000 ms  6.000 ms  7.000 ms`
	hops := parseTraceroute(raw)
	if len(hops) != 3 {
		t.Fatalf("expected 3 hops, got %d", len(hops))
	}
	if !hops[1].IsTimeout {
		t.Error("hop2 should be timeout")
	}
	if hops[2].RTTMin <= 0 || hops[2].RTTMax <= 0 {
		t.Error("hop3 should have RTT values")
	}
}

func TestParseTraceroute_OnlyTimeouts(t *testing.T) {
	raw := ` 1  * * *
 2  * * *`
	hops := parseTraceroute(raw)
	if len(hops) != 2 {
		t.Fatalf("expected 2 hops, got %d", len(hops))
	}
	for i, hop := range hops {
		if !hop.IsTimeout {
			t.Errorf("hop%d should be timeout", i)
		}
	}
}

func TestParseTraceroute_Empty(t *testing.T) {
	hops := parseTraceroute("")
	if len(hops) != 0 {
		t.Fatalf("expected 0 hops, got %d", len(hops))
	}
}

func TestParseTraceroute_NoHopPrefix(t *testing.T) {
	raw := `some random line
another line`
	hops := parseTraceroute(raw)
	if len(hops) != 0 {
		t.Fatalf("expected 0 hops, got %d", len(hops))
	}
}

func TestParseFloatMs_Valid(t *testing.T) {
	d := parseFloatMs("10.5")
	if d != 10500000 { // 10.5ms в наносекундах
		t.Fatalf("expected 10.5ms, got %v", d)
	}
}

func TestParseFloatMs_Zero(t *testing.T) {
	d := parseFloatMs("0")
	if d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestParseFloatMs_Negative(t *testing.T) {
	d := parseFloatMs("-5")
	if d != 0 {
		t.Fatalf("expected 0 for negative, got %v", d)
	}
}

func TestParseFloatMs_Invalid(t *testing.T) {
	d := parseFloatMs("abc")
	if d != 0 {
		t.Fatalf("expected 0 for invalid, got %v", d)
	}
}

func TestParseFloatMs_Empty(t *testing.T) {
	d := parseFloatMs("")
	if d != 0 {
		t.Fatalf("expected 0 for empty, got %v", d)
	}
}
