# Developer documentation (`docs/dev/`)

Каталог для сгенерированной и справочной документации для разработчиков.

## Автогенерация godoc

Скрипт `scripts/gen-godoc.ps1` (Windows) / `scripts/gen-godoc.sh` (Linux/macOS)
генерирует API-документацию по внутренним пакетам.

```powershell
# Windows
pwsh -ExecutionPolicy Bypass -File ./scripts/gen-godoc.ps1
```

```bash
# Linux / macOS
./scripts/gen-godoc.sh
```

Поведение:

1. Если установлен `godoc` (`go install golang.org/x/tools/cmd/godoc@latest`),
   генерируется HTML-индекс в `docs/dev/godoc/`.
2. Иначе создаётся текстовый дамп `docs/dev/godoc/GODOC_DUMP.txt` через `go doc -all`
   по всем пакетам `internal/*`.

## Где искать документацию

| Документ | Назначение |
|----------|------------|
| [../ARCHITECTURE.md](../ARCHITECTURE.md) | Архитектура, DI-контейнер, событийная шина (eventbus) |
| [../TECHNICAL.md](../TECHNICAL.md) | Техническая документация для разработчиков |
| [../PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md) | Структура каталогов |
| [../swagger.yaml](../swagger.yaml) | Спецификация REST API (включая Bearer-аутентификацию) |
| [../ROADMAP.md](../ROADMAP.md) | Дорожная карта проекта |
| [CONTRIBUTING.md](../../CONTRIBUTING.md) | Процесс контрибьюции, code review |

> **Примечание:** содержимое `docs/dev/godoc/` — генерируемый артефакт и не
> коммитится в репозиторий (см. `.gitignore`).
