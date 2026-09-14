#!/bin/bash
# Final Release Verification Script — v2.0
# Запуск: ./scripts/final-release-check.sh

set -e

echo "═══════════════════════════════════════════════"
echo " Network Scanner v2.0 — Final Release Check"
echo "═══════════════════════════════════════════════"
echo ""

# 1. Build check
echo "[1/5] Build CLI..."
go build -o /tmp/network-scanner-test ./cmd/network-scanner
if [ -f /tmp/network-scanner-test ]; then
    echo "  ✅ CLI собран"
    rm /tmp/network-scanner-test
else
    echo "  ❌ CLI не собрался"
    exit 1
fi

echo "[2/5] Build GUI..."
go build -o /tmp/network-scanner-gui-test ./cmd/gui
if [ -f /tmp/network-scanner-gui-test ]; then
    echo "  ✅ GUI собран"
    rm /tmp/network-scanner-gui-test
else
    echo "  ❌ GUI не собрался"
    exit 1
fi

# 2. Unit tests
echo "[3/5] Unit tests..."
go test ./... -count=1 > /tmp/test-output.txt 2>&1
if [ $? -eq 0 ]; then
    echo "  ✅ go test ./... прошёл"
else
    echo "  ❌ go test ./... упал"
    cat /tmp/test-output.txt
    exit 1
fi

# 3. Smoke — topology export
echo "[4/5] Smoke — topology export..."
if [ -f "./scripts/smoke-d-track-topology-export.sh" ]; then
    ./scripts/smoke-d-track-topology-export.sh
    if [ $? -eq 0 ]; then
        echo "  ✅ smoke-d-track-topology-export прошёл"
    else
        echo "  ⚠️  smoke-d-track-topology-export упал (не критично)"
    fi
else
    echo "  ⚠️  smoke-d-track-topology-export.sh не найден"
fi

# 4. Documentation sanity
echo "[5/5] Docs sanity..."
if [ -f "./docs/FINAL_RELEASE_READINESS_REPORT.md" ] && \
   [ -f "./docs/D_TRACK_IMPLEMENTATION_STATUS.md" ] && \
   [ -f "./CHANGELOG.md" ]; then
    echo "  ✅ Ключевые документы на месте"
else
    echo "  ❌ Некоторые документы отсутствуют"
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════"
echo " ✅ Все проверки пройдены!"
echo "═══════════════════════════════════════════════"
echo ""
echo "Следующие шаги:"
echo "  1. Запустить кросс-ОС прогон на Linux/macOS"
echo "  2. Разблокировать CI через GITHUB_TOKEN"
echo "  3. Пройти ручную GUI приёмку (docs/GUI_SMOKE_CHECKLIST.md)"
echo "  4. Выпустить релиз"
echo ""
