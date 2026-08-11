package ports

import (
	"testing"
)

// ============================================================================
// formatIANAServiceName — покрытие всех веток (178–225)
// ============================================================================

func TestFormatIANAServiceName_Acronym(t *testing.T) {
	if got := formatIANAServiceName("ssh"); got != "SSH" {
		t.Errorf("formatIANAServiceName(ssh) = %q, want SSH", got)
	}
}

func TestFormatIANAServiceName_Hyphenated(t *testing.T) {
	got := formatIANAServiceName("http-alt")
	if got != "HTTP-Alt" {
		t.Errorf("formatIANAServiceName(http-alt) = %q, want HTTP-Alt", got)
	}
}

func TestFormatIANAServiceName_PostgreSQL(t *testing.T) {
	if got := formatIANAServiceName("postgresql"); got != "PostgreSQL" {
		t.Errorf("formatIANAServiceName(postgresql) = %q, want PostgreSQL", got)
	}
}

func TestFormatIANAServiceName_MongoDB(t *testing.T) {
	if got := formatIANAServiceName("mongodb"); got != "MongoDB" {
		t.Errorf("formatIANAServiceName(mongodb) = %q, want MongoDB", got)
	}
}

func TestFormatIANAServiceName_Empty(t *testing.T) {
	if got := formatIANAServiceName(""); got != "" {
		t.Errorf("formatIANAServiceName(empty) = %q, want empty", got)
	}
}

func TestFormatIANAServiceName_SingleWord(t *testing.T) {
	got := formatIANAServiceName("mysql")
	if got == "" {
		t.Error("formatIANAServiceName(mysql) should not be empty")
	}
}

func TestFormatIANAServiceName_NoAcronym(t *testing.T) {
	// Служба без акронима в segmentAcronyms
	got := formatIANAServiceName("some-unknown-service")
	if got == "" {
		t.Error("formatIANAServiceName should return non-empty for known service")
	}
}

func TestFormatIANAServiceName_AllCaps(t *testing.T) {
	got := formatIANAServiceName("FTP-DATA")
	if got != "FTP-Data" {
		t.Errorf("formatIANAServiceName(FTP-DATA) = %q, want FTP-Data", got)
	}
}

func TestFormatIANAServiceName_MixedCase(t *testing.T) {
	got := formatIANAServiceName("HTTP-Proxy")
	if got != "HTTP-Proxy" {
		t.Errorf("formatIANAServiceName(HTTP-Proxy) = %q, want HTTP-Proxy", got)
	}
}

func TestFormatIANAServiceName_SinglePartNoHyphen(t *testing.T) {
	got := formatIANAServiceName("radius")
	if got == "" {
		t.Error("formatIANAServiceName(radius) should not be empty")
	}
}

// ============================================================================
// LookupServiceName — покрытие с overrides и IANA
// ============================================================================

func TestLookupServiceName_Overrides(t *testing.T) {
	overrides := map[int]string{
		20:   "FTP-Data",
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		67:   "DHCP",
		68:   "DHCP-Client",
		69:   "TFTP",
		80:   "HTTP",
		88:   "Kerberos",
		110:  "POP3",
		123:  "NTP",
		135:  "MSRPC",
		139:  "NetBIOS-SSN",
		143:  "IMAP",
		161:  "SNMP",
		162:  "SNMP-Trap",
		389:  "LDAP",
		443:  "HTTPS",
		445:  "SMB",
		465:  "SMTPS",
		514:  "Syslog",
		587:  "SMTP-Submission",
		636:  "LDAPS",
		873:  "RSync",
		993:  "IMAPS",
		995:  "POP3S",
		1194: "OpenVPN",
		1433: "MSSQL",
		1723: "PPTP",
		2049: "NFS",
		3000: "Node.js",
		3306: "MySQL",
		3389: "RDP",
		5000: "Flask",
		5060: "SIP",
		5061: "SIPS",
		5432: "PostgreSQL",
		5900: "VNC",
		5901: "VNC-1",
		5902: "VNC-2",
		6379: "Redis",
		8000: "HTTP-Alt",
		8001: "HTTP-Alt",
		8008: "HTTP-Alt",
		8080: "HTTP-Proxy",
		8081: "HTTP-Proxy-Alt",
		8443: "HTTPS-Alt",
		8880: "HTTP-Alt",
		8888: "HTTP-Alt",
		9000: "SonarQube",
		9090: "Prometheus",
		27015: "Steam",
		25565: "Minecraft",
		27017: "MongoDB",
	}

	for port, want := range overrides {
		got := LookupServiceName(port)
		if got != want {
			t.Errorf("LookupServiceName(%d) = %q, want %q", port, got, want)
		}
	}
}

