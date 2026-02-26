# Установка кросс-компилятора для Linux на Windows

## ⚠️ Проблема с автоматической установкой

Автоматическая установка кросс-компилятора может столкнуться с проблемой длинных путей в Windows. Ниже приведены несколько альтернативных методов установки.

## Метод 1: Использование скрипта установки (рекомендуется)

Запустите скрипт установки:

```powershell
.\scripts\install-linux-cross-compiler.ps1
```

Если скрипт не сможет распаковать архив из-за длинных путей, используйте один из альтернативных методов ниже.

## Метод 2: Использование WSL (если установлен)

Если у вас установлен WSL (Windows Subsystem for Linux), используйте его для распаковки:

```powershell
# Скачать архив (если еще не скачан)
$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri "https://musl.cc/x86_64-linux-musl-cross.tgz" -OutFile "$env:TEMP\x86_64-linux-musl-cross.tgz"

# Распаковать через WSL
wsl bash -c "cd /mnt/c && mkdir -p cross-compilers && cd cross-compilers && tar -xzf /mnt/c/Users/$env:USERNAME/AppData/Local/Temp/x86_64-linux-musl-cross.tgz"

# Добавить в PATH
$binPath = "C:\cross-compilers\x86_64-linux-musl-cross\bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binPath", "User")
    $env:Path += ";$binPath"
}
```

## Метод 3: Использование 7-Zip

Если у вас установлен 7-Zip:

```powershell
# Скачать архив
$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri "https://musl.cc/x86_64-linux-musl-cross.tgz" -OutFile "$env:TEMP\x86_64-linux-musl-cross.tgz"

# Распаковать через 7-Zip
& "C:\Program Files\7-Zip\7z.exe" x "$env:TEMP\x86_64-linux-musl-cross.tgz" -o"C:\cross-compilers" -y

# Если это tar.gz, сначала распаковать gz, затем tar
& "C:\Program Files\7-Zip\7z.exe" x "$env:TEMP\x86_64-linux-musl-cross.tgz" -o"$env:TEMP" -y
& "C:\Program Files\7-Zip\7z.exe" x "$env:TEMP\x86_64-linux-musl-cross.tar" -o"C:\cross-compilers" -y

# Добавить в PATH
$binPath = "C:\cross-compilers\x86_64-linux-musl-cross\bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binPath", "User")
    $env:Path += ";$binPath"
}
```

## Метод 4: Ручная установка

1. **Скачайте архив вручную:**
   - Откройте браузер
   - Перейдите на https://musl.cc/x86_64-linux-musl-cross.tgz
   - Скачайте файл

2. **Распакуйте архив:**
   - Используйте 7-Zip, WinRAR или другой архиватор
   - Распакуйте в `C:\cross-compilers\`
   - Должна получиться структура: `C:\cross-compilers\x86_64-linux-musl-cross\bin\`

3. **Добавьте в PATH:**
   ```powershell
   $binPath = "C:\cross-compilers\x86_64-linux-musl-cross\bin"
   $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
   [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binPath", "User")
   ```

4. **Перезапустите терминал** для применения изменений PATH

## Проверка установки

После установки проверьте:

```powershell
# Проверить наличие компилятора
x86_64-linux-musl-gcc --version

# Должно вывести информацию о версии GCC
```

## Использование для Go кросскомпиляции

После установки используйте следующие переменные окружения:

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CC = "x86_64-linux-musl-gcc"
$env:CGO_ENABLED = "1"
go build -o network-scanner-gui-linux-amd64 ./cmd/gui
```

## Альтернативные варианты

Если установка кросс-компилятора проблематична, рассмотрите альтернативы:

### Вариант 1: Использование CI/CD
Используйте GitHub Actions или другой CI/CD для автоматической сборки Linux версии:
- GitHub Actions предоставляет Linux runners
- Автоматическая сборка при коммитах

### Вариант 2: Использование Docker
Создайте Docker контейнер для сборки:
```dockerfile
FROM golang:1.21
RUN apt-get update && apt-get install -y gcc
```

### Вариант 3: Сборка на Linux машине
Если у вас есть доступ к Linux машине, соберите там и перенесите бинарник.

## Устранение проблем

### Ошибка: "x86_64-linux-musl-gcc: command not found"
**Причина:** Компилятор не в PATH  
**Решение:** 
1. Проверьте, что путь `C:\cross-compilers\x86_64-linux-musl-cross\bin` добавлен в PATH
2. Перезапустите терминал
3. Проверьте: `$env:Path -split ';' | Select-String -Pattern 'cross-compilers'`

### Ошибка: "Can't create file: Invalid argument"
**Причина:** Проблема с длинными путями в Windows  
**Решение:** Используйте WSL или 7-Zip для распаковки (см. методы выше)

### Ошибка при кросскомпиляции: "CGO_ENABLED=0"
**Причина:** CGO отключен  
**Решение:** 
```powershell
$env:CGO_ENABLED = "1"
```

---

**Последнее обновление:** 2024


