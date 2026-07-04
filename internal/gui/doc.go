// Package gui предоставляет графический интерфейс для Network Scanner.
//
// GUI построено на базе Fyne framework и включает следующие компоненты:
//
//   - Сканирование сети: TCP/UDP порты, ping, ARP, banner grabbing
//   - Результаты: таблица, карточки, фильтры, сортировка
//   - Топология: SNMP сбор данных, построение графа сети
//   - Инструменты: ping, traceroute, DNS, whois, Wake-on-LAN
//   - Инвентаризация: SQLite база данных, сравнение снапшотов
//
// # Структура приложения
//
// App — основной struct приложения, содержит все виджеты и состояние.
//
//	NewApp() *App — создает новый экземпляр GUI приложения.
//	Run() — запускает GUI приложение.
//	Stop() — останавливает сканирование и закрывает приложение.
//
// # Результаты сканирования
//
// Результаты представляются в виде []scanner.Result и обрабатываются через:
//
//	FormatResultsForDisplay() — форматирует результаты в Markdown для отображения.
//	sortedResultsForDisplay() — сортирует результаты по IP или hostname.
//	filterResultsForDisplay() — фильтрует результаты по query.
//	openPortLabels() — генерирует чипы открытых портов.
//
// # Фильтры и пресеты
//
// Поддерживаются следующие типы фильтров:
//
//   - Text filter: поиск по hostname, IP, MAC, device type
//   - CIDR filter: фильтрация по подсети (например 192.168.1.0/24)
//   - Port state filter: "has_open", "has_closed", "has_filtered"
//   - Type filter: "Network Device", "Computer", "Server", "Unknown"
//   - Open ports only: только устройства с открытыми портами
//
// Пресеты фильтров сохраняются в preferences и могут применяться одним кликом.
//
// # Топология сети
//
// Построение топологии включает:
//
//   - SNMP сбор данных (OID queries)
//   - Определение типов устройств (router, switch, host)
//   - Построение графа связей
//   - Экспорт в DOT/GraphML/PNG
//
// # Operations Center
//
// Operations Center управляет фоновыми задачами:
//
//   - Retry: повторный запуск неудачных операций
//   - Cancel: отмена текущих операций
//   - History: история выполненных операций
//
// # Настройки и preferences
//
// Настройки сохраняются в виде key-value пар:
//
//   - scan.* — параметры сканирования (network, ports, timeout, threads)
//   - scan.results_* — параметры отображения результатов
//   - scan.ui.* — параметры UI (split offsets, view mode)
//   - scan.tools.* — параметры инструментов
//
// # DPI Scaling
//
// На Windows приложение автоматически масштабируется под DPI экрана.
// Установите FYNE_SCALE=1 для принудительного масштаба 1:1.
//
// # Пример использования
//
//	import "network-scanner/internal/gui"
//
//	func main() {
//	    app := gui.NewApp()
//	    app.Run()
//	}
package gui
