#!/usr/bin/env bash
# Regenerate the public API snapshots in api-report/ (the Go counterpart of the
# japicmp baselines in skyflow-java). Commit the result with the change that caused it.
#
#   scripts/contract-snapshot-update.sh           # rewrite api-report/<module>.api
#   scripts/contract-snapshot-update.sh --check   # fail if a snapshot is out of date
set -euo pipefail

MODULES=(common flowvault v2)
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOL_DIR="$ROOT/scripts/apicontract"
CHECK=false
[[ "${1:-}" == "--check" ]] && CHECK=true

BIN="$(mktemp -d)/apicontract"
trap 'rm -rf "$(dirname "$BIN")"' EXIT
(cd "$TOOL_DIR" && GOWORK=off go build -o "$BIN" .)

status=0
for m in "${MODULES[@]}"; do
  out="$ROOT/api-report/$m.api"
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
