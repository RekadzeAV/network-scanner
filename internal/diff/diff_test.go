package diff

import (
	"testing"

	"network-scanner/internal/scanner"
)

func TestCompareScanResults_NoChanges(t *testing.T) {
	host := scanner.HostResult{
		IP:         "192.168.1.1",
		Hostname:   "test-host",
		MAC:        "aa:bb:cc:dd:ee:ff",
		DeviceType: "Computer",
	}

	prev := []scanner.HostResult{host}
	curr := []scanner.HostResult{host}

	report := CompareScanResults(prev, curr)

	if report.TotalNew != 0 {
		t.Errorf("Ожидалось 0 новых хостов, получено %d", report.TotalNew)
	}
	if report.TotalGone != 0 {
		t.Errorf("Ожидалось 0 ушедших хостов, получено %d", report.TotalGone)
	}
	if report.TotalChanged != 0 {
		t.Errorf("Ожидалось 0 изменённых хостов, получено %d", report.TotalChanged)
	}
}

func TestCompareScanResults_NewHost(t *testing.T) {
	prev := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
	}
	curr := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}

	report := CompareScanResults(prev, curr)

	if report.TotalNew != 1 {
		t.Errorf("Ожидалось 1 новый хост, получено %d", report.TotalNew)
	}
	if len(report.NewHosts) != 1 || report.NewHosts[0].IP != "192.168.1.2" {
		t.Error("Ожидался новый хост 192.168.1.2")
	}
}

func TestCompareScanResults_GoneHost(t *testing.T) {
	prev := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	curr := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
	}

	report := CompareScanResults(prev, curr)

	if report.TotalGone != 1 {
		t.Errorf("Ожидалось 1 ушедший хост, получено %d", report.TotalGone)
	}
	if len(report.GoneHosts) != 1 || report.GoneHosts[0].IP != "192.168.1.2" {
		t.Error("Ожидался ушедший хост 192.168.1.2")
	}
}

func TestCompareScanResults_ChangedHost(t *testing.T) {
	prev := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "old-name", DeviceType: "Computer"},
	}
	curr := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "new-name", DeviceType: "Server"},
	}

	report := CompareScanResults(prev, curr)

	if report.TotalChanged != 1 {
		t.Errorf("Ожидалось 1 изменённый хост, получено %d", report.TotalChanged)
	}
	if len(report.ChangedHosts) != 1 {
		t.Fatal("Ожидался 1 элемент в ChangedHosts")
	}

	ch := report.ChangedHosts[0]
	if ch.IP != "192.168.1.1" {
		t.Errorf("Ожидался IP 192.168.1.1, получено %s", ch.IP)
	}
	if len(ch.Changes) != 2 {
		t.Errorf("Ожидалось 2 изменения, получено %d", len(ch.Changes))
	}
}

func TestCompareScanResults_Complete(t *testing.T) {
	prev := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1", DeviceType: "Computer"},
		{IP: "192.168.1.2", Hostname: "host2", DeviceType: "Router"},
	}
	curr := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1-new", DeviceType: "Server"}, // changed
		{IP: "192.168.1.3", Hostname: "host3", DeviceType: "Unknown"},     // new
	}

	report := CompareScanResults(prev, curr)

	if report.TotalNew != 1 {
		t.Errorf("Ожидалось 1 новый хост, получено %d", report.TotalNew)
	}
	if report.TotalGone != 1 {
		t.Errorf("Ожидалось 1 ушедший хост, получено %d", report.TotalGone)
	}
	if report.TotalChanged != 1 {
		t.Errorf("Ожидалось 1 изменённый хост, получено %d", report.TotalChanged)
	}
}

func TestCompareScanResults_EmptyPrevious(t *testing.T) {
	prev := []scanner.HostResult{}
	curr := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}

	report := CompareScanResults(prev, curr)

	if report.TotalNew != 2 {
		t.Errorf("Ожидалось 2 новых хоста, получено %d", report.TotalNew)
	}
	if report.TotalGone != 0 {
		t.Errorf("Ожидалось 0 ушедших хостов, получено %d", report.TotalGone)
	}
}

func TestCompareScanResults_EmptyCurrent(t *testing.T) {
	prev := []scanner.HostResult{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	curr := []scanner.HostResult{}

	report := CompareScanResults(prev, curr)

	if report.TotalNew != 0 {
		t.Errorf("Ожидалось 0 новых хостов, получено %d", report.TotalNew)
	}
	if report.TotalGone != 2 {
		t.Errorf("Ожидалось 2 ушедших хоста, получено %d", report.TotalGone)
	}
}

func TestDetectChanges_PortChanges(t *testing.T) {
	prev := scanner.HostResult{
		IP: "192.168.1.1",
		Ports: []scanner.PortInfo{
			{Port: 80, State: "open", Protocol: "tcp"},
			{Port: 443, State: "open", Protocol: "tcp"},
		},
	}
	curr := scanner.HostResult{
		IP: "192.168.1.1",
		Ports: []scanner.PortInfo{
			{Port: 443, State: "open", Protocol: "tcp"},
			{Port: 8080, State: "open", Protocol: "tcp"},
		},
	}

	changes := detectChanges(prev, curr)

	hasChange := false
	for _, c := range changes {
		if c.Field == "OpenPorts" {
			hasChange = true
			if c.Previous != "80,443" {
				t.Errorf("Ожидалось '80,443', получено '%s'", c.Previous)
			}
			if c.Current != "443,8080" {
				t.Errorf("Ожидалось '443,8080', получено '%s'", c.Current)
			}
		}
	}

	if !hasChange {
		t.Error("Ожидалось изменение открытых портов")
	}
}

func TestFormatReport(t *testing.T) {
	report := &DiffReport{
		NewHosts: []scanner.HostResult{
			{IP: "192.168.1.1", Hostname: "new-host"},
		},
		GoneHosts: []scanner.HostResult{
			{IP: "192.168.1.2", Hostname: "gone-host"},
		},
		ChangedHosts: []ChangedHost{
			{
				IP: "192.168.1.3",
				Changes: []Change{
					{Field: "Hostname", Previous: "old", Current: "new"},
				},
			},
		},
		TotalNew:    1,
		TotalGone:   1,
		TotalChanged: 1,
	}

	formatted := report.FormatReport()

	if formatted == "" {
		t.Error("Ожидался непустой отчёт")
	}
}

func TestSortHosts(t *testing.T) {
	hosts := []scanner.HostResult{
		{IP: "192.168.1.3"},
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}

	sortHosts(hosts)

	expected := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for i, ip := range expected {
		if hosts[i].IP != ip {
			t.Errorf("Ожидался IP %s на позиции %d, получено %s", ip, i, hosts[i].IP)
		}
	}
}

func TestPortsToString(t *testing.T) {
	tests := []struct {
		name   string
		ports  map[int]bool
		want   string
	}{
		{
			name:  "empty",
			ports: map[int]bool{},
			want:  "none",
		},
		{
			name: "single",
			ports: map[int]bool{
				80: true,
			},
			want: "80",
		},
		{
			name: "multiple",
			ports: map[int]bool{
				443: true,
				80:  true,
				22:  true,
			},
			want: "22,80,443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portsToString(tt.ports)
			if got != tt.want {
				t.Errorf("portsToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
