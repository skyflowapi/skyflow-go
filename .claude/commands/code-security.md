---
name: code-security
description: Security audit — credential exposure, input validation, path traversal, HTTP security, token lifecycle, dependency CVEs.
paths:
  - "**/*.go"
  - "**/go.mod"
  # - "**/<EXT>"       # add other extensions as needed
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"   # replace with your generated-code path
  # - "**/<GENERATED_DIR>/**"    # add other auto-generated dirs as needed
context: fork
---

You are a security engineer auditing a Go codebase for vulnerabilities.

## Audit Scope

Use `$ARGUMENTS` to determine target files. If none provided, run:
```bash
git diff main...HEAD --name-only | grep '\.go$' | grep -v 'vendor\|generated'
```

## Security Checks

### 1. Credential and secret exposure (Critical)
- Tokens, API keys, passwords, and private keys must never appear in logs, error messages, or `fmt.Sprintf` output
- Struct fields carrying secrets must not be passed to `log.*`, `fmt.Printf`, or similar
- No hardcoded credentials or secrets in source — use environment variables or a secrets manager

### 2. Input validation (High)
- All string inputs from callers must be validated for empty/nil before use
- File paths passed to `os.Open` / `os.ReadFile` / `os.WriteFile` must not allow path traversal (`../`)
- JSON strings parsed with `json.Unmarshal` must check and return decode errors
- Integer inputs used as slice indices or sizes must be bounds-checked

### 3. File and path handling (High)
- File paths must be sanitized with `filepath.Clean` and checked for `..` traversal sequences before any file I/O
- Output directories passed by callers must be validated before use in `filepath.Join` + write operations
- `os.Open` must be followed by `defer file.Close()` or replaced with `os.ReadFile`

### 4. HTTP security (Medium)
- All sensitive API calls must use HTTPS — verify no `http://` scheme is accepted for authenticated endpoints
- Authorization headers must not be logged at any log level
- HTTP clients must have `Timeout` configured — `http.DefaultClient` must not be used for external calls
- TLS certificate verification must not be disabled (`InsecureSkipVerify: true`)

### 5. Error information leakage (Medium)
- Error messages must not include raw server response bodies that could contain sensitive data or PII
- Stack traces must not be surfaced to external callers
- Internal implementation details must not leak through error strings returned to callers

### 6. Context cancellation (Medium)
- Long-running operations (HTTP requests, polling loops, file I/O) must respect `context.Context`
- `time.Sleep` inside loops must use `select` with `ctx.Done()` instead
- Goroutines must not outlive their owning context

### 7. Authentication and token lifecycle (Medium)
- Cached tokens must be checked for expiry before reuse
- Token refresh code paths must be safe under concurrent goroutines
- Signed/encrypted tokens must be validated before use — not just checked for presence

### 8. Dependency vulnerabilities (Low)
- Note any dependencies in `go.mod` with known CVEs
- Check for outdated major versions of security-sensitive packages (e.g. `crypto`, `jwt`, `tls`)

### 9. Concurrency safety (Medium)
- Shared mutable state must be protected by `sync.Mutex`, `sync.RWMutex`, or channels
- `sync/atomic` must be used correctly — operations on the same value must all be atomic
- Maps must not be read/written concurrently without a lock

## Output Format

For each finding:

```
### path/to/file.go : line N

**Severity:** Critical / High / Medium / Low / Info
**Risk:** What an attacker or failure mode could cause
**Trigger:** Input or code path that triggers the vulnerability
**Fix:** Concrete remediation with code example
**CWE:** CWE-NNN
```

End with a summary table and overall risk rating.
