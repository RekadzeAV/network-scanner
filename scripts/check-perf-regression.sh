#!/usr/bin/env bash
# =============================================================================
# Performance Regression Check — сравнение бенчмарков с baseline
# =============================================================================
# Использование:
#   ./scripts/check-perf-regression.sh [путь_к_baseline]
#
# Правила:
#   1. allocs/op не должен расти более чем на +1 alloc (шум планировщика)
#   2. bytes/op не должен расти более чем на 30% — только для CPU-bound
#   3. ns/op не должен расти более чем на 40% — только для CPU-bound
#      бенчмарков (список ниже). Сетевые и async бенчмарки не сравниваются.
# =============================================================================
set -euo pipefail

BASELINE="${1:-internal/benchmark/benchmarks_baseline.txt}"
CURRENT_RAW="$(mktemp)"
CURRENT="$(mktemp)"
trap 'rm -f "$CURRENT_RAW" "$CURRENT"' EXIT

# CPU-bound бенчмарки (ns/op стабилен между запусками)
CPU_ONLY_BENCHMARKS="
BenchmarkPluginRegistry_Register
BenchmarkPluginRegistry_GetAll
BenchmarkEventBus_Subscribe
BenchmarkEventBus_SubscribeAny
BenchmarkConfigValidation_Schema
BenchmarkConfigValidation_Specific
BenchmarkConfigValidation_InvalidConfig
BenchmarkConfigValidation_EmptySchema
BenchmarkPluginRegistry_MemoryAlloc
BenchmarkConfigValidation_MemoryAlloc
"

NS_THRESHOLD=40    # % (загруженная CI-машина даёт разброс до ~35%)
BYTES_THRESHOLD=30 # %

# Режим --update: перегенерация baseline текущими значениями
UPDATE_MODE=0
if [ "${2:-}" = "--update" ]; then
    UPDATE_MODE=1
fi

echo "=== Running benchmarks (benchtime=100x) ==="
go test ./internal/benchmark -bench=".*" -benchmem -benchtime=100x -count=1 -timeout=5m | tee "$CURRENT_RAW"

# Нормализация: "Name ns bytes allocs" (без -N суффикса CPU, дробные значения округляются)
normalize() {
    awk '/^Benchmark/ {
        name=$1; sub(/-[0-9]+$/, "", name);
        # ns/op — поле 3, B/op — поле 5, allocs/op — поле 7
        printf "%s %d %d %d\n", name, $3, $5, $7
    }' "$1"
}
normalize "$CURRENT_RAW" > "$CURRENT"

if [ "$UPDATE_MODE" -eq 1 ]; then
    {
        echo "# Performance Regression Baseline — Network Scanner"
        echo "# Обновлён: $(date -u +%Y-%m-%d) (check-perf-regression.sh --update)"
        echo "# Формат: <name> <ns_per_op> <bytes_per_op> <allocs_per_op>"
        echo "# Сгенерировано: go test ./internal/benchmark -bench=\".*\" -benchmem -benchtime=100x -count=1"
        cat "$CURRENT"
    } > "$BASELINE"
    echo ""
    echo "Baseline обновлён: $BASELINE ($(wc -l < "$CURRENT") benchmarks)"
    exit 0
fi

if [ ! -f "$BASELINE" ]; then
    echo "ERROR: baseline not found: $BASELINE" >&2
    exit 1
fi

is_cpu_only() {
    echo "$CPU_ONLY_BENCHMARKS" | grep -qx "$1"
}

fail=0
checked=0
while read -r name ns bytes allocs; do
    base_line="$(grep "^$name " "$BASELINE" || true)"
    if [ -z "$base_line" ]; then
        echo "SKIP $name (нет в baseline)"
        continue
    fi
    base_ns="$(echo "$base_line" | awk '{print $2}')"
    base_bytes="$(echo "$base_line" | awk '{print $3}')"
    base_allocs="$(echo "$base_line" | awk '{print $4}')"

    checked=$((checked+1))

    # 1. allocs/op: допуск +1 alloc (шум планировщика горутин в async бенчмарках)
    if [ "$allocs" -gt $((base_allocs + 1)) ]; then
        echo "FAIL $name: allocs/op grew $base_allocs -> $allocs"
        fail=1
        continue
    fi

    # 2. bytes/op +30% только для CPU-bound (async-бенчмарки шумят по памяти)
    if is_cpu_only "$name" && [ "$base_bytes" -gt 0 ]; then
        max_bytes=$(( base_bytes * (100 + BYTES_THRESHOLD) / 100 ))
        if [ "$bytes" -gt "$max_bytes" ]; then
            echo "FAIL $name: bytes/op grew $base_bytes -> $bytes (limit $max_bytes)"
            fail=1
            continue
        fi
    fi

    # 3. ns/op +25% только для CPU-bound
    if is_cpu_only "$name" && [ "$base_ns" -gt 0 ]; then
        max_ns=$(( base_ns * (100 + NS_THRESHOLD) / 100 ))
        if [ "$ns" -gt "$max_ns" ]; then
            echo "FAIL $name: ns/op grew $base_ns -> $ns (limit $max_ns, +${NS_THRESHOLD}%)"
            fail=1
            continue
        fi
    fi

    echo "OK   $name (ns=$ns allocs=$allocs bytes=$bytes)"
done < "$CURRENT"

echo ""
echo "Checked $checked benchmarks against $BASELINE"
if [ "$fail" -ne 0 ]; then
    echo "PERF REGRESSION DETECTED"
    exit 1
fi
echo "No performance regressions detected"
