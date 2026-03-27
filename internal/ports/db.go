package ports

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed service-names-port-numbers.csv
var serviceCSV []byte

var (
	tcpNames map[int]string
	udpNames map[int]string
	tcpDesc  map[int]string
	udpDesc  map[int]string

	titleEn = cases.Title(language.English)
)

// portLabelOverrides сохраняют прежние удобочитаемые подписи там, где они расходятся с сырыми именами IANA.
var portLabelOverrides = map[int]string{
	20: "FTP-Data",
	21: "FTP",
	22: "SSH",
	23: "Telnet",
	25: "SMTP",
	53: "DNS",
	67: "DHCP",
	68: "DHCP-Client",
	69: "TFTP",
	80: "HTTP",
	88: "Kerberos",
	110: "POP3",
	123: "NTP",
	135: "MSRPC",
	139: "NetBIOS-SSN",
	143: "IMAP",
	161: "SNMP",
	162: "SNMP-Trap",
	389: "LDAP",
	443: "HTTPS",
	445: "SMB",
	465: "SMTPS",
	514: "Syslog",
	587: "SMTP-Submission",
	636: "LDAPS",
	873: "RSync",
	993: "IMAPS",
	995: "POP3S",
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

// segmentAcronyms — типичные сокращения из реестра IANA (сегменты после '-' или целое имя).
var segmentAcronyms = map[string]string{
	"ftp": "FTP", "ssh": "SSH", "http": "HTTP", "https": "HTTPS", "smtp": "SMTP",
	"dns": "DNS", "pop3": "POP3", "imap": "IMAP", "tcp": "TCP", "udp": "UDP",
	"snmp": "SNMP", "ldap": "LDAP", "nfs": "NFS", "dhcp": "DHCP", "tftp": "TFTP",
	"ntp": "NTP", "mysql": "MySQL", "mongodb": "MongoDB", "redis": "Redis",
	"telnet": "Telnet", "ssl": "SSL", "tls": "TLS", "smb": "SMB", "sql": "SQL",
	"rdp": "RDP", "vnc": "VNC", "rpc": "RPC", "ms": "MS", "wbt": "WBT", "ssn": "SSN",
	"alt": "Alt", "data": "Data", "trap": "Trap",
	"postgresql": "PostgreSQL", "mongo": "Mongo",
}

func init() {
	tcpNames = make(map[int]string)
	udpNames = make(map[int]string)
	tcpDesc = make(map[int]string)
	udpDesc = make(map[int]string)

	data := bytes.TrimPrefix(serviceCSV, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(data))
	r.ReuseRecord = true
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return
	}
	for _, rec := range records[1:] {
		if len(rec) < 4 {
			continue
		}
		portStr := strings.TrimSpace(rec[1])
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 0 || port > 65535 {
			continue
		}
		proto := strings.ToLower(strings.TrimSpace(rec[2]))
		svc := strings.TrimSpace(rec[0])
		desc := strings.TrimSpace(rec[3])

		switch proto {
		case "tcp":
			if svc != "" && tcpNames[port] == "" {
				tcpNames[port] = svc
			}
			if desc != "" && tcpDesc[port] == "" {
				tcpDesc[port] = desc
			}
		case "udp":
			if svc != "" && udpNames[port] == "" {
				udpNames[port] = svc
			}
			if desc != "" && udpDesc[port] == "" {
				udpDesc[port] = desc
			}
		}
	}
}

// LookupServiceName возвращает имя службы по номеру порта (приоритет: локальные подписи → TCP → UDP).
func LookupServiceName(port int) string {
	if s, ok := portLabelOverrides[port]; ok {
		return s
	}
	if s, ok := tcpNames[port]; ok && s != "" {
		return formatIANAServiceName(s)
	}
	if s, ok := udpNames[port]; ok && s != "" {
		return formatIANAServiceName(s)
	}
	return "Unknown"
}

// Description возвращает описание из реестра IANA (TCP, иначе UDP).
func Description(port int) string {
	if d, ok := tcpDesc[port]; ok && d != "" {
		return d
	}
	if d, ok := udpDesc[port]; ok && d != "" {
		return d
	}
	return ""
}

// ProtocolLabel — краткое имя для аналитики «протоколы»; пустая строка, если порт не идентифицирован.
func ProtocolLabel(port int) string {
	s := LookupServiceName(port)
	if s == "Unknown" {
		return ""
	}
	return s
}

func formatIANAServiceName(raw string) string {
	if raw == "" {
		return ""
	}
	if full, ok := segmentAcronyms[strings.ToLower(raw)]; ok {
		return full
	}
	parts := strings.Split(raw, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		lp := strings.ToLower(p)
		if full, ok := segmentAcronyms[lp]; ok {
			parts[i] = full
			continue
		}
		rs := []rune(lp)
		rs[0] = unicode.ToUpper(rs[0])
		for j := 1; j < len(rs); j++ {
			rs[j] = unicode.ToLower(rs[j])
		}
		parts[i] = string(rs)
	}
	s := strings.Join(parts, "-")
	// Отдельные имена вроде postgresql одним словом
	if strings.EqualFold(raw, "postgresql") {
		return "PostgreSQL"
	}
	if strings.EqualFold(raw, "mongodb") {
		return "MongoDB"
	}
	// Несколько слов без дефиса в реестре
	if !strings.Contains(raw, "-") && len(raw) > 3 {
		return titleEn.String(strings.ToLower(raw))
	}
	return s
}
