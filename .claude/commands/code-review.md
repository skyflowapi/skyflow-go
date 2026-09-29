---
description: SDK-focused Go code review covering API design, idioms, backward compatibility, and test coverage
paths: !/v2/internal/generated/
---

Perform a thorough code review from an **SDK maintainer's perspective**.

**Always exclude files under `internal/generated/` from the review — these are auto-generated from OpenAPI specs and must not be edited manually.**

Determine the scope based on the argument `$ARGUMENTS`:

- **No argument** (default): Review only the files changed on this branch. Run `git diff main...HEAD --name-only` to get the list, filter out any path containing `internal/generated/`, then run `git diff main...HEAD` for the full diff and read each non-generated changed file in full.
- **`full review`**: Review the entire SDK codebase. Run `find v2 -name "*.go" -not -path "*/generated/*" -not -name "*_test.go"` to enumerate all non-generated source files, then read them all.
- **A file or directory path**: Review only that path. If it is a directory, run `find <path> -name "*.go" -not -path "*/generated/*"` to list all Go files within it (excluding generated), then read them all.

After reading the relevant files, apply the checklist below.

Structure the review under these headings. Under each heading list findings as `[BLOCKER]`, `[WARNING]`, or `[SUGGESTION]`. Skip any heading that has no findings.

---

## 1. Public API Surface

- Are new exported types, functions, and methods named following Go conventions?
  - PascalCase for exported identifiers; camelCase for unexported
  - Acronyms fully capitalised when exported (`VaultID`, `HTTPClient`), fully lower when unexported
  - `With*` prefix for functional option constructors
  - Receiver names short (1-2 letters), consistent across all methods of a type
- Does any change break existing callers? (renamed export, removed field, changed signature)
- Are new options/params added via functional options (`With*`) rather than positional arguments?
- Do all public operations accept `ctx context.Context` as the first parameter?
- Are new interfaces minimal and named for behaviour (verb/noun — `TokenProvider`, not `ITokenProvider`)?
- Do exported types have godoc comments that describe what, not how?

## 1a. Cross-SDK Nomenclature (Go-specific)

Verify these specific naming requirements from the cross-SDK specification. Flag any violation as `[BLOCKER]` since they are part of the public API contract across all Skyflow SDKs.

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

## 2. Error Handling

- Are all public methods returning `*error.SkyflowError` (never raw `error` or `nil` on failure)?
- Are errors wrapped with cause context (`Cause: err`)?
- Are new error message strings added to `utils/messages/` rather than inlined?
- Does any error message expose sensitive data (tokens, credentials, PII)?
- Is `panic` used in library code? (never acceptable in an SDK — convert to error return)

## 3. Go Idioms and Best Practices

- No naked `return` in long functions; named returns only when they substantially aid clarity
- No use of `fmt.Print*` or `log.*` in library code (use project logger)
- Goroutines: are all goroutines guarded by context cancellation or a done channel?
- No unexported global mutable state (race conditions in concurrent SDK use)
- Nil guards on receivers and map/slice fields before use
- `defer` used correctly — no deferred calls inside loops
- No shadowed `err` variables across if-blocks
- Proper use of `errors.Is` / `errors.As` rather than type-asserting raw errors
- Unused imports, unused variables — run `go vet` confirms clean?

## 4. Internal vs Public Boundary

- Is new business logic placed in `internal/` rather than exported packages?
- Are generated files under `internal/generated/` untouched (edited by hand)?
- Does `utils/common/` only contain data types, not logic?

## 5. Input Validation

- Are all user-supplied inputs (vault IDs, field names, tokens, file paths) validated in `internal/validation/` before use?
- Validation errors returned via `SkyflowError`, not panics
- No use of raw user input in log messages without sanitisation

## 6. Test Coverage

- Does every new exported function/method have at least one `It` block (happy path)?
- Are all validation-error branches covered?
- Are external HTTP calls mocked via `httptest.Server`, not by patching internal helpers?
- New `DescribeTable` entries for any table-driven scenario added?
- Tests compilable and passing: `cd v2 && go test ./...`

## 7. Backward Compatibility

- Are any exported identifiers removed or renamed? (BLOCKER — requires major version bump)
- Are struct fields that callers may embed or copy affected?
- Is `go.mod` `require` bumped for a dependency with a breaking change in its own API?

## 8. Concurrency and Resource Safety

- HTTP clients and vault configs: are they safely shared across goroutines (no mutation after construction)?
- File uploads: are temporary files and open handles cleaned up even on error paths?
- Context propagation: is `ctx` passed down to every HTTP call and blocking operation?

---

End with a **Summary** table:

| Category | Blockers | Warnings | Suggestions |
|---|---|---|---|
| Public API Surface | | | |
| Cross-SDK Nomenclature | | | |
| Error Handling | | | |
| Go Idioms | | | |
| Internal/Public Boundary | | | |
| Input Validation | | | |
| Test Coverage | | | |
| Backward Compatibility | | | |
| Concurrency/Resources | | | |

Then give an overall verdict: **Approve**, **Approve with minor changes**, or **Request changes**.

---

After completing the code review above, also run `/security-review` to perform a security audit of the same scope, and append its findings below the code review output.
