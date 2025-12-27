# Инструкция по установке и сборке для Windows

## Требования

Для сборки проекта на Windows необходимы:

1. **Go 1.21 или выше** ✅ (установлен: go1.25.5)
2. **C компилятор (GCC)** ✅ (установлен: WinLibs MinGW-w64 GCC 15.2.0)

## Установка C компилятора

Fyne требует CGO для работы на Windows, что в свою очередь требует C компилятор.

### Вариант 1: WinLibs через winget (рекомендуется, автоматическая установка) ✅

Установлено через winget:
```powershell
winget install BrechtSanders.WinLibs.POSIX.UCRT --accept-package-agreements --accept-source-agreements
```

После установки перезапустите терминал или обновите PATH:
```powershell
$env:PATH = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

### Вариант 2: TDM-GCC (простой)

1. Скачайте TDM-GCC с официального сайта: https://jmeubank.github.io/tdm-gcc/
2. Запустите установщик и следуйте инструкциям
3. При установке выберите опцию "Add to PATH"
4. Перезапустите терминал/командную строку

### Вариант 3: MinGW-w64 через MSYS2

1. Скачайте MSYS2: https://www.msys2.org/
2. Установите MSYS2
3. Откройте MSYS2 терминал и выполните:
   ```bash
   pacman -S mingw-w64-x86_64-gcc
   pacman -S mingw-w64-x86_64-toolchain
   ```
4. Добавьте `C:\msys64\mingw64\bin` в PATH

### Вариант 4: Chocolatey (если установлен)

```powershell
choco install mingw -y
```

После установки перезапустите терминал.

## Проверка установки

```powershell
gcc --version
```

Должно вывести информацию о версии GCC.

## Сборка проекта

### GUI версия

```powershell
# Перейдите в директорию проекта
cd "d:\Разработка через ИИ\network-scanner"

# Установите зависимости
go mod download

# Соберите GUI версию
go build -o network-scanner-gui.exe ./cmd/gui
```

### Использование скрипта сборки

```powershell
.\scripts\build.bat
```

## Запуск

После сборки запустите приложение:

```powershell
.\network-scanner-gui.exe
```

## Устранение проблем

### Ошибка: "gcc: executable file not found"

Убедитесь, что GCC установлен и добавлен в PATH:
```powershell
# Проверьте PATH
$env:PATH -split ';' | Select-String -Pattern 'gcc|mingw|tdm'

# Добавьте путь к GCC в PATH (замените путь на ваш)
$env:PATH += ";C:\TDM-GCC-64\bin"
```

### Ошибка: "CGO_ENABLED=0"

CGO должен быть включен для Fyne. Проверьте:
```powershell
go env CGO_ENABLED
```

Должно быть `1`. Если `0`, установите:
```powershell
$env:CGO_ENABLED = "1"
```

### Ошибка с правами доступа

Для работы с сетевыми интерфейсами может потребоваться запуск от имени администратора:
- Кликните правой кнопкой мыши на `network-scanner-gui.exe`
- Выберите "Запуск от имени администратора"

## Альтернативные варианты

Если установка C компилятора проблематична, можно:

1. Собрать проект на Linux/macOS и перенести бинарник
2. Использовать Docker для сборки
3. Использовать GitHub Actions или другой CI/CD для сборки

