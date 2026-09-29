---
name: code-review
description: Full code review — language patterns, naming, error handling, test coverage, then runs /code-smell and /code-security. Works for any language; Go-specific rules applied automatically for .go files.
paths:
  - "**/*.go"
  # - "**/*.py"        # uncomment for Python projects
  # - "**/*.ts"        # uncomment for TypeScript projects
  # - "**/*.js"        # uncomment for JavaScript projects
  # - "**/*.sh"        # uncomment for shell scripts
  # - "**/<EXT>"       # add other extensions as needed
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"   # replace with your generated-code path
  # - "**/node_modules/**"       # uncomment for JS/TS projects
  # - "**/<GENERATED_DIR>/**"    # add other auto-generated dirs as needed
context: fork
---

You are a senior engineer performing a thorough code review.

## Pre-requisite

Before starting the review, confirm `/code-quality` has been run and passed (build, lint, tests, coverage). If it has not been run, run it now before proceeding.

## Scope

Use `$ARGUMENTS` to determine scope:

**1. No argument (empty)** — review only files changed in the current working diff (staged + unstaged) vs HEAD:
```bash
git diff HEAD --name-only | grep -v 'vendor\|generated\|node_modules'
git diff --cached --name-only | grep -v 'vendor\|generated\|node_modules'
```
If that returns nothing (clean working tree), fall back to files changed on the current branch vs `main`:
```bash
git diff main...HEAD --name-only | grep -v 'vendor\|generated\|node_modules'
```

**2. A file or directory path** — review only that specific path. Do not expand scope beyond what was passed.

**3. `full`** — review the entire codebase on the current branch. Do not diff against any other branch. Enumerate all source files recursively:
```bash
find . -name "*.go" | grep -v 'vendor\|internal/generated\|node_modules'
```

Detect the primary language(s) from the files in scope and apply the matching language-specific checks in Step 1.

---

## Step 1 — Language Pattern Review

Review all files in scope against best practices for their language. Check every rule category below. Apply the **General** rules to all files, then apply the matching **Language-specific** section.

Group findings by file and produce a table:

```
### path/to/file.ext

| Severity | Line | Finding |
|----------|------|---------|
| Critical | 42   | error silently discarded |
| Bug      | 87   | integer overflow possible on 32-bit platform |
| Quality  | 103  | magic number 3600 — extract to a named constant |
```

**Severities:**

| Level | Meaning |
|---|---|
| **Critical** | Data loss, silent failure, security risk — must fix before merge |
| **Bug** | Wrong behaviour, incorrect output — must fix before merge |
| **Edge Case** | Unhandled input that will cause runtime failure — fix before merge |
| **Quality** | Maintainability issue, naming violation, missing pattern — fix before merge |

---

### General rules (all languages)

**Naming**
- Names are accurate and proportional to scope: short locals, descriptive exports
- No single-letter names outside loop counters or well-known math variables
- Boolean names start with `is`, `has`, `should`, `can`, or similar
- Constants are distinguishable from variables by name or casing convention for the language

**Error / exception handling**
- No errors or exceptions silently swallowed without a comment
- Error messages include enough context to diagnose without a debugger
- No `exit`/`abort`/`process.exit` in library code
- No `panic` / `throw` / `raise` in library code except for truly unrecoverable programmer errors

**API / interface design**
- Functions with more than 4 parameters use a config/options struct
- Public constructors return errors for fallible initialization
- Exported/public functions are documented

**Test coverage**
- Every exported / public function has at least a happy-path and an error-path test
- Test names describe the condition under test, not just the function name
- No sleep-based synchronization

---

### Go-specific rules (`.go` files only)

**Naming**
- Exported names use Go-standard abbreviations: `ID` not `Id`, `URL` not `Url`, `API` not `Api`
- Acronyms are fully uppercase: `HTTPClient`, `JSONPayload`
- Receiver names are consistent within a type

**Error handling**
- Errors wrapped with `fmt.Errorf("doing X: %w", err)` — not re-returned raw when context would help
- No bare `log.Fatal` / `os.Exit` in library packages

**Interface and API design**
- Interfaces defined at the consumer, not the producer
- All public API functions accept `context.Context` as their first parameter

**Concurrency**
- Goroutines have clear ownership and shutdown paths
- Shared mutable state protected by `sync.Mutex` or channels
- `sync.WaitGroup.Add` called before goroutine launch; `Done` via `defer`

**Tests**
- Table-driven tests used where multiple similar cases exist
- `t.Parallel()` where appropriate
- Use the project's established test framework (standard `testing`, Ginkgo, testify, etc.)


---

## Step 1.5 — Runtime Safety

Check every function in scope for conditions that cause panics or silent failures at runtime. These are distinct from linter findings — they require reading control flow:

| Risk | Check |
|---|---|
| Panic | `x.(T)` type assertion without ok check — must use `x, ok := x.(T)` |
| Panic | Nil map write: `m[k] = v` where `m` may be nil — initialize first |
| Panic | Slice/array index access without a preceding bounds check |
| Silent failure | Error return discarded with `_` or call result ignored entirely |
| Resource leak | `http.Response.Body`, files, DB connections — `defer close` missing on all paths including error paths |
| Panic | Send on a nil or already-closed channel |
| Panic | Goroutine with no `recover()` — especially background workers spawned at startup |
| Panic / wrong result | Integer division where denominator could be zero from user input or computed value |
| Process exit in library | `os.Exit` or `log.Fatal` — must only appear in `main` |
| Data race | `sync.Mutex` or `sync.WaitGroup` copied by value — pass by pointer or embed in a struct passed by pointer |
| Timer leak | `time.After` inside a loop — leaks timers until they fire; use `time.NewTimer` + `Reset` + `Stop` |
| Data race | `sync.WaitGroup.Add()` called inside the goroutine it counts — always call `Add` before `go` |
| Key collision | `context.WithValue` key is a built-in or exported type — use unexported struct type to prevent cross-package collision |

Report as **Edge Case** or **Bug** severity in the per-file table from Step 1.

---

## Step 2 — Code Smell Analysis

Read the file `.claude/commands/code-smell.md` and follow all of its instructions for the same files in scope. Produce its full output (per-file smell table + smell summary + recommendation).

---

## Step 3 — Security Audit

Read the file `.claude/commands/code-security.md` and follow all of its instructions for the same files in scope. Produce its full output (per-finding blocks + summary table + overall risk rating).

---

## Final Verdict

After all three steps, close with:
1. A tech-debt summary table grouped by category (Patterns / Error Handling / Naming / Tests / Smells / Security)
2. A verdict: `APPROVE` / `APPROVE WITH FIXES` / `REQUEST CHANGES`
3. Remind: run `/code-quality` again after any fixes before merging.
