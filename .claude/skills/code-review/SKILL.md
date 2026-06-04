---
name: code-review
description: Project context — conventions, design principles, error handling, build commands. Loaded automatically for all source files.
paths:
  - "**/*.go"
  - "**/*.md"
  - "**/*.sh"
  - "**/*.py"
  - "**/*.ts"
  - "**/*.js"
  - "**/go.mod"
  - "**/go.sum"
  - "**/Makefile"
exclude:
  - "**/internal/generated/**"
  - "**/vendor/**"
  - "**/node_modules/**"
---

> **Template setup** — Replace the placeholders below to match your project.
>
> | Placeholder | Current value | Description |
> |---|---|---|
> | `{{MODULE_PATH}}` | `github.com/skyflowapi/skyflow-go/v2` | Full Go module path |
> | `{{GENERATED_PATH}}` | `v2/internal/generated/` | Auto-generated path — never edit |
> | `{{ERROR_TYPE}}` | `*skyflowError.SkyflowError` | SDK error type for all public methods |
> | `{{ERROR_PACKAGE}}` | `v2/utils/error` | Package that defines the error type |
> | `{{MESSAGES_PACKAGE}}` | `v2/utils/messages/` | Package where all log/error strings live |
> | `{{CONSTANTS_FILE}}` | `v2/internal/constants/constants.go` | SDK-internal constants file |
> | `{{TEST_FRAMEWORK}}` | `Ginkgo v2 + Gomega` | Test framework (`Describe` / `It` / `Expect`) |
> | `{{JIRA_PREFIX}}` | `SK` | Jira project key prefix (e.g. `SK` → `SK-1234`) |
> | `{{MODULE_DIR}}` | `v2` | Subdirectory where the Go module lives (use `.` if root) |

# Project Context

See `CLAUDE.md` in the project root for the full project context. Key reference points:

- **Module:** `{{MODULE_PATH}}`
- **Generated paths:** `{{GENERATED_PATH}}` — never edit
- **Error type:** `{{ERROR_TYPE}}` (package `{{ERROR_PACKAGE}}`)
- **Messages package:** `{{MESSAGES_PACKAGE}}` — all log/error strings live here
- **Constants package:** `{{CONSTANTS_FILE}}`
- **Test framework:** {{TEST_FRAMEWORK}} (`Describe` / `It` / `Expect`)
- **Ticket pattern:** `[A-Z]{1,10}-[0-9]+` (e.g. `{{JIRA_PREFIX}}-1234`)

## Go Naming Conventions (intentional deviations for cross-SDK consistency)
See `CLAUDE.md` — Naming Conventions section for the full list of `Id`/`Url`/`Api` suffix deviations and example field names. Do not flag these as naming violations during review.

The `revive` `var-naming` allowlist in `{{MODULE_DIR}}/.golangci.yml` covers all listed exceptions.

## Custom Slash Commands
- `/code-review [path|full]` — full review: patterns + code smells + security (any language)
- `/code-smell [path|full]` — standalone structural smell analysis
- `/code-security [path|full]` — standalone security audit
- `/sdk-sample <feature>` — generate a sample file for a feature
- `/code-quality [path|full]` — build → lint → test → coverage check
- `/git-commit <description>` — Jira-aware commit with quality gate

When reviewing or writing code, always apply the conventions and rules from `CLAUDE.md`.
