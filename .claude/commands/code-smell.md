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

## Skyflow Go SDK — Smell Rules

Skyflow-specific smells, on top of the generic Go catalogue above — which already covers inline/repeated
literals, misplaced validation, business logic in data structs, no-assertion tests, and untracked skips.
Apply the placement rules in [CLAUDE.md](../../CLAUDE.md) → *Code placement* to name the concrete targets:

- **Misplaced code:** logic in the data-only `utils/common/`, validation outside `v2/internal/validation/`,
  or options structs folded into request structs.
- **v1 boundary:** new feature / non-trivial refactor under the v1 (root) module — new work belongs in `v2/`.
- **Ginkgo focused specs:** `FIt`/`FDescribe` left in — silently narrows the suite to a subset.
