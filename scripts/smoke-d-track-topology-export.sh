#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Smoke: D-track topology export =="

TMP_DIR="$(mktemp -d)"
SMOKE_BIN=""
cleanup() {
  rm -rf "${TMP_DIR:-}" "${SMOKE_BIN:-}"
}
trap cleanup EXIT

SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-dtrack-smoke.XXXXXX")"
go build -o "$SMOKE_BIN" ./cmd/network-scanner

BASE_ARGS=(--network 127.0.0.1/32 --timeout 1 --ports 1-8 --topology)
JSON_OUT="$TMP_DIR/topology.json"
GRAPHML_OUT="$TMP_DIR/topology.graphml"
PNG_OUT="$TMP_DIR/topology.png"
PNG_FALLBACK_JSON="$TMP_DIR/topology.json"

"$SMOKE_BIN" "${BASE_ARGS[@]}" --output-format json --output-file "$JSON_OUT" >/dev/null 2>&1
"$SMOKE_BIN" "${BASE_ARGS[@]}" --output-format graphml --output-file "$GRAPHML_OUT" >/dev/null 2>&1

test -f "$JSON_OUT"
test -f "$GRAPHML_OUT"

if ! rg -q '"Devices"' "$JSON_OUT"; then
  echo "Smoke failed: JSON export misses Devices field"
  exit 1
fi
if ! rg -q "<graphml" "$GRAPHML_OUT"; then
  echo "Smoke failed: GraphML export misses graphml root"
  exit 1
fi

python3 - "$JSON_OUT" "$GRAPHML_OUT" <<'PY'
import json
import sys
import xml.etree.ElementTree as ET

json_path = sys.argv[1]
graphml_path = sys.argv[2]

def node_identity(hostname, ip, mac):
    for value in (hostname, ip, mac):
        v = (value or "").strip().lower()
        if v:
            return v
    return ""

def undirected_edge(a, b):
    a = (a or "").strip().lower()
    b = (b or "").strip().lower()
    return "<->".join(sorted([a, b]))

with open(json_path, "r", encoding="utf-8") as f:
    data = json.load(f)

json_nodes = []
id_by_json_node = {}
for _, dev in (data.get("Devices") or {}).items():
    ident = node_identity(dev.get("Hostname"), dev.get("IP"), dev.get("MAC"))
    if ident:
        json_nodes.append(ident)

json_edges = []
for link in (data.get("Links") or []):
    src = link.get("Source") or {}
    dst = link.get("Target") or {}
    src_ident = node_identity(src.get("Hostname"), src.get("IP"), src.get("MAC"))
    dst_ident = node_identity(dst.get("Hostname"), dst.get("IP"), dst.get("MAC"))
    json_edges.append(undirected_edge(src_ident, dst_ident))

tree = ET.parse(graphml_path)
root = tree.getroot()
ns = {"g": "http://graphml.graphdrawing.org/xmlns"}
nodes = root.findall(".//g:node", ns)
edges = root.findall(".//g:edge", ns)

label_by_id = {}
graphml_nodes = []
for node in nodes:
    node_id = (node.attrib.get("id") or "").strip()
    label = ""
    for d in node.findall("g:data", ns):
        if (d.attrib.get("key") or "").strip() == "label":
            label = (d.text or "").strip().lower()
            break
    label_by_id[node_id] = label
    if label:
        graphml_nodes.append(label)

graphml_edges = []
for edge in edges:
    src = label_by_id.get((edge.attrib.get("source") or "").strip(), "")
    dst = label_by_id.get((edge.attrib.get("target") or "").strip(), "")
    graphml_edges.append(undirected_edge(src, dst))

if sorted(json_nodes) != sorted(graphml_nodes):
    print(f"Smoke failed: node set mismatch json={sorted(json_nodes)} graphml={sorted(graphml_nodes)}")
    sys.exit(1)
if sorted(json_edges) != sorted(graphml_edges):
    print(f"Smoke failed: edge set mismatch json={sorted(json_edges)} graphml={sorted(graphml_edges)}")
    sys.exit(1)
PY

set +e
"$SMOKE_BIN" "${BASE_ARGS[@]}" --output-format png --output-file "$PNG_OUT" >"$TMP_DIR/png.log" 2>&1
PNG_EXIT=$?
set -e
if [ "$PNG_EXIT" -ne 0 ]; then
  echo "Smoke failed: png export command returned non-zero"
  cat "$TMP_DIR/png.log"
  exit 1
fi

if [ -f "$PNG_OUT" ]; then
  echo "PNG export produced image via Graphviz."
elif [ -f "$PNG_FALLBACK_JSON" ]; then
  if ! rg -q "Graphviz недоступен|fallback JSON" "$TMP_DIR/png.log"; then
    echo "Smoke failed: fallback JSON exists but expected message missing"
    cat "$TMP_DIR/png.log"
    exit 1
  fi
  echo "PNG export fallback to JSON works when dot is unavailable."
else
  echo "Smoke failed: neither PNG nor fallback JSON was produced"
  cat "$TMP_DIR/png.log"
  exit 1
fi

echo "Smoke passed: D-track topology exports are healthy."
