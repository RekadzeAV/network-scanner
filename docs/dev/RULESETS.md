# GitHub Rulesets — настройка защиты веток (Этап 0)

Rulesets настраиваются **в веб-интерфейсе GitHub** (или через `gh api`) — сами
по себе из репозитория они не создаются. Этот документ — источник истины для
их параметров.

Откройте: **Settings → Rules → Rulesets → New branch ruleset**.

## 1. Правила для `main`

| Параметр | Значение |
|----------|----------|
| Bypass list | — (не добавлять) |
| Restrict creations | ✅ |
| Restrict updates | ✅ |
| Restrict deletions | ✅ |
| Require linear history | ✅ (squash/merge commit без merge-коммитов) |
| Require status checks to pass | ✅, указать checks из `.github/workflows/go.yml`: `test`, `build`, `lint`, `Security scan (govulncheck)`, `D-Track Smoke Tests` |
| Require conversation resolution | ✅ |
| Require pull request before merging | ✅, approvals: 1 |
| Require signed commits | ⬜ (по желанию) |
| Block force pushes | ✅ |
| Block deletions | ✅ |

> **Примечание:** `build` запускается в матрице (linux/amd64, linux/arm64,
> windows/amd64, windows/arm64, darwin/arm64). На main требуется прохождение
> всех активных job'ов CI.

## 2. Правило имён веток (все ветки, кроме `main`)

Правило **Create** — паттерн:

```
main|feature/*|fix/*|chore/*|docs/*|security/*|release/*
```

Снимите ограничения `Restrict updates/deletions` для этого ruleset'а — они
покрываются правилом выше для `main`.

## 3. Иммутабельность релизов (теги `v*`)

Отдельный Ruleset для **tags**:

| Параметр | Значение |
|----------|-----------|
| Target | Tags matching `v*` |
| Restrict creations | ✅ (только release-workflow может создать тег) |
| Restrict updates | ✅ |
| Restrict deletions | ✅ |
| Bypass list | GitHub Actions (`GITHUB_TOKEN` release workflow) |

## 4. Проверка через gh CLI

```bash
# Посмотреть существующие rulesets
gh api repos/RekadzeAV/network-scanner/rulesets

# Создать ruleset из JSON (пример для main)
gh api --method POST repos/RekadzeAV/network-scanner/rulesets \
  -f name=protect-main -f target=branch \
  -f enforcement=active \
  -F conditions[ref_name][include][0]=~default_branch \
  -F rules[0][type]=deletion \
  -F rules[1][type]=non_fast_forward \
  -F rules[2][type]=pull_request \
  -F rules[2][parameters][required_approving_review_count]=1
```

## Связанные документы

- Модель веток: [BRANCH_PURPOSE.md](BRANCH_PURPOSE.md)
- Процесс вклада: [../../CONTRIBUTING.md](../../CONTRIBUTING.md)
