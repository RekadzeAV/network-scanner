# Contributing

## Branching and commits

### Branching модель: Trunk-based with feature branches

- **`main`** — стабильная ветка, deploy-ready. Все feature-ветки ветвятся от `main`.
- **Feature branches** — короткоживущие (`feature/<name>`), максимум 1–3 дня.
- **Release / hotfix** — ветвятся от `main` только при необходимости срочного патча.
- **Squash merge** — все feature-ветки сливаются в `main` через squash (чистая история).
- **No long-lived feature branches** — исключает риск diverging.

### Commit convention (Conventional Commits)

Формат: `<type>(<scope>): <subject>`

```
feat(scanner): add ICMP ping probe
fix(topology): resolve nil-pointer in Export when graph is empty
docs: add CLI_REFERENCE.md
chore: remove internal/legacy/ artifacts
refactor(api): dedupe topology request pipeline
test(inventory): add boundary tests for SaveGraphMLToBytes
security(remoteexec): enforce allowlist by default
```

| Type | Use |
|------|-----|
| `feat` | Новая фича |
| `fix` | Исправление бага |
| `refactor` | Рефакторинг без изменения поведения |
| `test` | Добавление/изменение тестов |
| `docs` | Документация |
| `chore` | Служебные задачи (зависимости, сборка) |
| `perf` | Оптимизация |
| `security` | Изменения в безопасности |

**Subject:** максимум 72 символа, lowercase, без точки в конце.  
**Body (опционально):** wrap at 72 chars, blank line after subject.  
**Footer:** `Fixes #<issue>`, `BREAKING CHANGE: ...` при необходимости.

Every PR must:
- Reference an issue/task in the PR description.
- Include a **test plan**.
- Include **rollback notes** if release behavior changes.

## Response / CODEOWNERS

| Пакет | Owner |
|-------|-------|
| `internal/scanner/` | @RekadzeAV |
| `internal/topology/` | @RekadzeAV |
| `internal/inventory/` | @RekadzeAV |
| `internal/api/` | @RekadzeAV |
| `internal/devicecontrol/` | @RekadzeAV |
| `internal/remoteexec/` | @RekadzeAV |
| `internal/security/` | @RekadzeAV |
| `internal/plugin/`, `internal/eventbus/`, `internal/commands/`, `internal/apperror/`, `internal/configvalidation/` | @RekadzeAV |
| `docs/` | @RekadzeAV |
| `scripts/` | @RekadzeAV |

## Development workflow

1. Run project bootstrap:
   - Linux/macOS: `./scripts/bootstrap.sh`
   - Windows PowerShell: `.\scripts\bootstrap.ps1`
2. Implement changes.
3. Run checks:
   - `make test` (or `go test ./...`)
   - smoke checks from `scripts/`
4. Update documentation when behavior/API changes.

## Naming and style

- Keep package names short and lowercase.
- Use `gofmt` for all Go files.
- Keep new docs in `docs/` using Markdown.

## Pull requests

- Use repository PR template/checklist.
- Include test plan and rollback notes if release behavior changes.
