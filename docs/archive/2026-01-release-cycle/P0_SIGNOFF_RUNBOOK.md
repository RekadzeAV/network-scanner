# P0 Sign-off Runbook

Пошаговый runbook для закрытия блокирующего `P0` (Cross-OS closure + CI evidence).

## 1) Prerequisites

- Доступ к GitHub репозиторию `RekadzeAV/network-scanner`.
- `GITHUB_TOKEN` с правами на `repo` и `workflow` (или один раз `gh auth login` после установки GitHub CLI — см. ниже).
- Локально: рабочая копия ветки с актуальными изменениями.

### Windows: установка `make` и GitHub CLI (при необходимости)

Если в PowerShell нет `make` или `gh`, можно поставить через winget (после установки перезапустите терминал, чтобы обновился `PATH`):

```powershell
winget install -e --id ezwinports.make --accept-package-agreements --accept-source-agreements
winget install -e --id GitHub.cli --scope user --accept-package-agreements --accept-source-agreements
gh auth login
```

Скрипты sign-off подхватывают токен в таком порядке: переменная процесса `GITHUB_TOKEN` → пользовательская/системная переменная `GITHUB_TOKEN` → `gh auth token`.

### Быстрый preflight (Windows)

```powershell
.\scripts\p0-signoff-preflight.ps1
# или
make p0-preflight-win
```

Preflight проверяет:
- наличие `GITHUB_TOKEN`;
- доступность `bash/sh` для Unix closure-скриптов;
- наличие успешного recent run для `ci.yml`.

### Агрегированный статус Stage2 sign-off (Windows)

```powershell
.\scripts\stage2-signoff-status.ps1
# или
make stage2-signoff-status-win
```

Агрегатор выполняет одним запуском:
- `stage2-p1/p2/p3` closure;
- `docs-link-check`;
- `p0-signoff-preflight`.

## 2) Cross-OS closure evidence

### Linux/macOS (в целевых средах)

Выполнить:

```bash
go test ./...
./scripts/p1-closure-check.sh
./scripts/p2-closure-check.sh
./scripts/p3-closure-check.sh
./scripts/stage2-p1-closure-check.sh
./scripts/stage2-p2-closure-check.sh
./scripts/stage2-p3-closure-check.sh
```

Критерий успеха:
- Все команды завершаются с `exit code 0`.
- Нет новых ошибок в smoke/closure шагах.

## 3) CI evidence (обязательный green run)

### Windows (PowerShell)

```powershell
$env:GITHUB_TOKEN = "<your_token>"
make p3-close-all-win
```

### Linux/macOS (bash)

```bash
export GITHUB_TOKEN="<your_token>"
make p3-close-all
```

Критерий успеха:
- Скрипт завершен без ошибок.
- В `CI` green required jobs:
  - `Lint`
  - `Test*`
  - `Build and Smoke*`
  - `Stage2 P1 Closure`
  - `Stage2 P3 Closure`

## 4) Обновление чеклистов после green run

Обновить:
- [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md):
  - `Status P3` -> `closed`
  - `CI run URL (green)` -> фактический URL
  - `Дата` и `Owner`
- [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md):
  - закрыть CI evidence пункты
  - закрыть cross-OS closure пункты (по факту прогонов)
- [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md):
  - отметить выполненные `P0` пункты

## 5) Quick verification

Перед финальным sign-off:

```bash
# или powershell-эквиваленты
./scripts/check-ci-status.sh
```

Ожидаемый результат:
- обнаружен успешный recent run;
- required jobs отмечены как `OK`;
- `All required jobs green: YES`.
