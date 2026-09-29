---
description: SDK-focused security audit covering credentials, tokens, input validation, HTTP hardening, and data leakage
---

Perform a security-focused review of the changes on the current branch. Run `git diff main...HEAD --name-only` to get the list of changed files, filter out any path containing `internal/generated/` (auto-generated OpenAPI code — do not review), then run `git diff main...HEAD` and read each remaining file in full. For each finding use severity labels: `[CRITICAL]`, `[HIGH]`, `[MEDIUM]`, `[LOW]`, or `[INFO]`.

Focus exclusively on security impact. Do not comment on style, naming, or non-security logic unless it directly enables an attack.

---

## 1. Credential and Secret Handling

- Are credentials (API keys, bearer tokens, private keys, passwords) ever logged — even at DEBUG level?
- Are credential structs or maps printed via `%v` / `%+v` in log lines or error messages?
- Are secrets held in `string` fields when they should be zeroed after use? (`string` cannot be zeroed; use `[]byte` and `defer` wipe for high-sensitivity material)
- Is any credential written to a temp file or disk path that a low-privileged process could read?
- Is the JWT private key (RSA/EC) used for service-account token signing properly scoped and not stored beyond the signing operation?
- Are Skyflow credentials passed through `context.Context`? (Acceptable only if the context scope is controlled; flag if request-scoped contexts leak upward)

## 2. Token Security (JWT / Bearer)

- Is token expiry checked **before** each API call, not just at construction?
- Is the JWT signature verified using the correct algorithm? (`alg: none` or algorithm confusion?)
- Are `exp`, `iat`, and `iss` claims validated on tokens received from external sources?
- Is clock skew handled with a small tolerance (e.g., 10–30 s) to avoid replay near expiry boundary?
- Are bearer tokens transmitted only over TLS (never over plain HTTP)?
- Are tokens cached in memory only — never written to disk, logs, or error messages?

## 3. Input Validation and Injection

- Is every caller-supplied string (vault ID, table name, column name, query string, file path) validated before it reaches an HTTP request or file operation?
- Could a malformed vault ID or field name inject unexpected path segments into the API URL? (path traversal / SSRF via URL construction)
- Are SQL-like query parameters passed to the Skyflow Query API sanitised or parameterised?
- Are file paths for `UploadFile` validated to prevent directory traversal (`../` sequences)?
- Is user-controlled data ever placed in a log format string (format string injection)?
- Is any `os.Exec` or shell command constructed from user input?

## 4. HTTP and TLS Hardening

- Is `tls.Config.InsecureSkipVerify` ever set to `true`? (CRITICAL — removes MITM protection)
- Is a request timeout set on all HTTP clients (`http.Client{Timeout: ...}`)? (missing timeout enables slowloris / resource exhaustion)
- Are redirect policies controlled? (default Go client follows redirects; redirects to `http://` from `https://` downgrade protection)
- Are response bodies always closed via `defer resp.Body.Close()`? Unclosed bodies cause connection leaks.
- Is response body size bounded before reading into memory (`io.LimitReader`)? Unbounded read enables memory exhaustion.
- Are custom headers set by callers sanitised to prevent header injection (CRLF in values)?

## 5. Error and Log Leakage

- Do error messages returned to callers ever include raw server responses that may contain sensitive data?
- Are stack traces exposed in `SkyflowError.Message` visible to end users?
- Do log lines at any level emit token values, field-level data, or PII record content?
- Are internal error codes/states safe to expose externally (no internal system paths, DB details)?

## 6. Concurrency and State Safety

- Is any credential or token cache shared across goroutines without proper synchronisation (`sync.Mutex`, `sync.RWMutex`, or atomic)?
- Could a race condition cause two goroutines to simultaneously refresh a token, generating duplicate network calls or a TOCTOU window on expiry?
- Are maps used as shared configuration mutated after the client is constructed? (concurrent map read/write = data race + undefined behaviour)

## 7. Dependency and Supply Chain

- Does `go.mod` introduce a new direct dependency? If so:
  - Is it from a reputable source with an active maintainer?
  - Does it have known CVEs? (check `govulncheck` or OSV)
  - Is the minimum version pinned (`require` without a floating pseudo-version)?
- Are indirect dependencies updated in a way that pulls in a vulnerable version?

## 8. File and OS Operations

- Are temp files created with `os.CreateTemp` (random suffix, not predictable path)?
- Are temp files cleaned up with `defer os.Remove(...)` even on error paths?
- Are file permissions restricted (`0600` for sensitive files, never world-readable)?
- Is any user-supplied filename used in `os.Open` / `os.Create` without path sanitisation?

## 9. Context and Cancellation

- Can a cancelled context cause a partial write to the Skyflow vault that leaves data in an inconsistent state? Document the SDK's guarantee (or lack thereof) in godoc.
- Are context values used to pass security-sensitive data (credentials, tokens)? If so, is the key an unexported type to prevent collision?

---

## Summary Table

| Category | Critical | High | Medium | Low | Info |
|---|---|---|---|---|---|
| Credentials & Secrets | | | | | |
| Token Security | | | | | |
| Input Validation / Injection | | | | | |
| HTTP / TLS Hardening | | | | | |
| Error & Log Leakage | | | | | |
| Concurrency / State | | | | | |
| Dependencies | | | | | |
| File / OS Operations | | | | | |
| Context & Cancellation | | | | | |

**Overall risk**: Critical / High / Medium / Low — and a one-sentence recommendation.
