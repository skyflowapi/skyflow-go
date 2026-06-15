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

Skyflow-specific checks layered on the generic Go audit above — which already covers credential/secret
exposure (§1), input validation (§2), file/path handling (§3), TLS/transport (§4), error-information
leakage (§5), token lifecycle/TOCTOU (§7), and concurrency (§10). Do not restate those; assume the
conventions in [CLAUDE.md](../../CLAUDE.md). Use the generic audit's per-finding format and risk rating.

These name the concrete Skyflow attack surfaces the generic checks must be applied to:

- **Query API (§2):** SQL-like parameters passed to the Skyflow Query API must be sanitised or parameterised.
- **Vault custom headers (§2):** caller-supplied vault-config headers must be sanitised against CRLF header injection.
- **Vault URL construction (§2/§3):** a malformed vault ID or field name must not inject path segments into
  the API URL (path traversal / SSRF via URL building) — apply the traversal checks to URL construction, not just file I/O.
- **Token-signing key (beyond §7):** the JWT private key used for service-account token signing must be
  scoped to the signing operation and not retained afterward (§7 covers validating *received* tokens; this is the signing side).
- **Credentials in context:** if credentials pass through `context.Context`, the key must be an unexported
  type and request-scoped — flag contexts that leak credential scope upward.
