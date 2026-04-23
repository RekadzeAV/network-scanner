package nettools

import "testing"

func TestParseTracerouteWindowsStyle(t *testing.T) {
	raw := `Tracing route to 8.8.8.8 over a maximum of 30 hops

  1    <1 ms    <1 ms    <1 ms  192.168.1.1
  2     *        *        *     Request timed out.
  3    12 ms    13 ms    11 ms  10.0.0.1

Trace complete.`

	hops := parseTraceroute(raw)
	if len(hops) != 3 {
		t.Fatalf("expected 3 hops, got %d", len(hops))
	}
	if hops[0].Index != 1 || hops[0].Address != "192.168.1.1" {
		t.Fatalf("unexpected hop1 parsed: %+v", hops[0])
	}
	if !hops[1].IsTimeout {
		t.Fatalf("expected hop2 timeout, got %+v", hops[1])
	}
	if hops[2].Index != 3 || hops[2].Address != "10.0.0.1" || hops[2].Measurements != 3 {
		t.Fatalf("unexpected hop3 parsed: %+v", hops[2])
	}
}
