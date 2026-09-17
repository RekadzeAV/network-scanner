package network

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// HostListFormat определяет формат импортируемого списка хостов
type HostListFormat int

const (
	// FormatAuto автоматически определяет формат
	FormatAuto HostListFormat = iota
	// FormatCSV — CSV файл (IP,Comment)
	FormatCSV
	// FormatTXT — текстовый файл (один IP на строку)
	FormatTXT
	// FormatJSON — JSON массив строк
	FormatJSON
)

// HostEntry представляет одну запись в списке хостов
type HostEntry struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname,omitempty"`
	Comment  string `json:"comment,omitempty"`
	IsCIDR   bool   `json:"is_cidr,omitempty"`
	IsIPv6   bool   `json:"is_ipv6,omitempty"`
}

// HostListImporter импортирует списки хостов из файлов
type HostListImporter struct {
	maxEntries int
}

// NewHostListImporter создаёт новый импортер
func NewHostListImporter(maxEntries int) *HostListImporter {
	if maxEntries <= 0 {
		maxEntries = 65536
	}
	return &HostListImporter{
		maxEntries: maxEntries,
	}
}

// DefaultHostListImporter создаёт импортер с дефолтными параметрами
func DefaultHostListImporter() *HostListImporter {
	return NewHostListImporter(0)
}

// ImportFromFile импортирует список хостов из файла
func (h *HostListImporter) ImportFromFile(path string, format HostListFormat) ([]HostEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("путь к файлу не указан")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл %s: %w", path, err)
	}
	defer file.Close()

	if format == FormatAuto {
		format = detectFormat(path)
	}

	var entries []HostEntry
	switch format {
	case FormatCSV:
		entries, err = h.importCSV(file)
	case FormatTXT:
		entries, err = h.importTXT(file)
	case FormatJSON:
		entries, err = h.importJSON(file)
	default:
		return nil, fmt.Errorf("неподдерживаемый формат: %d", format)
	}

	if err != nil {
		return nil, fmt.Errorf("импорт из %s: %w", path, err)
	}

	// Валидация и расширение CIDR
	validated := make([]HostEntry, 0, len(entries))
	for _, entry := range entries {
		if !h.isValidEntry(entry) {
			continue
		}
		if entry.IsCIDR {
			// Расширяем CIDR в отдельные IP
			cidrEntries, err := expandCIDR(entry)
			if err != nil {
				continue
			}
			validated = append(validated, cidrEntries...)
		} else {
			validated = append(validated, entry)
		}
	}

	if len(validated) > h.maxEntries {
		return nil, fmt.Errorf("слишком много хостов: %d (максимум %d)", len(validated), h.maxEntries)
	}

	return validated, nil
}

// ImportFromString импортирует список хостов из строки
func (h *HostListImporter) ImportFromString(content string, format HostListFormat) ([]HostEntry, error) {
	if content == "" {
		return nil, fmt.Errorf("содержимое пусто")
	}

	var entries []HostEntry
	var err error

	switch format {
	case FormatCSV:
		entries, err = h.parseCSVString(content)
	case FormatTXT:
		entries, err = h.parseTXTString(content)
	case FormatJSON:
		entries, err = h.parseJSONString(content)
	default:
		return nil, fmt.Errorf("неподдерживаемый формат: %d", format)
	}

	if err != nil {
		return nil, fmt.Errorf("парсинг: %w", err)
	}

	// Валидация
	validated := make([]HostEntry, 0, len(entries))
	for _, entry := range entries {
		if h.isValidEntry(entry) {
			validated = append(validated, entry)
		}
	}

	if len(validated) > h.maxEntries {
		return nil, fmt.Errorf("слишком много хостов: %d (максимум %d)", len(validated), h.maxEntries)
	}

	return validated, nil
}

// detectFormat определяет формат файла по расширению
func detectFormat(path string) HostListFormat {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		return FormatCSV
	case strings.HasSuffix(lower, ".json"):
		return FormatJSON
	default:
		return FormatTXT
	}
}

// importCSV импортирует CSV файл
func (h *HostListImporter) importCSV(file *os.File) ([]HostEntry, error) {
	reader := csv.NewReader(bufio.NewReader(file))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // variable number of fields

	var entries []HostEntry
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entries, fmt.Errorf("строка %d: ошибка чтения CSV: %w", lineNum+1, err)
		}
		lineNum++

		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		entry := HostEntry{
			IP: strings.TrimSpace(record[0]),
		}
		if len(record) > 1 {
			entry.Hostname = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			entry.Comment = strings.TrimSpace(record[2])
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// importTXT импортирует текстовый файл
func (h *HostListImporter) importTXT(file *os.File) ([]HostEntry, error) {
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var entries []HostEntry
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Разделяем по пробелу: IP hostname comment
		parts := strings.Fields(line)
		entry := HostEntry{
			IP: parts[0],
		}
		if len(parts) > 1 {
			entry.Hostname = parts[1]
		}
		if len(parts) > 2 {
			comment := strings.Join(parts[2:], " ")
			// Удаляем префикс комментария (# или //)
			comment = strings.TrimPrefix(comment, "#")
			comment = strings.TrimPrefix(comment, "//")
			comment = strings.TrimSpace(comment)
			entry.Comment = comment
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("строка %d: ошибка чтения: %w", lineNum, err)
	}

	return entries, nil
}

// importJSON импортирует JSON файл
func (h *HostListImporter) importJSON(file *os.File) ([]HostEntry, error) {
	var entries []HostEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return nil, fmt.Errorf("парсинг JSON: %w", err)
	}

	return entries, nil
}

