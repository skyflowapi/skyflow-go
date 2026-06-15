---
description: Structural code smell analysis for skyflow-go — runs the generic Go smell catalogue then applies skyflow-go-specific rules.
paths:
  - "**/*.go"
  - "**/go.mod"
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"
context: fork
---

## Instructions

@../../../common/.claude/commands/go/code-smell.md

---

## Skyflow Go SDK — Repo-Specific Smell Rules

Apply these in addition to the generic Go smell catalogue above. Where a rule below conflicts with a generic rule, the repo-specific rule wins.

### Generated code boundary
- Never flag smells inside `v2/internal/generated/` — Fern/OpenAPI-generated, excluded from analysis. If a smell originates there, report it as a generation issue, not a hand-fix.

### Messages and constants (misplaced literals)
- Inline string literals for log/error messages are a smell — they belong as constants in `v2/utils/messages/`.
- SDK-internal string constants belong in `v2/internal/constants/constants.go`; error-code strings in `v2/utils/error/error_codes.go`.
- Treat `goconst`-style repeated literals as a duplication smell — extract a named constant.

### Misplaced logic
- Validation logic outside `v2/internal/validation/` is misplaced — request structs are data holders, not validators.
- Business logic in `utils/common/` is misplaced — that package holds data types only; logic belongs in `internal/`.
- Options structs (custom headers, redaction type, etc.) mixed into request structs — they must stay separate.

### Dead / boundary code
- New feature code or non-trivial refactors landing under `v1/` (maintenance mode, EOL 2026-10-31) — flag as misplaced; new work belongs in `v2/`.

### Test smells (Ginkgo v2 + Gomega)
- `It` blocks with no `Expect` assertion; focused specs (`FIt`/`FDescribe`) left in; pending/skipped specs with no tracking comment.
