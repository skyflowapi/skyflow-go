# Contributing

## Public API contract

The public API of the `common`, `flowvault` and `v2` modules is recorded in
`api-report/<module>.api`, one line per const, var, func, type, struct field and method.
The released tags are the approved contract. The **Contract Tests** workflow
(`.github/workflows/contract-tests.yml`) checks every pull request into `main` or a
`flowvault-release/*` branch.

When you change a public package:

1. Run `scripts/contract-snapshot-update.sh` and commit the updated `api-report/` files
   with the change. CI fails if they don't match the code.
2. Read the `api-report/` diff in the pull request. A `+` line is an addition. A
   removed or changed line is a breaking change.

### Breaking changes

`common` and `flowvault` are v1 modules and `v2` is a v2 module, so a breaking change
in a released package normally needs a new major version (a new module path).

- CI fails on a breaking change unless the pull request has the `api-break-approved`
  label.
- Only an SDK code owner adds the label, after agreeing the break can't be avoided.
  The pull request description says what breaks and how users migrate, for example
  with a guide in `docs/`.
- Moving a type to `common` and keeping `type X = common/pkg.X` in the old package is
  not a break. The snapshot follows the alias.
- Changing the value of an exported constant, including an error message, counts as a
  change.
- Adding a method to an exported interface counts as a change, because it breaks code
  that implements the interface. CI still flags an interface with unexported methods
  (marked `// has unexported methods` in the snapshot), but code outside the module
  can't implement it, so approving that change is usually safe.

Packages under `internal/` and generated code are not part of the API.
