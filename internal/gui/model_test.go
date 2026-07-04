package gui

import (
	"testing"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/data/binding"
)

func TestNewAppModel(t *testing.T) {
	model := NewAppModel()
	if model == nil {
		t.Fatal("NewAppModel() вернул nil")
	}
	if model.GetResults() == nil {
		t.Error("GetResults() вернул nil вместо пустого слайса")
	}
	if len(model.GetResults()) != 0 {
		t.Errorf("Ожидался пустой слайс, получено %d элементов", len(model.GetResults()))
	}
}

func TestAddHostResult(t *testing.T) {
	model := NewAppModel()

	result := scanner.Result{
		IP:      "192.168.1.1",
		IsAlive: true,
	}

	model.AddHostResult(result)

	results := model.GetResults()
	if len(results) != 1 {
		t.Errorf("Ожидался 1 результат, получено %d", len(results))
	}
	if results[0].IP != "192.168.1.1" {
		t.Errorf("Ожидался IP 192.168.1.1, получено %s", results[0].IP)
	}
}

// TestUpdateProgressDirect проверяет binding напрямую, без fyne.Do
func TestUpdateProgressDirect(t *testing.T) {
	model := NewAppModel()

	// Проверяем, что binding инициализирован
	progressVal := model.GetProgressValue()
	if progressVal == nil {
		t.Fatal("GetProgressValue() вернул nil")
	}

	// Устанавливаем значение напрямую через binding
	err := progressVal.Set(0.5)
	if err != nil {
		t.Fatalf("Ошибка установки прогресса: %v", err)
	}

	// Получаем значение
	val, err := progressVal.Get()
	if err != nil {
		t.Fatalf("Ошибка получения значения прогресса: %v", err)
	}
	if val != 0.5 {
		t.Errorf("Ожидался прогресс 0.5, получено %f", val)
	}
}

func TestSetStatusDirect(t *testing.T) {
	model := NewAppModel()

	statusVal := model.GetStatusText()
	if statusVal == nil {
		t.Fatal("GetStatusText() вернул nil")
	}

	// Устанавливаем значение напрямую
	err := statusVal.Set("Сканирование...")
	if err != nil {
		t.Fatalf("Ошибка установки статуса: %v", err)
	}

	// Получаем значение
	status, err := statusVal.Get()
	if err != nil {
		t.Fatalf("Ошибка получения статуса: %v", err)
	}
	if status != "Сканирование..." {
		t.Errorf("Ожидался статус 'Сканирование...', получено '%s'", status)
	}
}

func TestSetScanningDirect(t *testing.T) {
	model := NewAppModel()

	disabledVal := model.GetScanButtonDisabled()
	if disabledVal == nil {
		t.Fatal("GetScanButtonDisabled() вернул nil")
	}

	// Устанавливаем true (сканирование активно)
	err := disabledVal.Set(true)
	if err != nil {
		t.Fatalf("Ошибка установки состояния: %v", err)
	}

	disabled, err := disabledVal.Get()
	if err != nil {
		t.Fatalf("Ошибка получения состояния кнопки: %v", err)
	}
	if !disabled {
		t.Error("Ожидалось, что кнопка отключена при сканировании")
	}

	// Устанавливаем false (сканирование остановлено)
	err = disabledVal.Set(false)
	if err != nil {
		t.Fatalf("Ошибка сброса состояния: %v", err)
	}

	disabled, err = disabledVal.Get()
	if err != nil {
		t.Fatalf("Ошибка получения состояния кнопки: %v", err)
	}
	if disabled {
		t.Error("Ожидалось, что кнопка включена после остановки сканирования")
	}
}

func TestGetStatusDirect(t *testing.T) {
	model := NewAppModel()

	// Устанавливаем статус через binding
	statusVal := model.GetStatusText()
	err := statusVal.Set("Тестовый статус")
	if err != nil {
		t.Fatalf("Ошибка установки статуса: %v", err)
	}

	// Получаем статус через метод модели
	status := model.GetStatus()
	if status != "Тестовый статус" {
		t.Errorf("Ожидался статус 'Тестовый статус', получено '%s'", status)
	}
}

func TestIsScanningDirect(t *testing.T) {
	model := NewAppModel()

	// Сначала не сканируем (кнопка включена)
	disabledVal := model.GetScanButtonDisabled()
	disabledVal.Set(false)

	// Проверяем через binding
	val, _ := disabledVal.Get()
	if val {
		t.Error("Ожидалось, что сканирование не активно")
	}

	// Устанавливаем сканирование
	disabledVal.Set(true)

	// Проверяем
	val, _ = disabledVal.Get()
	if !val {
		t.Error("Ожидалось, что сканирование активно")
	}
}

func TestMultipleResults(t *testing.T) {
	model := NewAppModel()

	results := []scanner.Result{
		{IP: "192.168.1.1", IsAlive: true},
		{IP: "192.168.1.2", IsAlive: true},
		{IP: "192.168.1.3", IsAlive: false},
	}

	for _, r := range results {
		model.AddHostResult(r)
	}

	got := model.GetResults()
	if len(got) != 3 {
		t.Fatalf("Ожидалось 3 результата, получено %d", len(got))
	}

	// Проверяем порядок
	expectedIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for i, ip := range expectedIPs {
		if got[i].IP != ip {
			t.Errorf("Ожидался IP %s на позиции %d, получено %s", ip, i, got[i].IP)
		}
	}
}

func TestProgressRangeDirect(t *testing.T) {
	model := NewAppModel()
	progressVal := model.GetProgressValue()

	// Тестируем границы диапазона
	testCases := []struct {
		percent float64
		want    float64
	}{
		{0.0, 0.0},
		{0.25, 0.25},
		{0.5, 0.5},
		{0.75, 0.75},
		{1.0, 1.0},
	}

	for _, tc := range testCases {
		err := progressVal.Set(tc.percent)
		if err != nil {
			t.Fatalf("Ошибка установки прогресса %f: %v", tc.percent, err)
		}
		got, err := progressVal.Get()
		if err != nil {
			t.Fatalf("Ошибка получения прогресса %f: %v", tc.percent, err)
		}
		if got != tc.want {
			t.Errorf("Ожидался прогресс %f, получено %f", tc.want, got)
		}
	}
}

func TestBindingTypes(t *testing.T) {
	model := NewAppModel()

	// Проверяем, что binding инициализированы правильно
	if model.GetProgressValue() == nil {
		t.Error("progressValue не инициализирован")
	}
	if model.GetStatusText() == nil {
		t.Error("statusText не инициализирован")
	}
	if model.GetScanButtonDisabled() == nil {
		t.Error("scanButtonDisabled не инициализирован")
	}
	if model.GetResultsList() == nil {
		t.Error("resultsList не инициализирован")
	}

	// Проверяем типы
	var _ binding.Float = model.GetProgressValue()
	var _ binding.String = model.GetStatusText()
	var _ binding.Bool = model.GetScanButtonDisabled()
	var _ binding.UntypedList = model.GetResultsList()
}
