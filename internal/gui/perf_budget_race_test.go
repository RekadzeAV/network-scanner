//go:build race

package gui

// perfBudgetRaceMultiplier — бюджет под -race увеличивается: детектор гонок
// замедляет исполняемый код, и базовые пороги перестают быть показателями
// производительности.
const perfBudgetRaceMultiplier = 10
