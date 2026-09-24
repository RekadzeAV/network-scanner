package scanner

import "testing"

// TestParseLinuxARPOutput_ValidEntry — корректная запись /proc/net/arp.
func TestParseLinuxARPOutput_ValidEntry(t *testing.T) {
	lines := []string{
		"192.168.1.1 0x1 0x2 aa:bb:cc:dd:ee:ff * eth0",
		"192.168.1.50 0x1 0x2 11:22:33:44:55:66 * eth0",
	}
	mac, ok := parseLinuxARPOutput("192.168.1.50", lines)
	if !ok {
		t.Fatal("expected to find entry for 192.168.1.50")
	}
	if mac != "11:22:33:44:55:66" {
		t.Fatalf("got %q, want %q", mac, "11:22:33:44:55:66")
	}
}

// TestParseLinuxARPOutput_NotFound — IP отсутствует в таблице.
func TestParseLinuxARPOutput_NotFound(t *testing.T) {
	lines := []string{
		"192.168.1.1 0x1 0x2 aa:bb:cc:dd:ee:ff * eth0",
	}
	if _, ok := parseLinuxARPOutput("10.0.0.1", lines); ok {
		t.Fatal("expected no match for unknown IP")
	}
}

// TestParseLinuxARPOutput_IncompleteEntry — маркер <incomplete> игнорируется.
func TestParseLinuxARPOutput_IncompleteEntry(t *testing.T) {
	lines := []string{
		"192.168.1.7 0x1 0x0 <incomplete> * eth0",
	}
	if _, ok := parseLinuxARPOutput("192.168.1.7", lines); ok {
		t.Fatal("expected incomplete entry to be ignored")
	}
}

// TestParseLinuxARPOutput_ZeroMAC — нулевой MAC игнорируется.
func TestParseLinuxARPOutput_ZeroMAC(t *testing.T) {
	lines := []string{
		"192.168.1.7 0x1 0x0 00:00:00:00:00:00 * eth0",
	}
	if _, ok := parseLinuxARPOutput("192.168.1.7", lines); ok {
		t.Fatal("expected zero MAC to be ignored")
	}
}

// TestParseLinuxARPOutput_ShortLine — слишком короткая строка пропускается.
func TestParseLinuxARPOutput_ShortLine(t *testing.T) {
	lines := []string{"192.168.1.7 0x1 0x2"}
	if _, ok := parseLinuxARPOutput("192.168.1.7", lines); ok {
		t.Fatal("expected short line to be skipped")
	}
}

// TestParseLinuxARPOutput_Empty — пустой ввод.
func TestParseLinuxARPOutput_Empty(t *testing.T) {
	if _, ok := parseLinuxARPOutput("192.168.1.1", nil); ok {
		t.Fatal("expected no match on empty input")
	}
}

// TestParseDarwinARPLine_StandardFormat — типовой вывод `arp -n` на macOS.
func TestParseDarwinARPLine_StandardFormat(t *testing.T) {
	line := "? (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]"
	mac, ok := parseDarwinARPLine(line)
	if !ok {
		t.Fatal("expected MAC to be parsed")
	}
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("got %q, want %q", mac, "aa:bb:cc:dd:ee:ff")
	}
}

// TestParseDarwinARPLine_NoAt — строка без "at " не содержит MAC.
func TestParseDarwinARPLine_NoAt(t *testing.T) {
	if _, ok := parseDarwinARPLine("192.168.1.1 (192.168.1.1) en0"); ok {
		t.Fatal("expected no MAC when 'at ' is absent")
	}
}

// TestParseDarwinARPLine_Incomplete — маркер (incomplete) не является MAC.
func TestParseDarwinARPLine_Incomplete(t *testing.T) {
	line := "? (192.168.1.1) at (incomplete) on en0 ifscope [ethernet]"
	if _, ok := parseDarwinARPLine(line); ok {
		t.Fatal("expected (incomplete) to be rejected")
	}
}

// TestParseDarwinARPLine_ZeroMAC — нулевой MAC отклоняется.
func TestParseDarwinARPLine_ZeroMAC(t *testing.T) {
	line := "? (192.168.1.1) at 00:00:00:00:00:00 on en0"
	if _, ok := parseDarwinARPLine(line); ok {
		t.Fatal("expected zero MAC to be rejected")
	}
}

// TestParseDarwinARPLine_WrongLength — значение неверной длины не является MAC.
func TestParseDarwinARPLine_WrongLength(t *testing.T) {
	line := "? (192.168.1.1) at aa:bb:cc:dd:ee on en0"
	if _, ok := parseDarwinARPLine(line); ok {
		t.Fatal("expected short value to be rejected")
	}
}

// TestParseDarwinARPLine_EmptyAfterAt — после "at " нет токенов.
func TestParseDarwinARPLine_EmptyAfterAt(t *testing.T) {
	if _, ok := parseDarwinARPLine("? (192.168.1.1) at  "); ok {
		t.Fatal("expected no MAC when nothing follows 'at '")
	}
}

// TestParseDarwinARPLine_AtIsLastToken — "at" без пробела-разделителя в конце.
func TestParseDarwinARPLine_AtAtEnd(t *testing.T) {
	if _, ok := parseDarwinARPLine("? (192.168.1.1) at"); ok {
		t.Fatal("expected no MAC when line ends with 'at'")
	}
}
