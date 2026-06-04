---
name: code-quality
description: Quality pipeline — build, lint, test, coverage check. Pass a package path to target a specific package.
paths:
  - "**/*.go"
  - "**/go.mod"
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"
context: fork
---

Run the Go quality pipeline.

Use `$ARGUMENTS` to target a specific package path (e.g. `./internal/validation/...`). If empty, run against all packages (`./...`).

> Run `go test ./... 2>&1 | grep -E "FAIL|ok"` first to capture the baseline before reporting failures.

## Coverage Requirements

100% statement and branch coverage on all code written or modified in this session. Flag any gap as a blocker — **NEEDS FIXES** if coverage drops below 100% on code you touched.

---

## Pipeline

### Step 1 — Build
```bash
go build ./... 2>&1 | tail -20
```
Expected: no output (clean build). Report any errors.

### Step 2 — Lint
```bash
golangci-lint run ./... 2>&1 | tail -30
```
Report any lint violations. These are blockers. If no `golangci-lint` config is present, run `go vet ./...` as a fallback.

### Step 3 — Tests
If `$ARGUMENTS` is set:
```bash
go test $ARGUMENTS -v 2>&1 | tail -60
```
Otherwise:
```bash
go test ./... 2>&1 | tail -40
```
Report: packages tested, FAIL lines, PASS summary. Flag any failures beyond the known pre-existing baseline.

### Step 4 — Coverage analysis
```bash
go test ./... -coverprofile=coverage.out 2>/dev/null && go tool cover -func=coverage.out | grep -v "100.0%" | tail -30
```
Lines printed are functions below 100% coverage. For every function touched in this session:
- Verify a corresponding test exists
- Verify positive path (happy path) AND negative path (error/validation rejection) are covered
- Verify every branch (`if`/`switch`/`for`) is exercised

List all gaps. Any gap on code written or modified in this session is a **blocker**.

### Step 5 — Edge case identification
For any package below 100% coverage, identify missing scenarios:
- Nil / empty inputs to public functions
- Invalid enum / sentinel values
- Concurrent access (if state is shared)
- Error paths (network failure, missing input, expired state)

Write concrete test stubs for each gap (framework used by the project).

### Step 6 — Report

```
| Step             | Status    | Notes                             |
|------------------|-----------|-----------------------------------|
| Build            | ✅ / ❌   | ...                               |
| Lint             | ✅ / ❌   | ...                               |
| Tests            | ✅ / ❌   | N passed, M failed                |
| Coverage (100%)  | ✅ / ❌   | list functions with gaps          |
```

Conclude with **READY TO MERGE** or **NEEDS FIXES** and a prioritised fix list.
