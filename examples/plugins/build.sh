#!/usr/bin/env bash
# Сборка примеров плагинов Network Scanner (-buildmode=plugin).
#
# Каждый плагин собирается в <plugin>/dist/<entry>.so рядом с plugin.json.
# Требования: gcc/clang (CGO), linux/darwin.
#
# Использование:
#   ./build.sh                 # собрать все
#   ./build.sh onlyup          # собрать один
set -euo pipefail

ONLY="${1:-}"

PLUGIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${PLUGIN_ROOT}/../../.." && pwd)"

# имя каталога -> entry из plugin.json
declare -A PLUGINS=(
  ["onlyup"]="onlyup.so"
  ["json-exporter"]="jsonexporter.so"
  ["alert-report"]="alertreport.so"
)

if ! command -v gcc >/dev/null 2>&1; then
  echo "gcc не найден: сборка -buildmode=plugin требует CGO" >&2
  exit 1
fi

cd "${REPO_ROOT}"
for dir in onlyup json-exporter alert-report; do
  if [[ -n "${ONLY}" && "${dir}" != "${ONLY}" ]]; then
    continue
  fi
  src="${PLUGIN_ROOT}/${dir}"
  out="${src}/dist"
  mkdir -p "${out}"

  echo "==> Сборка ${dir}"
  CGO_ENABLED=1 go build -buildmode=plugin -o "${out}/${PLUGINS[$dir]}" "${src}"
  cp "${src}/plugin.json" "${out}/plugin.json"
  echo "    готово: ${out}"
done

echo
echo "Укажите каталог с dist/plugin.json в NETWORK_SCANNER_PLUGINS_DIR"
echo "и запустите: network-scanner plugins list"
