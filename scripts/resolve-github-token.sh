#!/usr/bin/env bash
# Source from repo scripts (or run standalone): fills GITHUB_TOKEN when unset via `gh auth token`.
if [[ -z "${GITHUB_TOKEN:-}" ]] && command -v gh >/dev/null 2>&1; then
  t="$(gh auth token 2>/dev/null || true)"
  if [[ -n "${t}" ]]; then
    export GITHUB_TOKEN="${t}"
  fi
fi
