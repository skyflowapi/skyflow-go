#!/usr/bin/env bash
# Regenerate the public API snapshots in api-report/ (the Go counterpart of the
# japicmp baselines in skyflow-java). Commit the result with the change that caused it.
#
# A module is under contract when api-report/<module>.api exists, so the snapshots
# already committed decide which modules are checked.
#
#   scripts/contract-snapshot-update.sh              # rewrite every api-report/<module>.api
#   scripts/contract-snapshot-update.sh --check      # fail if a snapshot is out of date
#   scripts/contract-snapshot-update.sh <module>...  # add or rewrite these snapshots
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOL_DIR="$ROOT/scripts/apicontract"
CHECK=false
MODULES=()
for arg in "$@"; do
  case "$arg" in
    --check) CHECK=true ;;
    -*) echo "unknown option: $arg" >&2; exit 2 ;;
    *) MODULES+=("${arg%/}") ;;
  esac
done
if [[ ${#MODULES[@]} -eq 0 ]]; then
  for f in "$ROOT"/api-report/*.api; do
    [[ -e "$f" ]] && MODULES+=("$(basename "$f" .api)")
  done
fi
if [[ ${#MODULES[@]} -eq 0 ]]; then
  echo "no snapshots in api-report/; add one with: scripts/contract-snapshot-update.sh <module>" >&2
  exit 1
fi

BIN="$(mktemp -d)/apicontract"
trap 'rm -rf "$(dirname "$BIN")"' EXIT
(cd "$TOOL_DIR" && GOWORK=off go build -o "$BIN" .)

mkdir -p "$ROOT/api-report"
status=0
for m in "${MODULES[@]}"; do
  out="$ROOT/api-report/$m.api"
  if [[ ! -f "$ROOT/$m/go.mod" ]]; then
    echo "::error file=api-report/$m.api::$m/go.mod not found; remove api-report/$m.api if the module was deleted."
    status=1
    continue
  fi
  tmp="$(mktemp)"
  "$BIN" snapshot "$ROOT/$m" > "$tmp"
  if $CHECK; then
    if ! diff -u "$out" "$tmp" > /dev/null 2>&1; then
      echo "::error file=api-report/$m.api::api-report/$m.api is out of date; run scripts/contract-snapshot-update.sh and commit the result."
      diff -u "$out" "$tmp" || true
      status=1
    fi
    rm -f "$tmp"
  else
    mv "$tmp" "$out"
    echo "wrote api-report/$m.api ($(wc -l < "$out") lines)"
  fi
done
exit $status
