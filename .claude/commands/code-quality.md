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

Run the generic Go pipeline above from `v2/`, applying the repo conventions in
[CLAUDE.md](../../CLAUDE.md) (pipeline commands, module/coverage rules). This command only maps a
repo-rule violation to a gate verdict: any coverage gap on new/modified v2 code, or any
`golangci-lint` (`v2/.golangci.yml`) violation, is a **blocker** → `NEEDS FIXES`.
