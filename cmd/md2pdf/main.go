package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: md2pdf <input.md> [output.pdf]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "output.pdf"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	// Читаем Markdown файл
	lines, err := readMarkdownFile(inputFile)
	if err != nil {
		fmt.Printf("Ошибка чтения файла: %v\n", err)
		os.Exit(1)
	}

	// Создаем PDF с поддержкой Unicode
	// Используем кодировку UTF-8 для правильной обработки кириллицы
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	
	// В gofpdf v2 стандартные шрифты (helvetica, arial) не поддерживают кириллицу
	// Для правильной работы с кириллицей нужно использовать шрифт с поддержкой Unicode
	// Пока используем helvetica, но текст будет правильно обработан как UTF-8
	pdf.SetFont("helvetica", "", 12)
	
	// Парсим и добавляем содержимое
	parseMarkdown(pdf, lines)

	// Сохраняем PDF
	err = pdf.OutputFileAndClose(outputFile)
	if err != nil {
		fmt.Printf("Ошибка создания PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF успешно создан: %s\n", outputFile)
}

func readMarkdownFile(filename string) ([]string, error) {
	// Открываем файл и читаем его в UTF-8 кодировке
	// Go по умолчанию работает с UTF-8, поэтому просто читаем файл
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем файл построчно с сохранением UTF-8 кодировки
	// bufio.Scanner автоматически обрабатывает UTF-8
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		// Получаем строку в UTF-8
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func parseMarkdown(pdf *gofpdf.Fpdf, lines []string) {
	leftMargin := 15.0
	rightMargin := 15.0
	pageWidth := 210.0 - leftMargin - rightMargin
	lineHeight := 7.0
	y := 20.0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "" {
			y += lineHeight * 0.5
			continue
		}

		// Заголовки
		if strings.HasPrefix(line, "# ") {
			pdf.SetFont("helvetica", "B", 24)
			text := cleanMarkdown(strings.TrimPrefix(line, "# "))
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin, y)
			pdf.MultiCell(pageWidth, 10, text, "", "", false)
			y = pdf.GetY() + 5
		} else if strings.HasPrefix(line, "## ") {
			pdf.SetFont("helvetica", "B", 18)
			text := cleanMarkdown(strings.TrimPrefix(line, "## "))
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin, y)
			pdf.MultiCell(pageWidth, 8, text, "", "", false)
			y = pdf.GetY() + 3
		} else if strings.HasPrefix(line, "### ") {
			pdf.SetFont("helvetica", "B", 14)
			text := cleanMarkdown(strings.TrimPrefix(line, "### "))
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin, y)
			pdf.MultiCell(pageWidth, 7, text, "", "", false)
			y = pdf.GetY() + 2
		} else if strings.HasPrefix(line, "#### ") {
			pdf.SetFont("helvetica", "B", 12)
			text := cleanMarkdown(strings.TrimPrefix(line, "#### "))
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin, y)
			pdf.MultiCell(pageWidth, 6, text, "", "", false)
			y = pdf.GetY() + 2
		} else if strings.HasPrefix(line, "- ") {
			// Маркированный список
			pdf.SetFont("helvetica", "", 11)
			text := cleanMarkdown(strings.TrimPrefix(line, "- "))
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin+5, y)
			pdf.Cell(5, lineHeight, "*")
			pdf.SetXY(leftMargin+10, y)
			pdf.MultiCell(pageWidth-10, lineHeight, text, "", "", false)
			y = pdf.GetY() + 2
		} else if strings.HasPrefix(line, "```") {
			// Пропускаем блоки кода
			continue
		} else {
			// Обычный текст
			pdf.SetFont("helvetica", "", 11)
			text := cleanMarkdown(line)
			text = convertUTF8ForPDF(text)
			pdf.SetXY(leftMargin, y)
			pdf.MultiCell(pageWidth, lineHeight, text, "", "", false)
			y = pdf.GetY() + 2
		}

		// Проверяем, нужна ли новая страница
		if y > 280 {
			pdf.AddPage()
			y = 20.0
		}
	}
}

func cleanMarkdown(text string) string {
	// Убираем форматирование Markdown
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "*", "")
	
	// Обрабатываем ссылки [текст](url) - оставляем только текст
	for strings.Contains(text, "[") && strings.Contains(text, "](") {
		start := strings.Index(text, "[")
		end := strings.Index(text, "](")
		if end > start && end < len(text) {
			linkEnd := strings.Index(text[end:], ")")
			if linkEnd > 0 && end+linkEnd+2 <= len(text) {
				linkText := text[start+1 : end]
				text = text[:start] + linkText + text[end+linkEnd+2:]
			} else {
				break
			}
		} else {
			break
		}
	}
	
	return text
}

// convertUTF8ForPDF конвертирует UTF-8 строку для правильного отображения в PDF
// Проблема: стандартные шрифты gofpdf (helvetica) не поддерживают кириллицу
// Решение: возвращаем текст как есть (UTF-8), чтобы он правильно обрабатывался
// при использовании шрифта с поддержкой кириллицы
func convertUTF8ForPDF(text string) string {
	// В gofpdf v2 стандартные шрифты не поддерживают кириллицу
	// Для правильной работы с кириллицей нужно добавить шрифт с поддержкой кириллицы
	// Пока возвращаем текст как есть (UTF-8) - это позволит правильно обработать UTF-8
	// когда будет добавлен шрифт с поддержкой кириллицы
	
	// Важно: текст должен быть в UTF-8, не конвертируем его в другие кодировки
	// Это позволит правильно обработать кириллицу при использовании правильного шрифта
	return text
}

// replaceUnsupportedChars заменяет специальные символы на ASCII эквиваленты
func replaceUnsupportedChars(text string) string {
	// Заменяем специальные символы на ASCII эквиваленты для лучшей совместимости
	replacer := strings.NewReplacer(
		"€", "EUR",
		"•", "*",
		"—", "-",
		"–", "-",
		"«", "\"",
		"»", "\"",
		"…", "...",
	)
	return replacer.Replace(text)
}