// parseCSVString парсит CSV из строки
func (h *HostListImporter) parseCSVString(content string) ([]HostEntry, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var entries []HostEntry

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entries, err
		}

		if len(record) == 0 {
			continue
		}

		entry := HostEntry{
			IP: strings.TrimSpace(record[0]),
		}
		if len(record) > 1 {
			entry.Hostname = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			entry.Comment = strings.TrimSpace(record[2])
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseTXTString парсит TXT из строки
func (h *HostListImporter) parseTXTString(content string) ([]HostEntry, error) {
	var entries []HostEntry
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		entry := HostEntry{
			IP: parts[0],
		}
		if len(parts) > 1 {
			entry.Hostname = parts[1]
		}
		if len(parts) > 2 {
			comment := strings.Join(parts[2:], " ")
			comment = strings.TrimPrefix(comment, "#")
			comment = strings.TrimPrefix(comment, "//")
			comment = strings.TrimSpace(comment)
			entry.Comment = comment
		}

		// Устанавливаем флаги IsCIDR и IsIPv6
		if strings.Contains(entry.IP, "/") {
			_, _, err := net.ParseCIDR(entry.IP)
			if err == nil {
				entry.IsCIDR = true
			}
		} else {
			ip := net.ParseIP(entry.IP)
			if ip != nil {
				entry.IsIPv6 = ip.To4() == nil
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseJSONString парсит JSON из строки
func (h *HostListImporter) parseJSONString(content string) ([]HostEntry, error) {
	var entries []HostEntry
	if err := json.Unmarshal([]byte(content), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// isValidEntry проверяет валидность записи
func (h *HostListImporter) isValidEntry(entry HostEntry) bool {
	entry.IP = strings.TrimSpace(entry.IP)
	if entry.IP == "" {
		return false
	}

	// Проверяем, является ли IP CIDR
	if strings.Contains(entry.IP, "/") {
		_, _, err := net.ParseCIDR(entry.IP)
		if err != nil {
			// Не валидный CIDR, пытаемся как IP
			ip := net.ParseIP(entry.IP)
			if ip == nil {
				return false
			}
			entry.IsCIDR = false
		} else {
			entry.IsCIDR = true
		}
	} else {
		ip := net.ParseIP(entry.IP)
		if ip == nil {
			return false
		}
		entry.IsCIDR = false
		entry.IsIPv6 = ip.To4() == nil
	}

	return true
}

// expandCIDR расширяет CIDR запись в список отдельных IP
func expandCIDR(entry HostEntry) ([]HostEntry, error) {
	_, ipnet, err := net.ParseCIDR(entry.IP)
	if err != nil {
		return nil, err
	}

	var entries []HostEntry
	ip := ipnet.IP
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// Ограничиваем количество хостов
	if hostBits > 16 {
		return nil, fmt.Errorf("слишком большой CIDR: %s", entry.IP)
	}

	count := 1 << hostBits
	if count > 65536 {
		return nil, fmt.Errorf("слишком много адресов в %s", entry.IP)
	}

	for i := 0; i < count; i++ {
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)

		// Инкрементируем IP
		for j := len(ip) - 1; j >= 0; j-- {
			ipCopy[j]++
			if ipCopy[j] > 0 {
				break
			}
		}

		ip = ipCopy
		entries = append(entries, HostEntry{
			IP:       ipCopy.String(),
			Hostname: entry.Hostname,
			Comment:  fmt.Sprintf("%s (из %s)", entry.Comment, entry.IP),
			IsCIDR:   false,
			IsIPv6:   ipCopy.To4() == nil,
		})
	}

	return entries, nil
}

// ExportToFile экспортирует список хостов в файл
func (h *HostListImporter) ExportToFile(entries []HostEntry, path string, format HostListFormat) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("создание файла %s: %w", path, err)
	}
	defer file.Close()

	switch format {
	case FormatCSV:
		return h.exportCSV(file, entries)
	case FormatTXT:
		return h.exportTXT(file, entries)
	case FormatJSON:
		return h.exportJSON(file, entries)
	default:
		return h.exportTXT(file, entries)
	}
}

// exportCSV экспортирует в CSV
func (h *HostListImporter) exportCSV(file *os.File, entries []HostEntry) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, entry := range entries {
		record := []string{entry.IP}
		if entry.Hostname != "" {
			record = append(record, entry.Hostname)
		}
		if entry.Comment != "" {
			record = append(record, entry.Comment)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// exportTXT экспортирует в TXT
func (h *HostListImporter) exportTXT(file *os.File, entries []HostEntry) error {
	for _, entry := range entries {
		line := entry.IP
		if entry.Hostname != "" {
			line += " " + entry.Hostname
		}
		if entry.Comment != "" {
			line += " #" + entry.Comment
		}
		line += "\n"
		if _, err := fmt.Fprint(file, line); err != nil {
			return err
		}
	}

	return nil
}

// exportJSON экспортирует в JSON
func (h *HostListImporter) exportJSON(file *os.File, entries []HostEntry) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// GetHostCount возвращает количество хостов в списке
func (h *HostListImporter) GetHostCount(entries []HostEntry) int {
	count := 0
	for _, entry := range entries {
		if entry.IsCIDR {
			_, ipnet, err := net.ParseCIDR(entry.IP)
			if err == nil {
				ones, bits := ipnet.Mask.Size()
				hostBits := bits - ones
				if hostBits <= 16 {
					count += 1 << hostBits
				}
			}
		} else {
			count++
		}
	}
	return count
}
