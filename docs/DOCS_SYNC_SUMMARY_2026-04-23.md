# Docs Sync Summary (2026-04-23)

## Что синхронизировано

- Примеры CLI и таблицы параметров переведены на актуальный формат флагов: `--network`, `--ports`, `--timeout`, `--threads`, `--show-closed`, `--help`, `--version`.
- Устаревшие примеры с короткими/старыми флагами (`-range`, `-ports`, `-timeout`, `-threads`) заменены в ключевых пользовательских и build-документах.
- Формулировки по версии Go обновлены до текущего baseline (`1.24+`) в установочных и cross-compilation документах.
- Docker-примеры в документации обновлены с `golang:1.21` на `golang:1.24`.
- В историческом блоке `CHANGELOG.md` (релиз `1.0.0`) добавлены пометки, что старые требования/ограничения относятся именно к моменту того релиза.

## Основные затронутые файлы

- `docs/README.md`
- `docs/GUI.md`
- `docs/INSTALL.md`
- `docs/INSTALL_WINDOWS.md`
- `docs/INSTALL_LINUX_CROSS_COMPILER.md`
- `docs/QUICKSTART-macOS.md`
- `docs/BUILD_REQUIREMENTS_WINDOWS.md`
- `docs/CROSS_COMPILATION_QUICKREF.md`
- `docs/CROSS_COMPILATION_WINDOWS.md`
- `docs/SETUP_WINDOWS_CROSS_COMPILE.md`
- `docs/TECHNICAL.md`
- `DEVELOPMENT_MAP.md`
- `Инструкция по эксплуатации.md`
- `scripts/build.bat`
- `scripts/build-os-separate.bat`
- `CHANGELOG.md`

## Что это дает

- Снижение риска путаницы между старым и текущим CLI синтаксисом.
- Согласованность между root-документацией, `docs/*` и шаблонами release README.
- Более предсказуемые шаги для установки/сборки за счет единого baseline по Go.

## Примечание

- В `CHANGELOG.md` исторические пункты старых релизов сохранены, но явно помечены как актуальные на момент соответствующего релиза.
