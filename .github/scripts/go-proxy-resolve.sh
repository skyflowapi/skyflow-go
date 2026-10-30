#!/usr/bin/env bash
# Asks proxy.golang.org to fetch and cache <module>@<query> and prints the module's JSON
# (Version, Origin.Hash, ...) on stdout. The proxy fetches new commits and tags from
# GitHub on demand, which can be slow or flaky, so a failed lookup is retried.
#
# Usage: go-proxy-resolve.sh <module-path> <version-or-commit>
set -euo pipefail

module=${1:?usage: go-proxy-resolve.sh <module-path> <version-or-commit>}
query=${2:?usage: go-proxy-resolve.sh <module-path> <version-or-commit>}
attempts=5
backoff_seconds=10

# Outside any module, so the query goes straight to the proxy.
cd "$(mktemp -d)"
for ((i = 1; i <= attempts; i++)); do
  if GOPROXY=https://proxy.golang.org GOFLAGS=-mod=mod go list -m -json "$module@$query"; then
    exit 0
  fi
  if ((i < attempts)); then
    sleep $((i * backoff_seconds))
  fi
done
echo "::error::proxy.golang.org could not resolve $module@$query" >&2
exit 1
