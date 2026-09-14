# Публикация релиза v2.2.0

## Шаг 1: Получение GITHUB_TOKEN

1. Перейдите на https://github.com/settings/tokens
2. Нажмите **Generate new token (classic)**
3. Выберите scopes: `repo` (полный доступ к репозиторию)
4. Скопируйте токен

## Шаг 2: Публикация через PowerShell

```powershell
# Установить токен
$env:GITHUB_TOKEN = "ваш_токен_здесь"

# Создать Release через API
$repo = "RekadzeAV/network-scanner"
$tag = "v2.2.0"
$title = "Network Scanner v2.2.0"
$body = Get-Content build/release/v2.2/RELEASE_NOTES.md -Raw

$response = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases" `
  -Method POST `
  -Headers @{
    "Authorization" = "token $env:GITHUB_TOKEN"
    "Content-Type" = "application/json"
  } `
  -Body @{
    tag_name = $tag
    name = $title
    body = $body
    draft = $false
    prerelease = $false
  } | ConvertTo-Json -Depth 10

# Загрузить ZIP-архив
$zipPath = "build/release/network-scanner-v2.2.0-windows-amd64.zip"
$releaseUrl = (Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/tags/$tag" `
  -Headers @{ "Authorization" = "token $env:GITHUB_TOKEN" }).upload_url

Invoke-RestMethod -Uri "$releaseUrl?name=network-scanner-v2.2.0-windows-amd64.zip" `
  -Method POST `
  -Headers @{
    "Authorization" = "Bearer $env:GITHUB_TOKEN"
    "Content-Type" = "application/zip"
  } `
  -InFile $zipPath
```

## Шаг 3: Ручная публикация (альтернатива)

1. Перейти на https://github.com/RekadzeAV/network-scanner/releases/new
2. Выбрать тег `v2.2.0`
3. Заголовок: `Network Scanner v2.2.0`
4. Описание: скопировать из `build/release/v2.2/RELEASE_NOTES.md`
5. Прикрепить файл `build/release/network-scanner-v2.2.0-windows-amd64.zip`
6. Нажать **Publish release**

## Проверка

После публикации проверить:
- https://github.com/RekadzeAV/network-scanner/releases/tag/v2.2.0
- ZIP-архив загружен
- Release published
