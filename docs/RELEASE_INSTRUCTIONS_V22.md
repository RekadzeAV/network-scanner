# Инструкция по релизу v2.2.0

## Сборка артефактов

### Windows (локально)
```powershell
# CLI
go build -o build/release/v2.2/network-scanner-windows-amd64.exe -ldflags="-X main.Version=2.2.0" ./cmd/network-scanner

# GUI
go build -o build/release/v2.2/network-scanner-gui-windows-amd64.exe ./cmd/gui
```

### Linux/macOS (в CI)
GitHub Actions workflow `.github/workflows/ci.yml` автоматически собирает для всех платформ.

## Артефакты

Все бинарники находятся в `build/release/v2.2/`:
- `network-scanner-windows-amd64.exe` (~58 MB)
- `network-scanner-gui-windows-amd64.exe` (~61 MB)

## Документация релиза

- `build/release/v2.2/README.md` — быстрый старт
- `build/release/v2.2/RELEASE_NOTES.md` — что нового
- `build/release/v2.2/CHECKSUMS.md` — проверки целостности

## Архив

ZIP-архив доступен:
- `build/release/network-scanner-v2.2.0-windows-amd64.zip` (~48 MB)

## Публикация

1. Создать GitHub Release
2. Загрузить ZIP-архив
3. Прикрепить CHECKSUMS.md
4. Указать тег `v2.2.0`
