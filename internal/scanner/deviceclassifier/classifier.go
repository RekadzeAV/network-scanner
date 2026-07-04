package deviceclassifier

import "strings"

const (
	CategoryUnknown        = "Unknown"
	CategoryRouterSwitch   = "Router/Switch"
	CategoryAccessPoint    = "Access Point"
	CategoryPrinter        = "Printer"
	CategoryCamera         = "Camera"
	CategoryNAS            = "NAS"
	CategoryIoT            = "IoT"
	CategoryDesktopLaptop  = "Desktop/Laptop"
	CategoryServer         = "Server"
	CategoryPhoneTablet    = "Phone/Tablet"
)

type Port struct {
	Port     int
	State    string
	Protocol string
}

type Input struct {
	Ports        []Port
	DeviceVendor string
	Hostname     string
}

func Classify(in Input) string {
	open := map[int]bool{}
	for _, p := range in.Ports {
		if strings.EqualFold(strings.TrimSpace(p.State), "open") {
			open[p.Port] = true
		}
	}
	if len(open) == 0 {
		return CategoryUnknown
	}
	vendor := strings.ToLower(strings.TrimSpace(in.DeviceVendor))
	host := strings.ToLower(strings.TrimSpace(in.Hostname))

	if open[515] || open[631] || open[9100] {
		return CategoryPrinter
	}
	if open[554] {
		return CategoryCamera
	}
	if open[548] || open[2049] {
		return CategoryNAS
	}
	if open[22] && open[80] && !open[3306] && !open[5432] {
		return CategoryRouterSwitch
	}
	if open[443] && open[80] && open[22] {
		if containsAny(vendor, "cisco", "netgear", "tp-link", "d-link", "asus", "linksys") || containsAny(host, "router", "gateway") {
			return CategoryRouterSwitch
		}
		return CategoryServer
	}
	if open[161] && (containsAny(host, "switch", "router", "gateway") || containsAny(vendor, "cisco", "netgear", "tp-link")) {
		return CategoryRouterSwitch
	}
	if open[3389] || (open[135] && open[445]) {
		return CategoryDesktopLaptop
	}
	if open[3306] || open[5432] || open[1433] || open[8080] || open[8443] {
		return CategoryServer
	}
	if open[22] {
		return CategoryServer
	}
	if open[80] && open[443] {
		return CategoryServer
	}
	if open[80] || open[443] {
		return CategoryIoT
	}
	if len(open) == 1 {
		return CategoryIoT
	}
	return CategoryUnknown
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