func TestLookupServiceName_Unknown(t *testing.T) {
	got := LookupServiceName(99999)
	if got != "Unknown" {
		t.Errorf("LookupServiceName(99999) = %q, want Unknown", got)
	}
}

func TestLookupServiceName_Zero(t *testing.T) {
	got := LookupServiceName(0)
	if got == "" {
		t.Error("LookupServiceName(0) should not be empty")
	}
}

func TestLookupServiceName_Negative(t *testing.T) {
	got := LookupServiceName(-1)
	if got == "" {
		t.Error("LookupServiceName(-1) should not be empty")
	}
}

// ============================================================================
// Description — TCP и UDP ветки
// ============================================================================

func TestDescription_TCP(t *testing.T) {
	d := Description(22)
	if d == "" {
		t.Error("Description(22) should not be empty")
	}
}

func TestDescription_Unknown(t *testing.T) {
	d := Description(99999)
	if d != "" {
		t.Errorf("Description(99999) = %q, want empty", d)
	}
}

func TestDescription_Zero(t *testing.T) {
	d := Description(0)
	// Описание может быть пустым или нет — главное без паники
	_ = d
}

// ============================================================================
// ProtocolLabel — обёртка над LookupServiceName
// ============================================================================

func TestProtocolLabel_Known(t *testing.T) {
	got := ProtocolLabel(80)
	if got == "" {
		t.Error("ProtocolLabel(80) should not be empty")
	}
}

func TestProtocolLabel_Unknown(t *testing.T) {
	got := ProtocolLabel(99999)
	if got != "" {
		t.Errorf("ProtocolLabel(99999) = %q, want empty", got)
	}
}

func TestProtocolLabel_Zero(t *testing.T) {
	got := ProtocolLabel(0)
	// Без паники
	_ = got
}

// ============================================================================
// segmentAcronyms — покрытие всех ключей
// ============================================================================

func TestSegmentAcronyms_ContainsAllExpected(t *testing.T) {
	expected := []string{
		"ftp", "ssh", "http", "https", "smtp",
		"dns", "pop3", "imap", "tcp", "udp",
		"snmp", "ldap", "nfs", "dhcp", "tftp",
		"ntp", "mysql", "mongodb", "redis",
		"telnet", "ssl", "tls", "smb", "sql",
		"rdp", "vnc", "rpc", "ms", "wbt", "ssn",
		"alt", "data", "trap", "postgresql", "mongo",
	}
	for _, key := range expected {
		if _, ok := segmentAcronyms[key]; !ok {
			t.Errorf("segmentAcronyms missing key: %q", key)
		}
	}
}

func TestSegmentAcronyms_ValueCount(t *testing.T) {
	if len(segmentAcronyms) < 35 {
		t.Errorf("expected at least 35 acronyms, got %d", len(segmentAcronyms))
	}
}

// ============================================================================
// portLabelOverrides — покрытие
// ============================================================================

func TestPortLabelOverrides_Count(t *testing.T) {
	if len(portLabelOverrides) < 40 {
		t.Errorf("expected at least 40 overrides, got %d", len(portLabelOverrides))
	}
}

func TestPortLabelOverrides_ContainsCriticalPorts(t *testing.T) {
	critical := []int{22, 80, 443, 3306, 5432, 8080}
	for _, port := range critical {
		if _, ok := portLabelOverrides[port]; !ok {
			t.Errorf("portLabelOverrides missing critical port: %d", port)
		}
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkLookupServiceName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LookupServiceName(80)
	}
}

func BenchmarkFormatIANAServiceName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatIANAServiceName("http-alt")
	}
}

func BenchmarkProtocolLabel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ProtocolLabel(443)
	}
}
