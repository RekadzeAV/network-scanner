package topology

import (
	"os"
	"strings"
	"testing"
)

func TestWriteTextBasic(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
				SNMPEnabled: true,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				Evidence:   "lldp_match",
				SourcePort: &Port{Index: 1, Name: "Gi0/1"},
			},
		},
	}

	var buf strings.Builder
	if err := topo.WriteText(&buf); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	text := buf.String()

	// Проверяем основные секции
	if !strings.Contains(text, "Network Topology Report") {
		t.Error("text should contain report header")
	}
	if !strings.Contains(text, "── DEVICES ──") {
		t.Error("text should contain devices section")
	}
	if !strings.Contains(text, "── LINKS ──") {
		t.Error("text should contain links section")
	}
	if !strings.Contains(text, "── SUMMARY ──") {
		t.Error("text should contain summary section")
	}

	// Проверяем данные устройств
	if !strings.Contains(text, "switch1") {
		t.Error("text should contain switch1 hostname")
	}
	if !strings.Contains(text, "host1") {
		t.Error("text should contain host1 hostname")
	}
	if !strings.Contains(text, "192.168.1.1") {
		t.Error("text should contain IP addresses")
	}

	// Проверяем данные связей
	if !strings.Contains(text, "Gi0/1") {
		t.Error("text should contain port name")
	}
	if !strings.Contains(text, "lldp") {
		t.Error("text should contain source type")
	}
	if !strings.Contains(text, "high") {
		t.Error("text should contain confidence level")
	}

	// Проверяем summary
	if !strings.Contains(text, "Devices: 2") {
		t.Error("text should show 2 devices")
	}
	if !strings.Contains(text, "Links:   1") {
		t.Error("text should show 1 link")
	}
}

func TestWriteTextWithSNMP(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:          "192.168.1.1",
				MAC:         "aa:bb:cc:dd:ee:01",
				Hostname:    "switch1",
				Type:        DeviceTypeSwitch,
				SNMPEnabled: true,
				LldpNeighbors: []*LldpNeighbor{
					{
						LocalIfIndex:    1,
						RemoteChassisID: "aa:bb:cc:dd:ee:02",
						RemotePortID:    "eth0",
						RemoteSysName:   "host1",
					},
				},
				MacTable: map[string]int{
					"aa:bb:cc:dd:ee:02": 1,
				},
			},
		},
		Links: []Link{},
	}

	var buf strings.Builder
	if err := topo.WriteText(&buf); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	text := buf.String()

	if !strings.Contains(text, "LLDP Neighbors:") {
		t.Error("text should contain LLDP neighbors section")
	}
	if !strings.Contains(text, "MAC Table:") {
		t.Error("text should contain MAC table section")
	}
	if !strings.Contains(text, "chassis=aa:bb:cc:dd:ee:02") {
		t.Error("text should contain chassis ID")
	}
}

func TestWriteTextNilTopology(t *testing.T) {
	var topo *Topology
	var buf strings.Builder
	err := topo.WriteText(&buf)
	if err == nil {
		t.Error("WriteText on nil topology should return error")
	}
}

func TestSaveAsText(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Hostname: "test-host"},
		},
		Links: []Link{},
	}

	tmpFile := t.TempDir() + "/topology.txt"
	err := topo.SaveAsText(tmpFile)
	if err != nil {
		t.Fatalf("SaveAsText error: %v", err)
	}

	// Проверяем, что файл создан и содержит данные
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "test-host") {
		t.Error("saved file should contain hostname")
	}
	if !strings.Contains(text, "Network Topology Report") {
		t.Error("saved file should contain report header")
	}
}

func TestSaveAsTextValidationFailure(t *testing.T) {
	// Топология с невалидным устройством (нет ID)
	topo := &Topology{
		Devices: map[string]*Device{
			"key": {IP: "", MAC: "", Hostname: ""}, // нет стабильного идентификатора
		},
		Links: []Link{},
	}

	tmpFile := t.TempDir() + "/topology.txt"
	err := topo.SaveAsText(tmpFile)
	if err == nil {
		t.Error("SaveAsText should fail for invalid topology")
	}
}
