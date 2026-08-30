# Release Summary - Stage2/P3

Краткий итог по направлению `Stage2/P3` (`CVE`, `Security Report`, `Remote Exec`) для PR/release-коммуникации.

## Что уже закрыто

- Добавлен базовый CVE-анализ (`internal/cve`) с фильтрами по `CVSS` и возрасту.
- Реализован HTML security report (`internal/report`) с секциями:
  - `CVE Findings`
  - `Risk Signature Findings`
  - `Scanned Hosts`
- Добавлены guardrails для security report:
  - redaction по умолчанию (`--security-report-redact=true`);
  - explicit consent для unredacted (`--security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT`);
  - индикатор `REDACTION: ON|OFF` + warning в unredacted-режиме;
  - metadata в отчете: `report-id`, `mode`, `policy`, `unsafe-consent`;
  - auto-именование файла с `report-id` (`--security-report-file auto`).
- Реализован `Remote Exec` MVP (`internal/remoteexec`) с ограничениями:
  - `ssh|wmi|winrm`;
  - allowlist по host/command;
  - strict policy file режим (`--remote-exec-policy-strict`);
  - обязательный consent (`--remote-exec-consent I_UNDERSTAND`);
  - audit trail в JSONL (`--remote-exec-audit-log`);
  - masking чувствительных данных в CLI/audit.
- Введен единый redaction-модуль (`internal/redact`) и покрыт тестами.
- Добавлены Stage2/P3 closure-скрипты:
  - `scripts/stage2-p3-closure-check.sh`
  - `scripts/stage2-p3-closure-check.ps1`
- CI расширен job `Stage2 P3 Closure` (Ubuntu + Windows).
- CI helper-скрипты `check-ci-status.*` и `trigger-ci-workflow.*` переведены в strict/fail-fast режим с проверкой `Stage2 P1` + `Stage2 P3`.

## Что осталось до formal close

- Подтвердить `Stage2 P3 Closure` в целевых средах Linux/macOS (ручной/операционный прогон).
- Получить и зафиксировать evidence успешного CI run с зелеными jobs:
  - `Lint`
  - `Test*`
  - `Build and Smoke*`
  - `Stage2 P1 Closure`
  - `Stage2 P3 Closure`
- Обновить sign-off поля в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (run URL, owner, дата).

## Минимальный operational runbook

```bash
./scripts/stage2-p3-closure-check.sh
./scripts/check-ci-status.sh RekadzeAV network-scanner ci.yml 10
./scripts/trigger-ci-workflow.sh RekadzeAV network-scanner ci.yml main 30 15
./scripts/finalize-p3-signoff.sh RekadzeAV network-scanner ci.yml docs/P3_CLOSURE_CHECKLIST.md RekadzeAV
```

```powershell
.\scripts\stage2-p3-closure-check.ps1
.\scripts\check-ci-status.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -Limit 10
.\scripts\trigger-ci-workflow.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -Ref main -TimeoutMinutes 30 -PollSeconds 15
.\scripts\finalize-p3-signoff.ps1 -Owner RekadzeAV -Repo network-scanner -WorkflowFile ci.yml -ChecklistPath docs/P3_CLOSURE_CHECKLIST.md -SignoffOwner RekadzeAV
```

## Связанные документы для PR/sign-off

- [RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md](RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md) - short/long PR description blocks (EN/RU).
- [FINAL_PR_COMMENT_STAGE2_P3_READY.md](FINAL_PR_COMMENT_STAGE2_P3_READY.md) - готовый финальный PR-комментарий (RU).
- [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) - официальный чеклист закрытия и sign-off поля.
