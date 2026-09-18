package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// currentThemeIsDark определяет активный вариант темы (для адаптивных цветов
// рендера, задаваемых в коде, а не через fyne.Theme). Порядок определения:
// 1) ModernTheme, установленный через applyTheme (прямая проверка isDark);
// 2) вариант темы приложения (режим «системная»);
// 3) светлый вариант по умолчанию.
func (a *App) currentThemeIsDark() bool {
	if a == nil {
		return false
	}
	if a.myApp != nil {
		if mt, ok := a.myApp.Settings().Theme().(*ModernTheme); ok {
			return mt.isDark
		}
		if a.myApp.Settings().ThemeVariant() == theme.VariantDark {
			return true
		}
	}
	return false
}

// chipBackground возвращает фон чипов портов, адаптивный к варианту темы
// (светлый: светло-серый; тёмный: тёмно-серый — вместо белёсого по умолчанию).
func (a *App) chipBackground() color.Color {
	if a.currentThemeIsDark() {
		return color.RGBA{R: 0x2D, G: 0x2D, B: 0x30, A: 255}
	}
	return color.RGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 255}
}

// rowBackground возвращает фон строк/карточек таблицы результатов,
// адаптивный к варианту темы.
func (a *App) rowBackground() color.Color {
	if a.currentThemeIsDark() {
		return color.RGBA{R: 0x25, G: 0x25, B: 0x28, A: 255}
	}
	return color.RGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 255}
}

// themeColorSuccess — акцент успеха (зелёный бейдж автопрофиля);
// цвета сохранены из прежнего UI (читаемы на светлой и тёмной теме).
func themeColorSuccess() color.Color {
	return color.RGBA{R: 60, G: 170, B: 80, A: 255}
}

// themeColorDisabled — цвет отключенного состояния.
func themeColorDisabled() color.Color {
	return color.RGBA{R: 140, G: 140, B: 140, A: 255}
}

func (a *App) buildPieChart(title string, data map[string]int) fyne.CanvasObject {
	cacheKey := buildPieChartCacheKey(title, data)
	if a.pieChartCache != nil {
		if res, ok := a.pieChartCache[cacheKey]; ok {
			pie := canvas.NewImageFromResource(res)
			pie.FillMode = canvas.ImageFillContain
			pie.SetMinSize(fyne.NewSize(200, 200))
			return widget.NewCard(title, "", container.NewVBox(pie, a.buildPieLegend(data)))
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, 260, 260))
	// Фон диаграммы адаптивен к теме: тёмный вариант вместо белёсого,
	// иначе pie-чарт выглядит как белый квадрат на тёмном фоне.
	pieBg := color.RGBA{R: 250, G: 250, B: 250, A: 255}
	if a.currentThemeIsDark() {
		pieBg = color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 255}
	}
	for y := 0; y < 260; y++ {
		for x := 0; x < 260; x++ {
			img.Set(x, y, pieBg)
		}
	}

	total := 0
	type statItem struct {
		key   string
		value int
	}
	items := make([]statItem, 0, len(data))
	for k, v := range data {
		if v <= 0 {
			continue
		}
		total += v
		items = append(items, statItem{key: k, value: v})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].value != items[j].value {
			return items[i].value > items[j].value
		}
		return items[i].key < items[j].key
	})
	if total > 0 {
		start := -math.Pi / 2
		cx, cy, radius := 130.0, 130.0, 96.0
		for i, item := range items {
			portion := float64(item.value) / float64(total)
			end := start + portion*2*math.Pi
			for y := 0; y < 260; y++ {
				for x := 0; x < 260; x++ {
					dx := float64(x) - cx
					dy := float64(y) - cy
					distance := math.Sqrt(dx*dx + dy*dy)
					if distance > radius {
						continue
					}
					angle := math.Atan2(dy, dx)
					if angleInSector(angle, start, end) {
						img.Set(x, y, piePalette[i%len(piePalette)])
					}
				}
			}
			start = end
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	resource := fyne.NewStaticResource("pie-"+title+".png", buf.Bytes())
	if a.pieChartCache != nil {
		a.pieChartCache[cacheKey] = resource
	}
	pie := canvas.NewImageFromResource(resource)
	pie.FillMode = canvas.ImageFillContain
	pie.SetMinSize(fyne.NewSize(200, 200))

	return widget.NewCard(title, "", container.NewVBox(pie, a.buildPieLegend(data)))
}

func (a *App) buildPieLegend(data map[string]int) fyne.CanvasObject {
	type statItem struct {
		key   string
		value int
	}
	total := 0
	items := make([]statItem, 0, len(data))
	for k, v := range data {
		if v <= 0 {
			continue
		}
		total += v
		items = append(items, statItem{key: k, value: v})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].value != items[j].value {
			return items[i].value > items[j].value
		}
		return items[i].key < items[j].key
	})
	legend := make([]fyne.CanvasObject, 0, len(items))
	for i, item := range items {
		colorSwatch := canvas.NewRectangle(piePalette[i%len(piePalette)])
		swatch := container.NewGridWrap(fyne.NewSize(12, 12), colorSwatch)
		percent := 0.0
		if total > 0 {
			percent = float64(item.value) * 100 / float64(total)
		}
		legend = append(legend, container.NewHBox(
			swatch,
			widget.NewLabel(fmt.Sprintf("%s: %d (%.1f%%)", item.key, item.value, percent)),
		))
	}
	if len(legend) == 0 {
		legend = append(legend, widget.NewLabel("Нет данных"))
	}
	return container.NewVBox(legend...)
}

func buildPieChartCacheKey(title string, data map[string]int) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("|")
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(strconv.Itoa(data[k]))
		sb.WriteString(";")
	}
	return sb.String()
}

func angleInSector(angle, start, end float64) bool {
	twoPi := 2 * math.Pi
	for angle < 0 {
		angle += twoPi
	}
	for start < 0 {
		start += twoPi
		end += twoPi
	}
	angle = math.Mod(angle, twoPi)
	start = math.Mod(start, twoPi)
	end = math.Mod(end, twoPi)
	if end >= start {
		return angle >= start && angle <= end
	}
	// Сектор пересекает 0 радиан.
	return angle >= start || angle <= end
}
