import codecs

filepath = r"e:\GitHub-Ai\network-scanner\docs\IMPLEMENTATION_PLAN.md"

with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

# Обновим статус
content = content.replace(
    "## Статус: В процессе (Этап 3.2 завершён ✅)",
    "## Статус: В процессе (Этап 3.3 завершён ✅)"
)
content = content.replace(
    "**Прогресс:** 6/13 задач завершено (46%)",
    "**Прогресс:** 7/13 задач завершено (54%)"
)

# Обновим задачу 3.3
content = content.replace(
    "### 3.3 GUI интерфейс\n**Приоритет:** 🟢 Low",
    "### 3.3 GUI интерфейс ✅\n**Приоритет:** 🟢 Low"
)

# Добавим задачи для 3.3
content = content.replace(
    "### 3.3 GUI интерфейс ✅\n**Приоритет:** 🟢 Low  \n**Оценка:** 5 часов  \n**Зависимости:** 2.1  \n**Описание:** Графический интерфейс для управления сканированием.",
    "### 3.3 GUI интерфейс ✅\n**Приоритет:** 🟢 Low  \n**Оценка:** 5 часов  \n**Зависимости:** 2.1  \n**Описание:** Графический интерфейс для управления сканированием.\n\n**Задачи:**\n- [x] GUI уже реализован в `internal/gui/` (Fyne GUI framework)\n- [x] Добавить CLI команду `gui`\n- [x] Полная интеграция с CLI\n- [x] Добавить тесты\n\n**Файлы:**\n- `internal/gui/app.go`\n- `internal/gui/*.go` (30+ файлов)\n- `cmd/network-scanner/cmd/scan.go`\n\n**Acceptance Criteria:**\n- GUI приложение запускается через `network-scanner gui`\n- Интерфейс включает: сканирование, топологию, инструменты, инвентаризацию\n- Поддержка SNMP, автопрофилирования, фильтров\n- Сохранение настроек и пресетов"
)

# Обновим итог
content = content.replace(
    "- **Этап 1:** 3/3 задач ✅\n- **Этап 2:** 4/4 задач ✅\n- **Этап 3:** 2/3 задач ✅\n\n**Общий прогресс:** 9/10 задач завершено (90%)",
    "- **Этап 1:** 3/3 задач ✅\n- **Этап 2:** 4/4 задач ✅\n- **Этап 3:** 3/3 задач ✅\n\n**Общий прогресс:** 10/10 задач завершено (100%)"
)

# Write back with UTF-8 BOM
with open(filepath, 'wb') as f:
    f.write(codecs.BOM_UTF8)
    f.write(content.encode('utf-8'))

print("Updated successfully")
