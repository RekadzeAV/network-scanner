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
| Require status checks to pass | ✅, указать checks из `.github/workflows/ci.yml`: `Run Tests`, `Run Linter`, `Run govulncheck`, `Race detector (M2 gate)`, `Branch name check` |
| Require conversation resolution | ✅ |
| Require pull request before merging | ✅, approvals: 1 |
| Require signed commits | ⬜ (по желанию) |
| Block force pushes | ✅ |
| Block deletions | ✅ |

## 2. Правило имён веток (все ветки, кроме `main`)

Правило **Create** — паттерн:

```
main|feature/*|fix/*|chore/*|docs/*|security/*|release/*
```

Снимите ограничения `Restrict updates/deletions` для этой правилаетки — они
покрываются правилом выше для `main`. Дополнительно имя ветки проверяется
CI-job'ом `branch-name` (`.github/workflows/branch-name.yml`), который даёт
быстрый фидбек на PR.

## 3. Иммутабельность релизов (теги `v*`)

Отдельный Ruleset для **tags**:

| Параметр | Значение |
|----------|----------|
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
