# Release Operations Cheat-Sheet

Краткий набор команд для дежурного релизного прогона.

## 6-command runbook (copy/paste)

```bash
./scripts/p1-closure-check.sh && ./scripts/p2-closure-check.sh && ./scripts/p3-closure-check.sh && ./scripts/stage2-p1-closure-check.sh && ./scripts/stage2-p2-closure-check.sh && ./scripts/stage2-p3-closure-check.sh
```

```powershell
.\scripts\p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p3-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p3-closure-check.ps1
```

## Unix (Linux/macOS)

```bash
# Stage 1
./scripts/p1-closure-check.sh
./scripts/p2-closure-check.sh
./scripts/p3-closure-check.sh

# Stage 2
./scripts/stage2-p1-closure-check.sh
./scripts/stage2-p2-closure-check.sh
```

## Windows (PowerShell)

```powershell
# Stage 1
.\scripts\p1-closure-check.ps1
.\scripts\p2-closure-check.ps1
.\scripts\p3-closure-check.ps1

# Stage 2
.\scripts\stage2-p1-closure-check.ps1
.\scripts\stage2-p2-closure-check.ps1
.\scripts\stage2-p3-closure-check.ps1
```

## Эквивалент через Makefile

```bash
make p1-check
make p2-check
make p3-check
make stage2-p1-check
make stage2-p2-check
make stage2-p3-check
```

```powershell
make p1-check-win
make p2-check-win
make p3-check-win
make stage2-p1-check-win
make stage2-p2-check-win
make stage2-p3-check-win
```

## Локальные бинарники (скрипты сборки)

Релизные скрипты (`scripts/build.bat`, `scripts/build.sh`, `build-release-windows-only.ps1` и др.) пишут артефакты в **`build/release/<YYYY-MM-DD-N>/`** в корне репозитория (на Windows бинарники — в **`windows\`**). Каталог **`build/`** в `.gitignore`.

## Минимальный релизный sanity

- Проверить, что все closure-команды проходят в целевой среде.
- Проверить [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md) и отметить подтвержденные пункты.
- Для security отчета убедиться, что в HTML есть секции:
  - `CVE Findings`
  - `Risk Signature Findings`

## P0 preflight (рекомендуется перед CI sign-off)

```powershell
.\scripts\p0-signoff-preflight.ps1
# или
make p0-preflight-win
```

Preflight сразу показывает блокеры по:
- `GITHUB_TOKEN`
- доступности `bash/sh` для Unix closure
- наличию успешного recent run `ci.yml`

Если preflight возвращает `BLOCKED`, использовать [P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md).

## Docs sanity (локальные markdown-ссылки)

```powershell
.\scripts\docs-link-check.ps1
# или
make docs-link-check-win
```

## Единый статус Stage2 sign-off (Windows)

```powershell
.\scripts\stage2-signoff-status.ps1
# или
make stage2-signoff-status-win
```

Скрипт агрегирует:
- `stage2-p1/p2/p3` closure
- `docs-link-check`
- `p0-signoff-preflight`

## Stage2/P3 Sign-off Commands

```bash
# 1) Локальный closure Stage2/P3
./scripts/stage2-p3-closure-check.sh

# 2) Статус CI (strict required jobs check)
./scripts/check-ci-status.sh RekadzeAV network-scanner ci.yml 10

# 3) Триггер и ожидание CI (strict required jobs check)
./scripts/trigger-ci-workflow.sh RekadzeAV network-scanner ci.yml main 30 15

# 4) Финализация sign-off чеклиста (после успешного CI)
./scripts/finalize-p3-signoff.sh RekadzeAV network-scanner ci.yml docs/P3_CLOSURE_CHECKLIST.md RekadzeAV
```

```powershell
# 1) Локальный closure Stage2/P3
.\scripts\stage2-p3-closure-check.ps1

# 2) Статус CI (strict required jobs check)
.\scripts\check-ci-status.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -Limit 10

# 3) Триггер и ожидание CI (strict required jobs check)
.\scripts\trigger-ci-workflow.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -Ref main -TimeoutMinutes 30 -PollSeconds 15

# 4) Финализация sign-off чеклиста (после успешного CI)
.\scripts\finalize-p3-signoff.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -ChecklistPath docs/P3_CLOSURE_CHECKLIST.md -SignoffOwner RekadzeAV
```

## Stage2/P3 PR коммуникация

- PR summary/snippets: [RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md](RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md)
- Final PR comment (RU): [FINAL_PR_COMMENT_STAGE2_P3_READY.md](FINAL_PR_COMMENT_STAGE2_P3_READY.md)
- Stage2 commit-ready summary: [STAGE2_100_COMMIT_READY.md](STAGE2_100_COMMIT_READY.md)
- Commit message templates: [COMMIT_READY_STAGE2_SIGNOFF.md](COMMIT_READY_STAGE2_SIGNOFF.md)
