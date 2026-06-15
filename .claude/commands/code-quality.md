---
description: Quality pipeline for skyflow-go — runs the generic Go pipeline (build, lint, tests, coverage, vuln, tidy) then applies skyflow-go-specific rules.
paths:
  - "**/*.go"
  - "**/go.mod"
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"
context: fork
---

## Instructions

@../../../common/.claude/commands/go/code-quality.md

---

## Skyflow Go SDK — Repo-Specific Quality Rules

Apply these in addition to the generic Go pipeline above. Where a rule below conflicts with a generic rule, the repo-specific rule wins.

### Module layout — run against v2
- Active development is the **`v2/`** module (its own `go.mod`). Run every pipeline step from there:
  ```bash
  cd v2 && go build ./...
  cd v2 && golangci-lint run ./...        # config: v2/.golangci.yml
  cd v2 && go test ./...
  cd v2 && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep -v "100.0%"
  cd v2 && govulncheck ./...
  cd v2 && go mod verify && git diff --exit-code go.mod go.sum
  ```
- **`v1/` (root module) is maintenance-only** (EOL 2026-10-31): build/test it only when the change touches v1; do not require new coverage there. New work is v2-only.

### Generated code
- Exclude `v2/internal/generated/` from coverage and lint expectations — it is Fern/OpenAPI-generated.

### Tests & coverage
- Test framework is **Ginkgo v2 + Gomega**; run via `go test ./...` (Ginkgo specs execute under it).
- 100% statement **and** branch coverage is required for all new/modified v2 code — any gap on changed code is a **blocker** (`NEEDS FIXES`).

### Lint
- `golangci-lint` uses `v2/.golangci.yml` (includes the `revive` var-naming allowlist for the cross-SDK `Id`/`Url`/`Api` casing). Treat its violations as blockers.
