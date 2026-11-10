# Contributing

## Public API contract

The public API of each module under contract is recorded in `api-report/<module>.api`,
one line per const, var, func, type, struct field and method. A module is under
contract when its snapshot is committed (today: `v2`). The released tags are the
approved contract. The **Contract Tests** workflow (`.github/workflows/contract-tests.yml`)
checks every pull request into `main` or a `flowvault-release/*` branch.

When you change a public package:

1. Run `scripts/contract-snapshot-update.sh` and commit the updated `api-report/` files
   with the change. CI fails if they don't match the code.
2. Read the `api-report/` diff in the pull request. A `+` line is an addition. A
   removed or changed line is a breaking change.

To put a new module under contract, run `scripts/contract-snapshot-update.sh <module>`
and commit `api-report/<module>.api`. Deleting a snapshot doesn't take a module out of
the check: CI then reports its whole API as removed.

### Breaking changes

A breaking change in a released package normally needs a new major version (a new
module path, such as `v2` → `v3`).

- CI fails on a breaking change unless the pull request has the `api-break-approved`
  label.
- Only an SDK code owner adds the label, after agreeing the break can't be avoided.
  The pull request description says what breaks and how users migrate, for example
  with a guide in `docs/`.
- Moving a type to another package and keeping `type X = other/pkg.X` in the old
  package is not a break. The snapshot follows the alias.
- Changing the value of an exported constant, including an error message, counts as a
  change.
- Adding a method to an exported interface counts as a change, because it breaks code
  that implements the interface. CI still flags an interface with unexported methods
  (marked `// has unexported methods` in the snapshot), but code outside the module
  can't implement it, so approving that change is usually safe.

Packages under `internal/` and generated code are not part of the API.
