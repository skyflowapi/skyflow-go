# skyflow-go — Project Conventions

Single source of truth for repo-wide conventions. The `.claude/commands/*` review/quality
commands rely on these facts instead of restating them; apply them through each command's lens.

## Module layout

- **`v2/` is the active module** (own `go.mod`, module path `github.com/skyflowapi/skyflow-go/v2`).
  All new work is v2-only. Run every Go tool **from `v2/`**:
  ```bash
  cd v2 && go build ./...
  cd v2 && golangci-lint run ./...                       # config: v2/.golangci.yml
  cd v2 && go test ./...                                  # Ginkgo specs run under go test
  cd v2 && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep -v "100.0%"
  cd v2 && govulncheck ./...
  cd v2 && go mod verify && git diff --exit-code go.mod go.sum
  ```
- **The repo root is the v1 module — deprecated, EOL 2026-10-31** (maintenance-only: security/bug
  fixes). There is no `v1/` subdirectory; the root `go.mod` is v1. Build/test it only when a change
  touches it; do not require new coverage there. Flag any new feature or non-trivial refactor against v1.

## Generated code

- `v2/internal/generated/` is Fern/OpenAPI-generated. **Never edit it by hand.** It is excluded from
  lint and coverage expectations. If a bug lives there, report it as a generation issue — don't patch.

## Naming & cross-SDK public-API contract

These keys/casings are part of the public API contract shared across all Skyflow SDKs. They
intentionally deviate from standard Go acronym casing — a violation is a contract-breaking finding.

- **Acronym casing:** `Id` not `ID` (`VaultId`, `ClusterId`, `ConnectionId`); `Url` not `URL`
  (`DownloadUrl`, `BaseVaultUrl`); `Api` not `API` (`ApiKey`). The `revive` var-naming allowlist
  for these lives in `v2/.golangci.yml`; use `//revive:disable-next-line:var-naming` only for genuine
  outliers it doesn't cover.
- **Credential JSON keys:** `ClientId`/`TokenUri`/`KeyId` struct fields → JSON `clientId`/`tokenUri`/`keyId`.
- **Record identifier:** `SkyflowId` in every response struct (Insert, Update, Get, Query, BulkDelete) —
  never `skyflow_id`, `skyflowId`, or `SkyflowID`.
- **Query response:** tokenized field is `TokenizedData` (always present — empty slice, never nil/absent).
- **BearerTokenOptions:** role list field is `RoleIds` (not `RoleIDs`/`RoleId`/`Roles`).
- **Errors field:** every response struct always includes an exported `Errors []SkyflowError`, serialized
  even on success (empty slice, never nil/absent).
- **Backward compatibility:** removing or renaming any exported identifier is contract-breaking and
  requires a major version bump. Watch struct fields callers may embed or copy.

## Code placement

- `utils/common/` holds **data types only** — no logic.
- Business logic goes in `internal/`, not exported packages.
- All input validation goes in `v2/internal/validation/` (`ValidateXxxRequest()`), before any HTTP/file
  use. Request structs are data holders, not validators. Options structs (headers, redaction type, …)
  stay separate from request structs.
- No inline string literals for log/error messages — extract constants:
  - log/error messages → `v2/utils/messages/`
  - SDK-internal constants → `v2/internal/constants/constants.go`
  - error-code strings → `v2/utils/error/error_codes.go`
  - `goconst` flags repeated literals — fix by extracting a named constant.

## Error handling (SDK contract)

- Public methods return `*skyflowError.SkyflowError` — never raw `error` or bare `nil` on failure.
- Call `logger.Error(...)` (from `v2/utils/logger`) before every `SkyflowError` return.
- Wrap underlying errors with cause context (`Cause: err`); never swallow — always propagate.
- No error message exposes sensitive data (tokens, credentials, PII).
- No `fmt.Println`, `log.Print*`, or `panic` in SDK code.

## API surface

- Every public API method takes `context.Context` as its first parameter. If credentials pass through
  a context, the key must be an unexported type and request-scoped.

## Tests

- **Ginkgo v2 + Gomega** (`Describe`/`It`/`Expect`); run via `cd v2 && go test ./...`.
- **100% statement and branch coverage** for all new/modified v2 code — any gap on changed code is a blocker.
- Don't mock the production struct under test — use interface substitution or function-variable injection.
- Mock external HTTP via `httptest.Server`, not by patching internal helpers.
