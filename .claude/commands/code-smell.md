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

Apply the repo conventions in [CLAUDE.md](../../CLAUDE.md) as smell signals, on top of the generic Go
smell catalogue above. Where a repo rule conflicts with a generic rule, the repo rule wins.

- **Misplaced code** (CLAUDE.md → *Code placement*) is a smell: validation outside
  `v2/internal/validation/`, logic in `utils/common/`, options structs mixed into request structs,
  inline log/error/constant literals (`goconst`-style repeats are a duplication smell — extract a constant).
- **Boundary code:** new feature / non-trivial refactor under the v1 (root) module is misplaced — new
  work belongs in `v2/`.
- **Test smells (Ginkgo v2 + Gomega):** `It` blocks with no `Expect`; focused specs (`FIt`/`FDescribe`)
  left in; pending/skipped specs with no tracking comment.
