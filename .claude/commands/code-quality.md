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

## Skyflow Go SDK — Quality Pipeline

Run the generic pipeline above from the **`v2/`** module (its own `go.mod`); the `golangci-lint`
config is `v2/.golangci.yml`. Module layout, the v1 (root) boundary, and test/coverage rules are in
[CLAUDE.md](../../CLAUDE.md). The generic pipeline's blocker and `NEEDS FIXES` verdict rules apply as-is.
