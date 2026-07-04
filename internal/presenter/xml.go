package presenter

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"network-scanner/internal/scanner"
)

// XMLPresenter exports scan results to an XML file.
type XMLPresenter struct{}

// xmlReport represents the full report structure for XML.
type xmlReport struct {
	XMLName     xml.Name  `xml:"nmaprun"`
	StartTime   string    `xml:"scaninfo,attr"`
	GeneratedAt string    `xml:"scaninfo,attr"`
	Hosts       []xmlHost `xml:"host"`
	TotalHosts  int       `xml:"stats,attr"`
	OpenPorts   int       `xml:"stats,attr"`
}

// xmlHost represents a single host in XML.
type xmlHost struct {
	Addresses  []xmlAddress  `xml:"address"`
	Ports      []xmlPort     `xml:"ports>port"`
	Hostnames  []xmlHostname `xml:"hostnames>hostname"`
	OS         []xmlOS       `xml:"os"`
	DeviceType string        `xml:"hostsummary>usagetype,attr,omitempty"`
}

// xmlAddress represents an IP or MAC address.
type xmlAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
	Vendor   string `xml:"vendor,attr,omitempty"`
}

// xmlPort represents a port in XML.
type xmlPort struct {
	Protocol string     `xml:"protocol,attr"`
	PortID   int        `xml:"portid,attr"`
	State    xmlState   `xml:"state"`
	Service  xmlService `xml:"service"`
}

// xmlState represents port state.
type xmlState struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr,omitempty"`
}

// xmlService represents a service running on a port.
type xmlService struct {
	Name    string `xml:"name,attr,omitempty"`
	Version string `xml:"version,attr,omitempty"`
}

// xmlHostname represents a hostname.
type xmlHostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr,omitempty"`
}

// xmlOS represents OS detection info.
type xmlOS struct {
	OSMatch []xmlOSMatch `xml:"osmatch"`
}

// xmlOSMatch represents an OS match.
type xmlOSMatch struct {
	Name     string `xml:"name,attr"`
	Accuracy string `xml:"accuracy,attr"`
}

// DisplayHeader is a no-op for XML presenter.
func (p XMLPresenter) DisplayHeader() {}

// DisplayHost is a no-op for XML presenter.
func (p XMLPresenter) DisplayHost(host scanner.HostResult) { _ = host }

// DisplaySummary is a no-op for XML presenter.
func (p XMLPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	_, _ = totalHosts, openPortsCount
}

// Export saves scan results to an XML file.
func (p XMLPresenter) Export(results []scanner.HostResult, format string) error {
	if format != "xml" {
		return fmt.Errorf("XMLPresenter supports only xml format, got %s", format)
	}

	report := xmlReport{
		XMLName:     xml.Name{Local: "nmaprun"},
		StartTime:   time.Now().Format("2006-01-02T15:04:05"),
		GeneratedAt: time.Now().Format("2006-01-02T15:04:05"),
		Hosts:       make([]xmlHost, 0, len(results)),
		TotalHosts:  len(results),
		OpenPorts:   countOpenPorts(results),
	}

	for _, r := range results {
		host := xmlHost{
			Addresses: []xmlAddress{
				{Addr: r.IP, AddrType: "ipv4"},
			},
			Ports:      make([]xmlPort, 0, len(r.Ports)),
			Hostnames:  make([]xmlHostname, 0),
			DeviceType: r.DeviceType,
		}

		if r.MAC != "" {
			host.Addresses = append(host.Addresses, xmlAddress{
				Addr:     r.MAC,
				AddrType: "mac",
				Vendor:   r.DeviceVendor,
			})
		}

		if r.Hostname != "" {
			host.Hostnames = append(host.Hostnames, xmlHostname{
				Name: r.Hostname,
				Type: "user",
			})
		}

		for _, port := range r.Ports {
			host.Ports = append(host.Ports, xmlPort{
				Protocol: port.Protocol,
				PortID:   port.Port,
				State: xmlState{
					State:  port.State,
					Reason: "syn-ack",
				},
				Service: xmlService{
					Name:    port.Service,
					Version: port.Version,
				},
			})
		}

		if r.GuessOS != "" {
			host.OS = []xmlOS{
				{
					OSMatch: []xmlOSMatch{
						{Name: r.GuessOS, Accuracy: r.GuessOSConfidence},
					},
				},
			}
		}

		report.Hosts = append(report.Hosts, host)
	}

	data, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML declaration
	xmlHeader := []byte(xml.Header)
	data = append(xmlHeader, data...)

	file, err := os.Create("scan-results.xml")
	if err != nil {
		return fmt.Errorf("failed to create XML file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write XML file: %w", err)
	}

	fmt.Println("XML report saved to: scan-results.xml")
	return nil
}
