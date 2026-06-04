---
name: code-smell
description: Structural smell analysis + spell check — long functions, dead code, misplaced validation, deep nesting, magic numbers. Does not check patterns or security.
paths:
  - "**/*.go"
  - ".claude/**/*.md"
  - "docs/**/*.md"
  # - "**/<EXT>"       # add other extensions as needed
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"   # replace with your generated-code path
  # - "**/<GENERATED_DIR>/**"    # add other auto-generated dirs as needed
context: fork
---

You are a senior Go engineer performing a code smell analysis.

## Scope

Use `$ARGUMENTS` to determine scope:
- A file or directory path — analyse only that path
- Empty / default — analyse files changed on current branch vs `main`:
  ```bash
  git diff main...HEAD --name-only | grep '\.go$' | grep -v 'vendor\|generated'
  ```

---

## Spell check

Before analysing smells, run cspell on the files in scope:

```bash
npx cspell --no-progress "**/*.go" ".claude/**/*.md" "docs/**/*.md" 2>&1 | grep "Unknown word"
```

Report any spelling violations at **Smell** severity in the per-file table. Add legitimate project-specific or domain terms to `.cspell.json` rather than marking them as typos.

---

## What Are Code Smells

Code smells are structural signals — they do not necessarily mean the code is broken, but they indicate areas of technical debt, reduced readability, or future maintenance risk. All findings are reported at **Smell** severity and do not block merge unless they indicate a design violation.

---

## Smell Catalogue

### Function & File Size

**Long function** — any function over 40 lines.
Signal: the function is doing too much. Candidate for decomposition into named helpers.

**Long file** — any file over 400 lines.
Signal: the file may be taking on too many responsibilities. Check if it can be split by concern.

**Large parameter list** — more than 4 parameters on a function.
Signal: consider a config/options struct or grouping related parameters.

---

### Responsibility Violations

**Business logic in data structs**
Structs that carry data should not contain conditional logic, field transformations, or computation beyond simple field access. Flag any such methods.

**Validation outside a dedicated validation layer**
If the project has a dedicated validation package or layer, any `if x == "" { return err }` guard outside that layer is misplaced.

**Message strings inline in application logic**
String literals passed directly to `log.*`, `fmt.Errorf`, or error constructors that could be named constants are a responsibility smell — especially when the same or similar strings appear more than once.

---

### Control Flow

**Deep nesting** — more than 3 levels of `if` / `for` / `switch` nesting.
Signal: extract inner blocks to named private functions or use early returns.

**Long if-else / switch chains** — more than 4 branches on the same condition.
Signal: consider a map-based dispatch or interface polymorphism.

**Repeated nil checks**
Multiple consecutive nil guards on the same value that could be collapsed or replaced with an early return.

---

### Data

**Magic numbers**
Literal integers or durations (`25`, `64`, `3600`, `time.Duration(60)`) without a named constant. Extract to a package-level `const`.

**Repeated string literals**
Any string appearing more than once that should be a named constant. The `goconst` linter flags these automatically — report literals that goconst would catch.

**Temporary field**
A struct field only populated in certain code paths, nil the rest of the time. Should be a local variable or function parameter instead.

---

### Dead Code

**Unused unexported functions** — unexported functions with no callers in the same package.

**Unused imports** — any `import` not referenced in the file (the compiler catches these, but flag if seen in generated or build-tagged files).

**Unreachable code** — code after `return` / `panic` in the same branch.

**Commented-out code** — blocks of commented code without a `// TODO: [ticket]` explaining why they are kept.

---

### Interface Design

**Interface with too many methods** — more than 5 methods on a single interface.
Signal: violates Interface Segregation; split into narrower role interfaces.

**`interface{}` / `any` where a concrete type would work**
Using `any` as a parameter or return type loses type safety. Flag unless the function is genuinely type-agnostic (e.g. a serialiser).

**Interface defined in the same package as its only implementation**
Interfaces should be defined where they are consumed, not where they are implemented. An interface with one implementation in the same package is likely premature abstraction.

---

### Constructor Patterns

**Exported struct with required fields but no `New*` constructor**
Callers can construct a zero-value instance that panics at first use. Every exported struct with required fields must have a `NewXxx(...)` constructor that validates and returns an initialised value.

**Constructor returning concrete type where an interface would decouple callers**
If callers depend on the concrete type, swapping implementations requires changing all call sites. Return the narrowest interface that satisfies callers.

---

### Writing Functions

For each changed function, apply this 5-step checklist:
1. Can you follow what the function does in one reading? If yes, stop.
2. High cyclomatic complexity (deep nesting, many branches)? → candidate for decomposition into named helpers.
3. Would a common data structure (map, tree, queue) make this simpler and more robust?
4. Hidden untested dependencies or values that could be factored into parameters?
5. Is the function name the best possible? Brainstorm 3 alternatives; flag if the current name is weaker.

---

### Test Smells

**`t.Skip(...)` / `Skip(...)` without a `// TODO: [ticket]` comment**
Skipped tests silently rot with no path back. Every skip must reference a ticket.

**`time.Sleep` in tests**
Non-deterministic; use channels, `sync.WaitGroup`, or Gomega `Eventually`/`Consistently` helpers instead.

**Test body with no assertions**
A test that calls functions but never asserts results always passes and catches nothing.

**100% coverage required**
Flag any new or modified function that has uncovered statements or branches. Report as **Smell** with the specific function and line range.

---

### Comments

**Explains what, not why**
A comment that restates what the code does (`// loop over records`) adds no value. Only flag comments that explain the *what* without explaining *why*.

**Stale comment**
A comment that contradicts the current code — references a removed parameter, old function name, or changed behaviour.

---

## Output Format

Group findings by file:

```
### path/to/file.go

| Smell                     | Line | Detail                                                        |
|---------------------------|------|---------------------------------------------------------------|
| Long function             | 42   | processResponse() is 67 lines — decompose                    |
| Magic number              | 103  | Literal 64 — extract to a named constant                     |
| Inline string literal     | 210  | "user id is required" appears 3 times — extract to const     |
| Dead code                 | 315  | unexported buildAuthHeader() has no callers                  |
```

End with a **Smell Summary** table:

```
| Category              | Count | Files affected              |
|-----------------------|-------|-----------------------------|
| Long functions        | 2     | response.go                 |
| Magic numbers         | 3     | validation.go               |
| Inline string literal | 5     | controller.go               |
| Dead code             | 2     | helpers.go                  |
```

Close with a recommendation: **CLEAN** / **MINOR DEBT** / **SIGNIFICANT DEBT** and a one-sentence summary.
