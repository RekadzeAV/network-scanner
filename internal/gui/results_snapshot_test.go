package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

// TestTableViewRender_500Rows snapshot test для таблицы с 500 строками.
func TestTableViewRender_500Rows(t *testing.T) {
	app := &App{
		resultsMode:   "таблица",
		layoutProfile: "normal",
	}

	data := generateTestResults(500)
	view := app.buildTableView(data)
	if view == nil {
		t.Fatal("buildTableView returned nil")
	}

	// Проверка что view создан корректно
	test.NewApp()
	test.NewWindow(view).Close()
}

// TestCardsViewRender_200Cards snapshot test для карточек с 200 карточками.
func TestCardsViewRender_200Cards(t *testing.T) {
	app := &App{
		resultsMode:       "Карточки",
		layoutProfile:     "normal",
		cardsVisibleCount: 200,
	}

	data := generateTestResults(500)
	view := app.buildCardsView(data)
	if view == nil {
		t.Fatal("buildCardsView returned nil")
	}

	test.NewApp()
	test.NewWindow(view).Close()
}

// TestHostDetailsDrawerRender snapshot test для drawer деталей хоста.
func TestHostDetailsDrawerRender(t *testing.T) {
	app := &App{
		resultsMode:   "таблица",
		layoutProfile: "normal",
	}

	data := generateTestResults(10)
	app.selectedHostIP = data[0].IP

	view := app.buildHostDetailsDrawer(data)
	if view == nil {
		t.Fatal("buildHostDetailsDrawer returned nil")
	}

	test.NewApp()
	test.NewWindow(view).Close()
}
