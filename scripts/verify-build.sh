#!/usr/bin/env bash
# verify-build.sh
# Проверяет, что артефакты сборки созданы и имеют корректное расширение.
# На Windows (Git Bash/WSL) исполняемые файлы ДОЛЖНЫ иметь суффикс .exe.
# На Linux/macOS расширение не требуется (или отсутствует).
#
# Использование:
#   ./scripts/verify-build.sh
#   ./scripts/verify-build.sh build
#
# Exit code: 0 при успехе, 1 если хотя бы один артефакт отсутствует.

set -euo pipefail

DIR="${1:-build}"

case "$(uname -s 2>/dev/null || echo unknown)" in
    MINGW*|MSYS*|CYGWIN*) EXE_SUFFIX=".exe" ;;
    *)                    EXE_SUFFIX="" ;;
esac

echo "=== Verify build artifacts ==="
echo "Directory: $DIR"
echo "Expected suffix: '${EXE_SUFFIX}'"

if [ ! -d "$DIR" ]; then
    echo "[FAIL] Директория сборки не найдена: $DIR" >&2
    exit 1
fi

failed=0
verified=()

for base in network-scanner network-scanner-gui; do
    name="${base}${EXE_SUFFIX}"
    path="$DIR/$name"

    if [ ! -f "$path" ]; then
        echo "[FAIL] Отсутствует артефакт: $path" >&2
        failed=1
        continue
    fi

    if [ ! -s "$path" ]; then
        echo "[FAIL] Пустой артефакт: $path" >&2
        failed=1
        continue
    fi

    echo "[OK]   $name ($(wc -c < "$path" | tr -d ' ') bytes)"
    verified+=("$name")
done

if [ "$failed" -ne 0 ]; then
    echo ""
    echo "Verification FAILED. Соберите заново: make build" >&2
    exit 1
fi

echo ""
echo "Verification OK: ${verified[*]}"
exit 0
