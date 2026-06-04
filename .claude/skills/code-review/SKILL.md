---
name: code-review
description: Project context — conventions, design principles, error handling, build commands. Loaded automatically for all source files.
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
  - "**/Makefile"
  # Add file types for other languages used in your project:
  # - "**/*.{{EXTRA_EXT_1}}"   # e.g. *.py, *.ts, *.sh
  # - "**/*.{{EXTRA_EXT_2}}"
exclude:
  - "{{GENERATED_GLOB}}"        # glob for auto-generated code, e.g. **/internal/generated/**
  - "**/vendor/**"
  - "**/node_modules/**"
  # - "{{EXTRA_EXCLUDE_1}}"     # add more project-specific excludes as needed
---

> **Template setup** — Fill in the placeholders for your project. Remove any line marked *optional* that does not apply.
>
> | Placeholder | Description | Example |
> |---|---|---|
> | `{{MODULE_PATH}}` | Full Go module path | `github.com/org/repo` |
> | `{{MODULE_DIR}}` | Subdirectory where the Go module lives | `v2` or `.` (root) |
> | `{{GENERATED_PATH}}` | Auto-generated code path — never edit manually | `internal/generated/` |
> | `{{ERROR_TYPE}}` | Custom error type returned by public methods *(optional — remove if using std `error`)* | `*MyError` |
> | `{{ERROR_PACKAGE}}` | Package that defines the custom error type *(optional)* | `pkg/apierror` |
> | `{{MESSAGES_PACKAGE}}` | Package where all log/error message strings live *(optional)* | `internal/messages/` |
> | `{{CONSTANTS_FILE}}` | File for SDK-internal string constants *(optional)* | `internal/constants/constants.go` |
> | `{{TEST_FRAMEWORK}}` | Test framework used by the project | `Ginkgo v2 + Gomega` or `testing` |
> | `{{TEST_STYLE}}` | Test block syntax | `` `Describe`/`It`/`Expect` `` or `TestXxx(t *testing.T)` |
> | `{{JIRA_PREFIX}}` | Issue tracker key prefix *(optional — remove if N/A)* | `SK` or `GH` |
> | `{{GENERATED_GLOB}}` | Glob pattern matching the auto-generated code directory | `**/internal/generated/**` |
> | `{{EXTRA_EXT_1}}`, `{{EXTRA_EXT_2}}` | Additional file extensions to load as context *(optional)* | `py`, `ts`, `sh` |
> | `{{EXTRA_EXCLUDE_1}}` | Additional paths to exclude from context *(optional)* | `**/testdata/**` |

# Project Context

See `CLAUDE.md` in the project root for full context. Key quick-references:

- **Module:** `{{MODULE_PATH}}`
- **Generated paths:** `{{GENERATED_PATH}}` — never edit
- **Error type:** `{{ERROR_TYPE}}` (package `{{ERROR_PACKAGE}}`) — *remove if using standard `error`*
- **Messages package:** `{{MESSAGES_PACKAGE}}` — all log/error strings live here — *remove if N/A*
- **Constants file:** `{{CONSTANTS_FILE}}` — *remove if N/A*
- **Test framework:** {{TEST_FRAMEWORK}} ({{TEST_STYLE}})
- **Ticket pattern:** `[A-Z]{1,10}-[0-9]+` (e.g. `{{JIRA_PREFIX}}-1234`) — *remove if N/A*

## Go Naming Conventions

Follow standard Go conventions (`ID`, `URL`, `API`, etc.) unless `CLAUDE.md` documents intentional deviations. If deviations exist, they are listed there — do not flag them as naming violations during review.

The `revive` `var-naming` allowlist in `{{MODULE_DIR}}/.golangci.yml` covers any documented exceptions.

## Slash Commands

- `/code-review [path|full]` — full review: patterns + code smells + security
- `/code-smell [path|full]` — standalone structural smell analysis
- `/code-security [path|full]` — standalone security audit
- `/code-quality [path|full]` — build → lint → test → coverage check
- `/git-commit <description>` — ticket-aware commit with quality gate
- `/sdk-sample <feature>` — generate a sample file *(remove if N/A)*

When reviewing or writing code, always apply the conventions and rules from `CLAUDE.md`.
