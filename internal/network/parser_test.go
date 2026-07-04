package network

import (
	"os"
	"testing"
)

func TestParseTargetsFromFile_SingleIP(t *testing.T) {
	content := "192.168.1.1\n192.168.1.2\n"
	tmpFile := createTempFile(t, content)

	ips, err := ParseTargetsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseTargetsFromFile() error = %v", err)
	}

	if len(ips) != 2 {
		t.Errorf("Ожидалось 2 IP, получено %d", len(ips))
	}

	expected := []string{"192.168.1.1", "192.168.1.2"}
	for i, ip := range expected {
		if ips[i] != ip {
			t.Errorf("Ожидался IP %s на позиции %d, получено %s", ip, i, ips[i])
		}
	}
}

func TestParseTargetsFromFile_CIDR(t *testing.T) {
	content := "192.168.1.0/30\n"
	tmpFile := createTempFile(t, content)

	ips, err := ParseTargetsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseTargetsFromFile() error = %v", err)
	}

	// /30 subnet has 4 addresses, but ParseNetworkRange excludes network and broadcast
	// So only 2 usable host addresses remain
	if len(ips) != 2 {
		t.Errorf("Ожидалось 2 IP (usable hosts из /30), получено %d", len(ips))
	}
}

func TestParseTargetsFromFile_WithComments(t *testing.T) {
	content := `# This is a comment
192.168.1.1

# Another comment
192.168.1.2
`
	tmpFile := createTempFile(t, content)

	ips, err := ParseTargetsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseTargetsFromFile() error = %v", err)
	}

	if len(ips) != 2 {
		t.Errorf("Ожидалось 2 IP (пропуск комментариев и пустых строк), получено %d", len(ips))
	}
}

func TestParseTargetsFromFile_IPRange(t *testing.T) {
	content := "192.168.1.1-3\n"
	tmpFile := createTempFile(t, content)

	ips, err := ParseTargetsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseTargetsFromFile() error = %v", err)
	}

	expected := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	if len(ips) != len(expected) {
		t.Errorf("Ожидалось %d IP, получено %d", len(expected), len(ips))
	}

	for i, ip := range expected {
		if ips[i] != ip {
			t.Errorf("Ожидался IP %s на позиции %d, получено %s", ip, i, ips[i])
		}
	}
}

func TestParseTargetsFromFile_InvalidIP(t *testing.T) {
	content := "invalid-ip\n"
	tmpFile := createTempFile(t, content)

	_, err := ParseTargetsFromFile(tmpFile.Name())
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного IP")
	}
}

func TestParseTargetsFromFile_InvalidCIDR(t *testing.T) {
	content := "192.168.1.999/24\n"
	tmpFile := createTempFile(t, content)

	_, err := ParseTargetsFromFile(tmpFile.Name())
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного CIDR")
	}
}

func TestParseTargetsFromFile_FileNotFound(t *testing.T) {
	_, err := ParseTargetsFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующего файла")
	}
}

func TestParseTargetsFromFile_EmptyFile(t *testing.T) {
	content := ""
	tmpFile := createTempFile(t, content)

	ips, err := ParseTargetsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseTargetsFromFile() error = %v", err)
	}

	if len(ips) != 0 {
		t.Errorf("Ожидалось 0 IP из пустого файла, получено %d", len(ips))
	}
}

func TestParseIPRange_InvalidFormat(t *testing.T) {
	_, err := parseIPRange("invalid")
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного формата диапазона")
	}
}

func TestParseIPRange_InvalidBaseIP(t *testing.T) {
	_, err := parseIPRange("invalid-10")
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного базового IP")
	}
}

func TestParseIPRange_InvalidEnd(t *testing.T) {
	_, err := parseIPRange("192.168.1.1-abc")
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного конца диапазона")
	}
}

// createTempFile creates a temporary file with the given content for testing.
func createTempFile(t *testing.T, content string) *os.File {
	t.Helper()
	file, err := os.CreateTemp("", "targets-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close and reopen for reading
	file.Close()
	file, err = os.Open(file.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}

	return file
}
