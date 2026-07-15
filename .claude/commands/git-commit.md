---
name: git-commit
description: Stage check + optional ticket-aware commit — validates staged files and runs quality gate before committing.
context: fork
---

Create a git commit for staged changes on the current branch.

Use `$ARGUMENTS` as the commit message description. If empty, ask the user for a description before proceeding.

## Step 1 — Extract ticket ID from branch name (optional)

```bash
git rev-parse --abbrev-ref HEAD
```

If the branch name contains a ticket ID matching `[A-Z]{1,10}-[0-9]+` (e.g. `SK-1234`, `GH-42`, `PROJ-789`), extract it and prepend it to the commit message.

- `alice/SK-1234-fix-foo` → prefix `SK-1234`
- `feature/GH-42-add-auth` → prefix `GH-42`
- `fix-typo` → no ticket found, proceed without prefix

## Step 2 — Check what is staged

```bash
git status --short
git diff --cached --stat
```

If nothing is staged, list the unstaged modified files and ask the user which files to stage. Do not run `git add .` — ask for explicit paths.

The following must **never** be staged:
- `credentials.json`, `*.json` credential files, `.env`, `*.env`
- Any file containing secrets, tokens, or private keys
- Files under `vendor/` unless this is an intentional vendor update

## Step 3 — Assemble the commit message

If a ticket ID was found in the branch name:
```
<ticket-id> <description>
```

If the user provided a Conventional Commits prefix (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`), prepend it:
```
feat: SK-1234 add file upload support
fix: GH-42 handle nil token on refresh
```

Without a ticket ID:
```
fix: handle nil token on refresh
```

## Step 4 — Quality check

Before committing, confirm `/code-quality` has been run and passed (build, lint, tests, coverage on changed code). If it has not been run, ask the user whether to run it now before proceeding.

## Step 5 — Commit

```bash
git commit -m "<assembled message>"
```

Report the resulting commit SHA and the commit message first line.
