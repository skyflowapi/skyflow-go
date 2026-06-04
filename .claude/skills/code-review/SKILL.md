---
name: project-context
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

# Skyflow Go SDK — Project Context

See `CLAUDE.md` in the project root for the full project context. Key reference points:

- **Module:** `github.com/skyflowapi/skyflow-go/v2`
- **Generated paths:** `v2/internal/generated/` — never edit
- **Error type:** `*skyflowError.SkyflowError` (package `v2/utils/error`)
- **Messages package:** `v2/utils/messages/` — all log/error strings live here
- **Constants package:** `v2/internal/constants/constants.go`
- **Test framework:** Ginkgo v2 + Gomega (`Describe` / `It` / `Expect`)
- **Ticket pattern:** `[A-Z]{1,5}-[0-9]+` (e.g. `SK-1234`)

## Go Naming Conventions (intentional deviations for cross-SDK consistency)
- `VaultId` not `VaultID` — `Id` suffix throughout
- `DownloadUrl` not `DownloadURL` — `Url` suffix throughout
- `ApiKey` not `APIKey`
- Revive `var-naming` allowlist in `v2/.golangci.yml` covers these

## Custom Slash Commands
- `/code-review [path]` — full review: patterns + code smells + security (any language)
- `/code-smell` — standalone structural smell analysis
- `/code-security` — standalone security audit
- `/sdk-sample <feature>` — generate a sample file for a feature
- `/code-quality [./path/...]` — build → lint → test → coverage check
- `/git-commit <description>` — Jira-aware commit with quality gate

When reviewing or writing code, always apply the conventions and rules from `CLAUDE.md`.
