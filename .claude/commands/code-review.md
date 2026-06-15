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

## Skyflow Go SDK — Repo-Specific Review Rules

Apply these checks on **every** review in addition to the generic Go rules above. Where a rule below conflicts with a generic rule, the repo-specific rule wins (it is part of the cross-SDK public-API contract). Use the severities, per-file tables, and final verdict defined in the generic review above — these rules add *checks*, not a new output format.

### Generated code boundary
- Flag any edit to `v2/internal/generated/` — these are Fern/OpenAPI-generated and must never be edited manually.
- If a bug exists in generated code, report it; do not patch it by hand.

### Naming conventions (cross-SDK consistency)
These intentionally deviate from standard Go acronym casing. Flag any violation:
- `Id` suffix, **not** `ID` — e.g. `VaultId`, `ClusterId`, `ConnectionId`
- `Url` suffix, **not** `URL` — e.g. `DownloadUrl`, `BaseVaultUrl`
- `Api` prefix, **not** `API` — e.g. `ApiKey`
- Use `//revive:disable-next-line:var-naming` only for genuine outliers not coverable by the allowlist in `v2/.golangci.yml`.

### Cross-SDK nomenclature (Go-specific public-API contract)
These keys are part of the public API contract across all Skyflow SDKs — treat any violation as a top-severity (contract-breaking) finding.

**Credential JSON key fields** (in `Credentials` struct and related types):
- Must use `ClientId` (not `ClientID`), `TokenUri` (not `TokenURI`), `KeyId` (not `KeyID`)
- Go struct field tags must produce JSON keys `clientId`, `tokenUri`, `keyId`
- Check all credential-related structs in `utils/common/` and `internal/`

**Response map keys — universal across all vault operations**:
- Skyflow record identifier key must be `SkyflowId` in all response structs (Insert, Update, Get, Bulk Delete, Query)
- Must NOT be `skyflow_id` (snake_case), `skyflowId` (camelCase), or `SkyflowID` (all-caps acronym)
- Check `InsertResponse`, `UpdateResponse`, `GetResponse`, `QueryResponse`, `BulkDeleteResponse`, and any other response types

**Query response**:
- Tokenized data field must be `TokenizedData` (not `tokenized_data`, `tokenizedData`, or `TokenizedDataField`)
- `TokenizedData` must always be present in query response records (empty slice, not absent/nil)

**BearerTokenOptions**:
- Role list field must be `RoleIds` (not `RoleIDs`, `RoleId`, or `Roles`)

**Always-present response fields**:
- All response structs must always include an `Errors` field (never omit it, even on success — emit empty slice, not nil/absent)
- Verify the field is exported (`Errors []SkyflowError`) and included in JSON serialization

### Error handling (SDK contract)
- All public methods must return `*skyflowError.SkyflowError` — never raw `error` or bare `nil` on failure.
- `logger.Error(...)` (from `v2/utils/logger`) must be called before every `SkyflowError` return.
- Errors wrapped with cause context (`Cause: err`).
- No `fmt.Println`, `log.Print*`, or `panic` anywhere in SDK code.
- Never swallow errors — always propagate to the caller.
- No error message may expose sensitive data (tokens, credentials, PII).

### Messages and constants
- No inline string literals for log/error messages — all such strings must be constants in `v2/utils/messages/`.
- All SDK-internal string constants go in `v2/internal/constants/constants.go`.
- Error code strings go in `v2/utils/error/error_codes.go`.
- `goconst` will flag repeated literals; fix by extracting a constant.

### Request / Response patterns
- Request structs are data holders only — all validation belongs in `v2/internal/validation/ValidateXxxRequest()`.
- Options structs (custom headers, redaction type, etc.) must stay separate from request structs.
- Every response struct must have an `Errors` field (slice, nil when empty per serialization above).

### Internal vs public boundary
- New business logic placed in `internal/` rather than exported packages.
- `utils/common/` contains data types only, not logic.

### Input validation
- All user-supplied inputs (vault IDs, field names, tokens, file paths) validated in `internal/validation/` before use.
- Validation errors returned via `SkyflowError`, never panics.

### Context
- Every public API method accepts `context.Context` as its first parameter.

### Tests
- Framework: Ginkgo v2 + Gomega — use `Describe` / `It` / `Expect` blocks.
- 100% statement and branch coverage required for all new/modified code.
- No mocking the production struct under test — use interface substitution or function-variable injection.
- External HTTP calls mocked via `httptest.Server`, not by patching internal helpers.
- Tests compilable and passing: `cd v2 && go test ./...`

### Backward compatibility
- Any removed or renamed exported identifier is contract-breaking — requires a major version bump.
- Watch struct fields that callers may embed or copy.

### v1 maintenance boundary
- `v1/` is in maintenance mode (EOL: October 31, 2026) — security and bug fixes only.
- Flag any new feature or non-trivial refactor proposed against `v1/`.
- New customer-facing functionality belongs in `v2/` only.

---

After completing the code review above, also run `/security-review $ARGUMENTS` to perform a security audit of the same scope, and append its findings below the code review output.
