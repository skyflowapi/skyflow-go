---
description: Security audit for skyflow-go — runs the generic Go security audit then applies skyflow-go-specific rules.
paths:
  - "**/*.go"
  - "**/go.mod"
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"
context: fork
---

## Instructions

@../../../common/.claude/commands/go/code-security.md

---

## Skyflow Go SDK — Repo-Specific Security Rules

Apply these in addition to the generic Go security checks above. Focus exclusively on security impact. Report findings using the per-finding block format, severities, summary table, and overall risk rating defined in the generic audit above — these rules add *checks*, not a new output format.

### SkyflowError leakage
- `SkyflowError.Message` returned to callers must not include raw server response bodies that may contain field-level data or PII.
- Stack traces and internal system paths/DB details must not surface through `SkyflowError`.
- Internal error codes/states must be safe to expose externally.

### Skyflow credential and token handling
- `Credentials` struct fields (`clientId`, `tokenUri`, `keyId`, API keys, bearer tokens, private keys) must never be logged — even at DEBUG — nor printed via `%v` / `%+v`.
- The JWT private key used for service-account token signing must be scoped to the signing operation and not retained beyond it.
- Cached bearer tokens must be checked for expiry before each API call and held in memory only — never written to disk, logs, or error messages.
- If Skyflow credentials are passed through `context.Context`, the context key must be an unexported type and the scope must be controlled (flag request-scoped contexts that leak upward).

### Skyflow API input handling
- Every caller-supplied string (vault ID, table/column name, query string, file path) must be validated in `internal/validation/` before reaching an HTTP request or file operation.
- A malformed vault ID or field name must not be able to inject path segments into the API URL (path traversal / SSRF via URL construction).
- SQL-like parameters passed to the Skyflow Query API must be sanitised or parameterised.
- File paths for `UploadFile` must be validated against directory traversal (`../`).
- Caller-supplied custom headers (vault config) must be sanitised to prevent CRLF header injection.

### HTTP / TLS (Skyflow specifics)
- Bearer tokens and `Authorization` headers must be transmitted only over TLS — never plain HTTP — and never logged at any level.
- `tls.Config.InsecureSkipVerify` must never be `true`.
- Every HTTP client used for vault/connection calls must set a `Timeout`.

### Concurrency on shared credential/token state
- Any credential or token cache shared across goroutines must be synchronised (`sync.Mutex` / `sync.RWMutex` / atomic).
- Guard against a TOCTOU race where two goroutines simultaneously refresh a token.
- Vault config maps must not be mutated after the client is constructed (concurrent map read/write data race).
