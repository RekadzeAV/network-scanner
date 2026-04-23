#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

GUI_EXE="${1:-./network-scanner-gui}"
if [[ ! -f "$GUI_EXE" ]]; then
  echo "GUI binary not found: $GUI_EXE"
  echo "Usage: $0 [path-to-network-scanner-gui]"
  exit 1
fi

echo "== Smoke: GUI resolution matrix =="
echo "Binary: $GUI_EXE"
echo
echo "Run checks manually for each profile:"
echo "  1) 1366x768 @100% and @125%"
echo "  2) 1920x1080 @125%"
echo "  3) 2560x1440 @100%"
echo "  4) 3840x2160 @150%"
echo
echo "Acceptance checklist:"
echo "  - Scan tab: controls visible, no critical clipping."
echo "  - Results tab: table/cards usable, Host Details actions accessible."
echo "  - Topology tab: preview and text both reachable in windowed mode."
echo "  - Tools tab: tool buttons and Operations controls available."
echo "  - Same functionality in windowed and fullscreen modes."
echo
echo "Launching GUI..."
"$GUI_EXE" &
echo "Close the GUI after verification and record findings in release notes."
