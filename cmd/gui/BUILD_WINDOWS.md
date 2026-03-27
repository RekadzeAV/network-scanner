# Сборка GUI приложения для Windows

## Скрытие консольного окна

Для скрытия консольного окна при запуске GUI приложения на Windows используется флаг линкера `-H windowsgui`.

## Сборка

### На Windows

```powershell
go build -ldflags="-s -w -H windowsgui" -o network-scanner-gui.exe ./cmd/gui
```

### Использование скрипта

```powershell
.\scripts\build.bat
```

Скрипт автоматически добавит флаг `-H windowsgui` для GUI версии.

### Кросскомпиляция из Linux/macOS

```bash
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 go build -ldflags="-s -w -H windowsgui" -o network-scanner-gui-windows-amd64.exe ./cmd/gui
```

Или используйте скрипт:

```bash
./scripts/build-windows.sh
```

## Примечания

- Флаг `-H windowsgui` указывает линкеру создать Windows GUI приложение вместо консольного приложения
- Это предотвращает появление консольного окна при запуске GUI приложения
- Флаг работает только для Windows (`GOOS=windows`)
