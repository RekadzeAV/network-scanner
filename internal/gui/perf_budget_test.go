package gui

// perfBudgetMultiplier — поправка на замедление при сборке с -race.
//
// Perf-бюджеты измеряются по wall-clock, а детектор гонок замедляет код
// в разы, поэтому базовые пороги под -race заведомо превышаются и тесты
// становятся плавающими. Значение задаётся в файлах с build-тегами
// (perf_budget_race_test.go / perf_budget_norace_test.go).
const perfBudgetMultiplier = perfBudgetRaceMultiplier
