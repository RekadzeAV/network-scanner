#!/bin/bash
# save_benchmarks.sh — Запуск бенчмарков и сохранение результатов в artifacts
# Используется в CI для контроля регрессий производительности

set -e

ARTIFACT_DIR="release-build/benchmarks"
mkdir -p "$ARTIFACT_DIR"

echo "=== Запуск бенчмарков ==="
go test -bench=. -benchmem -run=^$ -count=3 ./internal/scanner/ > "$ARTIFACT_DIR/scanner.txt" 2>&1
go test -bench=. -benchmem -run=^$ -count=3 ./internal/network/ > "$ARTIFACT_DIR/network.txt" 2>&1
go test -bench=. -benchmem -run=^$ -count=3 ./internal/topology/ > "$ARTIFACT_DIR/topology.txt" 2>&1

echo "=== Результаты сохранены в $ARTIFACT_DIR ==="
ls -la "$ARTIFACT_DIR"

echo ""
echo "=== Краткая сводка ==="
for f in "$ARTIFACT_DIR"/*.txt; do
    echo "--- $(basename $f) ---"
    grep -E "^Benchmark|^PASS|^ok" "$f" | head -20
    echo ""
done
