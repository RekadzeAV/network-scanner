# P3 Performance Baseline and Budget

Дата замера: `2026-04-23`  
Окружение: Windows 10, `amd64`, CPU `AMD Ryzen 7 7800X3D`

## Цель

Зафиксировать воспроизводимый baseline для P3 и ввести пороги (perf budget), чтобы отслеживать регрессии.

## Методика

### 1) Базовый runtime сканирования (CLI)

Команда:

```powershell
$r1 = Measure-Command { go run ./cmd/network-scanner --network 127.0.0.0/24 --ports 80 --timeout 1 --threads 128 | Out-Null }
$r2 = Measure-Command { go run ./cmd/network-scanner --network 127.0.0.0/23 --ports 80 --timeout 1 --threads 128 | Out-Null }
```

Результат:

- `/24`: `4.44s`
- `/23`: `5.25s`

### 2) Baseline форматирования результатов (proxy для render-path)

Команда:

```powershell
go test ./internal/display -run ^$ -bench BenchmarkFormatResultsAsTextLarge -benchmem
```

Результат:

- `329014 ns/op`
- `925960 B/op`
- `5964 allocs/op`

Примечание: это benchmark форматирования (`FormatResultsAsText`) и он выступает как стабильный proxy для контроля деградации пути подготовки данных к отображению/экспорту.

## Perf budget (P3)

- CLI scan `/24` (профиль выше): `<= 6.0s`
- CLI scan `/23` (профиль выше): `<= 8.0s`
- `BenchmarkFormatResultsAsTextLarge`:  
  - `<= 420000 ns/op`
  - `<= 1.2 MB/op`
  - `<= 7000 allocs/op`

## Правила контроля регрессий

- Изменение считается регрессией, если метрика выходит за budget или деградирует более чем на `20%` от baseline.
- При изменении алгоритма сканирования/рендера разрешено обновить baseline только с отдельным комментарием "почему это ожидаемо".
- Обновление baseline выполняется осознанно, а не автоматически.
