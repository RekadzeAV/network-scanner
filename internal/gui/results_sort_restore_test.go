package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// sortTestResult — хост с указанным числом открытых/закрытых портов.
func sortTestResult(ip, hostname, deviceType string, open, closed int) scanner.Result {
	res := scanner.Result{IP: ip, Hostname: hostname, DeviceType: deviceType}
	for i := 0; i < open; i++ {
		res.Ports = append(res.Ports, scanner.PortInfo{Port: 1000 + i, State: "open"})
	}
	for i := 0; i < closed; i++ {
		res.Ports = append(res.Ports, scanner.PortInfo{Port: 2000 + i, State: "closed"})
	}
	return res
}

// unsortedResults — набор, в котором порядок по IP отличается от порядка по
// hostname, по типу и по числу открытых портов.
func unsortedResults() []scanner.Result {
	return []scanner.Result{
		sortTestResult("10.0.0.20", "zeta", "router", 1, 2),
		sortTestResult("10.0.0.3", "alpha", "switch", 5, 0),
		sortTestResult("10.0.0.10", "mike", "host", 0, 4),
	}
}

func ipsOf(results []scanner.Result) []string {
	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, r.IP)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSortedResultsForDisplayWithMode проверяет, что каждый режим из селектора
// действительно меняет порядок строк (P0-6: выбор сортировки ни на что не влиял).
func TestSortedResultsForDisplayWithMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want []string
	}{
		// IP: числовой порядок адресов, а не лексикографический.
		{name: "IP", mode: resultsSortByIP, want: []string{"10.0.0.3", "10.0.0.10", "10.0.0.20"}},
		// Hostname: alpha < mike < zeta.
		{name: "Hostname", mode: resultsSortByHostname, want: []string{"10.0.0.3", "10.0.0.10", "10.0.0.20"}},
		// Тип: host < router < switch.
		{name: "Тип", mode: resultsSortByType, want: []string{"10.0.0.10", "10.0.0.20", "10.0.0.3"}},
		// Открытые порты: 5 > 1 > 0.
		{name: "Открытые порты", mode: resultsSortByOpenPorts, want: []string{"10.0.0.3", "10.0.0.20", "10.0.0.10"}},
		{name: "пустой режим = IP", mode: "", want: []string{"10.0.0.3", "10.0.0.10", "10.0.0.20"}},
		{name: "неизвестный режим = IP", mode: "Что угодно", want: []string{"10.0.0.3", "10.0.0.10", "10.0.0.20"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ipsOf(sortedResultsForDisplayWithMode(unsortedResults(), tc.mode))
			if !equalStrings(got, tc.want) {
				t.Fatalf("порядок по %q: получили %v, ожидали %v", tc.mode, got, tc.want)
			}
		})
	}
}

// TestSortedResultsForDisplayWithMode_DoesNotMutateInput — сортировка работает на
// копии, исходный срез результатов не переставляется.
func TestSortedResultsForDisplayWithMode_DoesNotMutateInput(t *testing.T) {
	src := unsortedResults()
	_ = sortedResultsForDisplayWithMode(src, resultsSortByOpenPorts)
	if !equalStrings(ipsOf(src), []string{"10.0.0.20", "10.0.0.3", "10.0.0.10"}) {
		t.Fatalf("исходный срез изменён: %v", ipsOf(src))
	}
}

// TestNormalizeResultsSortMode — значения, которые попадают в настройку сортировки.
func TestNormalizeResultsSortMode(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "IP", value: "IP", want: resultsSortByIP},
		{name: "Hostname", value: "Hostname", want: resultsSortByHostname},
		{name: "Тип", value: "Тип", want: resultsSortByType},
		{name: "Открытые порты", value: "Открытые порты", want: resultsSortByOpenPorts},
		{name: "обрамление пробелами", value: "  Тип\n", want: resultsSortByType},
		{name: "унаследованное HostName", value: "HostName", want: resultsSortByHostname},
		{name: "пустое", value: "", want: ""},
		{name: "неизвестное", value: "Размер окна", want: ""},
		{name: "регистр значим", value: "ip", want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeResultsSortMode(tc.value); got != tc.want {
				t.Fatalf("normalizeResultsSortMode(%q) = %q, ожидали %q", tc.value, got, tc.want)
			}
		})
	}
}

// newAppWithResultsWidgets — App с построенным блоком результатов: селектор
// сортировки создаётся в buildResultsContainer (initResultsFilterWidgets),
// initScanUI сам по себе его не создаёт.
func newAppWithResultsWidgets() *App {
	app := &App{}
	app.initResultsStateAndWidgets()
	app.initResultsModeAndInventoryWidgets()
	app.initResultsFilterWidgets()
	return app
}

// TestResultsSortOptionsMatchSelector гарантирует, что опции селектора, whitelist
// восстановления и константы сортировки не расходятся.
func TestResultsSortOptionsMatchSelector(t *testing.T) {
	app := newAppWithResultsWidgets()
	if app.resultsSortSel == nil {
		t.Fatal("resultsSortSel не создан")
	}
	if !equalStrings(app.resultsSortSel.Options, resultsSortOptions) {
		t.Fatalf("опции селектора %v расходятся с resultsSortOptions %v",
			app.resultsSortSel.Options, resultsSortOptions)
	}
	for _, opt := range app.resultsSortSel.Options {
		if normalizeResultsSortMode(opt) == "" {
			t.Fatalf("опция %q не проходит normalizeResultsSortMode — не восстановится из настроек", opt)
		}
	}
}

// TestResolveResultsSortMode — свёртка «значение из виджета/настроек + текущий
// режим -> что записать в App.resultsSort».
func TestResolveResultsSortMode(t *testing.T) {
	tests := []struct {
		name    string
		current string
		value   string
		want    string
	}{
		{name: "корректное значение", current: resultsSortByIP, value: resultsSortByType, want: resultsSortByType},
		{name: "унаследованное HostName", current: resultsSortByIP, value: "HostName", want: resultsSortByHostname},
		{name: "мусор сохраняет текущий", current: resultsSortByHostname, value: "мусор", want: resultsSortByHostname},
		{name: "пустое сохраняет текущий", current: resultsSortByType, value: "", want: resultsSortByType},
		{name: "мусор + мусор = IP", current: "мусор", value: "мусор", want: resultsSortByIP},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveResultsSortMode(tc.current, tc.value); got != tc.want {
				t.Fatalf("resolveResultsSortMode(%q, %q) = %q, ожидали %q",
					tc.current, tc.value, got, tc.want)
			}
		})
	}
}

// TestResultsSortSelectCallbackUpdatesState — колбэк селектора обязан записывать
// выбранный режим в App.resultsSort (регрессия P0-6: колбэк только сохранял
// настройки, порядок строк не менялся).
func TestResultsSortSelectCallbackUpdatesState(t *testing.T) {
	app := newAppWithResultsWidgets()
	app.resultsSort = resultsSortByIP // начальный дефолт
	if app.resultsSortSel == nil || app.resultsSortSel.OnChanged == nil {
		t.Fatal("селектор сортировки без колбэка")
	}

	// Битое значение не должно менять сортировку.
	app.resultsSortSel.OnChanged("мусор")
	if app.resultsSort != resultsSortByIP {
		t.Fatalf("неизвестное значение изменило сортировку на %q", app.resultsSort)
	}

	for _, opt := range resultsSortOptions {
		app.resultsSortSel.OnChanged(opt)
		if app.resultsSort != opt {
			t.Fatalf("после выбора %q состояние = %q", opt, app.resultsSort)
		}
	}
}
