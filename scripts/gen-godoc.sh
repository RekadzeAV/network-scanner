#!/usr/bin/env bash
# Docs generation: godoc HTML для внутренних пакетов (M3).
#
# Генерирует статический HTML-сайт godoc в docs/dev/godoc/.
# Требует: Go 1.25+ и `go install golang.org/x/tools/cmd/godoc@latest`
# (или использует `go doc` для текстового дампа, если godoc недоступен).
#
# Usage:
#   ./scripts/gen-godoc.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
OUT_DIR="$REPO_ROOT/docs/dev/godoc"

echo "==> Generating godoc into $OUT_DIR"
mkdir -p "$OUT_DIR"

if command -v godoc >/dev/null 2>&1; then
    cd "$REPO_ROOT"
    godoc -http=:0 -goroot=. -path="$REPO_ROOT" -write_index_to="$OUT_DIR" >/dev/null 2>&1 || true
    echo "==> godoc index written to $OUT_DIR"
    exit 0
fi

echo "WARNING: godoc not installed; falling back to 'go doc' text dump." >&2
DUMP="$OUT_DIR/GODOC_DUMP.txt"
{
    echo "# godoc dump generated $(date -u +%Y-%m-%dT%H:%M:%SZ)"
} > "$DUMP"

cd "$REPO_ROOT"
for dir in internal/*/; do
    pkg="$(basename "$dir")"
    if compgen -G "$dir*.go" > /dev/null; then
        import_path="network-scanner/internal/$pkg"
        {
            echo ""
            echo "===== $import_path ====="
            echo ""
        } >> "$DUMP"
        go doc -all "$import_path" >> "$DUMP" 2>&1 || true
    fi
done

echo "==> Fallback dump written to $DUMP"
echo "    Install godoc for full HTML: go install golang.org/x/tools/cmd/godoc@latest"
