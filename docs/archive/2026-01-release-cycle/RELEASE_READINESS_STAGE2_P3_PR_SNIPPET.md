# PR Snippet: Stage2/P3

Готовые short/long блоки для PR по направлению `Stage2/P3` (`CVE`, `Security Report`, `Remote Exec`).

## Short (EN)

Stage2/P3 implementation is functionally complete and hardened: CVE matching, security report generation, and Remote Exec policy guardrails are in place.  
Security report now includes redaction controls (`REDACTION: ON|OFF`), explicit unsafe consent for unredacted mode, report metadata (`report-id`, mode, policy, unsafe-consent), and auto naming with report-id.  
CI/ops checks are updated to strict validation of `Stage2 P1 Closure` + `Stage2 P3 Closure`.

## Short (RU)

Реализация Stage2/P3 функционально завершена и усилена по guardrails: CVE matching, генерация security report и policy-ограничения для Remote Exec внедрены.  
Security report поддерживает redaction-контур (`REDACTION: ON|OFF`), explicit consent для unredacted-режима, metadata (`report-id`, mode, policy, unsafe-consent) и auto-именование с report-id.  
CI/операционные проверки обновлены до strict-валидации `Stage2 P1 Closure` + `Stage2 P3 Closure`.

## Long (EN)

### Delivered
- CVE matching module (`internal/cve`) with filtering controls.
- Security report (`internal/report`) with:
  - `CVE Findings`, `Risk Signature Findings`, `Scanned Hosts`;
  - redaction on by default;
  - explicit consent required for unredacted mode;
  - visible redaction status + warning banner for unredacted reports;
  - metadata block (`report-id`, generation mode, policy version, unsafe consent);
  - auto output naming: `security-report-{redacted|unredacted}-<report-id>.html`.
- Remote Exec MVP (`internal/remoteexec`) with:
  - allowlist by host/command;
  - strict policy-file mode;
  - explicit consent;
  - JSONL audit trail;
  - sensitive output masking.
- Closure/CI:
  - `stage2-p3-closure-check.sh/.ps1` added;
  - CI job `Stage2 P3 Closure` added (Ubuntu + Windows);
  - `check-ci-status.*` and `trigger-ci-workflow.*` enforce strict required jobs (`Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`).

### Remaining before formal close
- Run Stage2/P3 closure on target Linux/macOS environments.
- Capture successful CI evidence and update [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) final sign-off fields.

## Long (RU)

### Что реализовано
- Модуль CVE matching (`internal/cve`) с фильтрами.
- Security report (`internal/report`) с:
  - секциями `CVE Findings`, `Risk Signature Findings`, `Scanned Hosts`;
  - redaction по умолчанию;
  - explicit consent для unredacted-режима;
  - видимым статусом redaction + warning-блоком для unredacted;
  - metadata-блоком (`report-id`, mode, policy version, unsafe-consent);
  - auto-именованием файла: `security-report-{redacted|unredacted}-<report-id>.html`.
- Remote Exec MVP (`internal/remoteexec`) с:
  - allowlist по host/command;
  - strict policy-file режимом;
  - обязательным consent;
  - JSONL audit trail;
  - маскированием чувствительных данных в выводе.
- Closure/CI:
  - добавлены `stage2-p3-closure-check.sh/.ps1`;
  - добавлен CI job `Stage2 P3 Closure` (Ubuntu + Windows);
  - скрипты `check-ci-status.*` и `trigger-ci-workflow.*` переведены на strict required jobs (`Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`).

### Что осталось до formal close
- Выполнить Stage2/P3 closure в целевых Linux/macOS средах.
- Зафиксировать успешный CI evidence и обновить финальные поля в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md).
