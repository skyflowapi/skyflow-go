# Contributing

## Public API compatibility

The **Contract Tests** workflow (`.github/workflows/contract-tests.yml`) checks every pull
request into `main`, a `release/x.x.x` branch (skyvault, the `v2` module) or a
`flowvault-release/x.x.x` branch. It compares the public API of each module at the pull
request's base commit with the pull request, using
[apidiff](https://pkg.go.dev/golang.org/x/exp/cmd/apidiff), the Go team's API
compatibility checker. The released tags are the approved contract.

- Modules checked: `v2`, `common` and `flowvault`, each in its own job, **Contract
  Tests (v2)**, **Contract Tests (common)** and **Contract Tests (flowvault)**, so the
  pull request's checks show which module broke. A module missing on a branch passes as
  skipped. The list is the workflow's matrix and `DEFAULT_MODULES` in
  `scripts/apidiff/apidiff.sh`; keep them in sync.
- Every package except `internal/` packages and `main` packages is compared. A removed
  package is an incompatible change, and a new package is a compatible one.
- `internal/` packages are left out because users can't import them. That only holds
  while no internal type is reachable from the public API, so the job also fails if an
  exported name exposes one: a public alias of an internal type, an exported or embedded
  field, a function parameter or result, a method, or a type argument or constraint. The
  `api-break-approved` label doesn't apply; move the type to a public package instead.
  `scripts/apidiff/internalcheck` does this check.
- Incompatible changes fail the module's job and appear on the pull request's Checks tab
  as an annotation, **Breaking API change in &lt;module&gt;**, listing each change (the first
  20; the job summary has all of them). With the `api-break-approved` label the
  annotation is a warning and the job passes. Compatible changes, such as a new field or
  function, only appear in the job summary.

Run the same check locally before you push (it compares with the merge-base with
`origin/main` by default, or with any ref you pass):

```sh
scripts/apidiff/apidiff.sh                       # all modules
scripts/apidiff/apidiff.sh -r origin/flowvault-release/26.11.0 flowvault
```

apidiff is pinned in `scripts/apidiff/go.mod` and verified through `go.sum`. To upgrade
it, run `go get -tool golang.org/x/exp/cmd/apidiff@<version>` in `scripts/apidiff`, then
`go mod tidy`.

### Breaking changes

A breaking change in a released package normally needs a new major version (a new
module path, such as `v2` → `v3`). apidiff follows Go's compatibility rules; see its
[README](https://github.com/golang/exp/blob/master/apidiff/README.md) for what counts
as incompatible.

- CI fails on an incompatible change unless the pull request has the
  `api-break-approved` label.
- Only an SDK code owner adds the label, after agreeing the break can't be avoided.
  The pull request description says what breaks and how users migrate, for example
  with a guide in `docs/`.
- Changing the value of an exported constant, including an error message, is
  reported as incompatible. Today the SDK version is part of the exported message
  constants, so the pull request that bumps `SDK_VERSION` needs the label.
- Moving a type to another package behind `type X = other/pkg.X` keeps user code
  compiling, but apidiff reports it as incompatible on the pull request that makes the
  move. Approve it with the label after checking that the moved type is unchanged.
