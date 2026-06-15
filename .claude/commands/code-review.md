---
description: Full code review for skyflow-go — runs the generic Go review then applies skyflow-go-specific rules.
paths:
  - "**/*.go"
  - "**/go.mod"
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"
context: fork
---

## Instructions

@../../../common/.claude/commands/go/code-review.md

---

## Skyflow Go SDK — Review Rules

Review against the repo conventions in [CLAUDE.md](../../CLAUDE.md) in addition to the generic Go
rules above. Use the severities, per-file tables, and final verdict from the generic review — these
rules add *checks*, not a new output format. A violation of the cross-SDK public-API contract
(naming/nomenclature/backward-compat in CLAUDE.md) is a top-severity, contract-breaking finding.

Flag, on every review:
- **Public-API contract** (CLAUDE.md → *Naming & cross-SDK public-API contract*): acronym casing,
  credential JSON keys, `SkyflowId`/`TokenizedData`/`RoleIds`/`Errors` field shapes, and any
  removed/renamed exported identifier (major-version bump).
- **Code placement & error handling** (CLAUDE.md → *Code placement*, *Error handling*): misplaced
  validation/logic, inline message/constant literals, missing `logger.Error` before a `SkyflowError`,
  raw `error`/`nil` returns.
- **Tests** (CLAUDE.md → *Tests*): Ginkgo v2 idioms, no mocking the struct under test, and
  `httptest.Server` (not patched helpers) for external HTTP — coverage itself is gated by the generic pipeline.
- **v1 boundary:** flag new features / non-trivial refactors against the v1 (root) module.

