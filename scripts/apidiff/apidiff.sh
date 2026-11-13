#!/usr/bin/env bash
# Public API compatibility check.
#
# Compares the public API of each module at a base commit with the working tree, using
# golang.org/x/exp/cmd/apidiff at the version pinned in scripts/apidiff/go.mod. Every
# non-internal, non-main package is compared on its own (apidiff's module mode would also
# compare internal packages). A removed package counts as an incompatible change.
#
# Skipping internal packages is safe only while none of their types is reachable from the
# public API, so scripts/apidiff/internalcheck also fails the check if one is.
#
#   scripts/apidiff/apidiff.sh                     # all modules, base = merge-base with origin/main
#   scripts/apidiff/apidiff.sh -r origin/main v2   # one module against a given ref
#   scripts/apidiff/apidiff.sh -a warning v2       # annotate approved breaks as warnings
#
# Exit status: 0 = compatible, 1 = incompatible changes found, 2 = the check itself failed,
# 3 = the public API exposes an internal type (whatever else was found).
# When GITHUB_STEP_SUMMARY is set, the report is also written there as Markdown. In GitHub
# Actions each module with incompatible changes, or with internal types in its public API,
# also gets one annotation that names the module and lists the changes. -a sets the level
# of the incompatible-change annotations (error, the default, or warning once approved).
set -euo pipefail
trap 'exit 2' ERR

# Modules under contract. A module missing from both sides is skipped, so the same list
# works on main (v2 only) and on the flowvault release branches.
DEFAULT_MODULES=(v2 common flowvault)

# Most changes listed in one annotation; the job summary has the full report.
ANNOTATION_MAX_LINES=20

usage() {
  echo "usage: $0 [-r BASE_REF] [-a error|warning] [MODULE...]" >&2
  exit 2
}

die() {
  echo "apidiff.sh: $*" >&2
  exit 2
}

BASE_REF=""
BREAK_LEVEL=error
while getopts ":r:a:h" opt; do
  case "$opt" in
    r) BASE_REF="$OPTARG" ;;
    a) BREAK_LEVEL="$OPTARG" ;;
    *) usage ;;
  esac
done
[[ "$BREAK_LEVEL" == error || "$BREAK_LEVEL" == warning ]] || usage
shift $((OPTIND - 1))
MODULES=("$@")
[[ ${#MODULES[@]} -gt 0 ]] || MODULES=("${DEFAULT_MODULES[@]}")

command -v go >/dev/null 2>&1 ||
  die "go not found in PATH; install Go (version in scripts/apidiff/go.mod) or add it to PATH"

ROOT="$(git rev-parse --show-toplevel)"
TOOL_DIR="$ROOT/scripts/apidiff"
if [[ -z "$BASE_REF" ]]; then
  BASE_REF="$(git -C "$ROOT" merge-base HEAD origin/main)" || die "no merge-base with origin/main; pass -r BASE_REF"
fi
BASE_SHA="$(git -C "$ROOT" rev-parse --verify "$BASE_REF^{commit}")" || die "unknown base ref: $BASE_REF"

# Each module is read with the dependency versions in its own go.mod, as users get it.
export GOWORK=off GOFLAGS=-mod=readonly

WORK="$(mktemp -d)"
BASE_TREE="$WORK/base"
cleanup() {
  git -C "$ROOT" worktree remove --force "$BASE_TREE" >/dev/null 2>&1 || true
  rm -rf "$WORK"
}
trap cleanup EXIT

git -C "$ROOT" worktree add --quiet --detach "$BASE_TREE" "$BASE_SHA"
APIDIFF="$WORK/apidiff"
(cd "$TOOL_DIR" && go build -o "$APIDIFF" golang.org/x/exp/cmd/apidiff)
INTERNALCHECK="$WORK/internalcheck"
(cd "$TOOL_DIR" && go build -o "$INTERNALCHECK" ./internalcheck)

# public_packages DIR prints the import paths of the module's public packages.
public_packages() {
  local list
  list="$(cd "$1" && go list -f '{{if ne .Name "main"}}{{.ImportPath}}{{end}}' ./...)" ||
    die "go list failed in $1"
  awk 'NF && $0 !~ /\/internal(\/|$)/' <<<"$list" | sort -u
}

summary() {
  if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
    printf '%s\n' "$@" >>"$GITHUB_STEP_SUMMARY"
  fi
}

say() {
  printf '%s\n' "$@"
  summary "$@"
}

# annotate LEVEL TITLE LINES writes one GitHub Actions annotation (only inside Actions).
# LINES is cut to ANNOTATION_MAX_LINES; the job summary always has the full report.
annotate() {
  [[ "${GITHUB_ACTIONS:-}" == true ]] || return 0
  local level="$1" title="$2" lines="$3" total body
  total="$(grep -c . <<<"$lines" || true)"
  body="$(head -n "$ANNOTATION_MAX_LINES" <<<"$lines")"
  if ((total > ANNOTATION_MAX_LINES)); then
    body+=$'\n'"... and $((total - ANNOTATION_MAX_LINES)) more; see the job summary."
  fi
  # Escape as GitHub's workflow commands require: data, then the title property.
  body="${body//%/%25}"
  body="${body//$'\r'/%0D}"
  body="${body//$'\n'/%0A}"
  title="${title//%/%25}"
  title="${title//,/%2C}"
  title="${title//:/%3A}"
  printf '::%s title=%s::%s\n' "$level" "$title" "$body"
}

# incompatible_lines PKG prints the incompatible changes of one apidiff report (stdin)
# as "PKG: change".
incompatible_lines() {
  awk -v pkg="$1" '
    /^Incompatible changes:/ { on = 1; next }
    /^Compatible changes:/   { on = 0; next }
    on && sub(/^- /, "")     { print pkg ": " $0 }'
}

say "## Public API compatibility" "" "Base: \`$BASE_SHA\`" ""

broken=()
exposed=()
for m in "${MODULES[@]}"; do
  base_dir="$BASE_TREE/$m"
  head_dir="$ROOT/$m"
  if [[ ! -f "$base_dir/go.mod" && ! -f "$head_dir/go.mod" ]]; then
    echo "$m: not on this branch, skipped"
    continue
  fi
  # A module only on one side is new (every package added) or removed (every package removed).
  base_pkgs=""
  head_pkgs=""
  if [[ -f "$base_dir/go.mod" ]]; then
    base_pkgs="$(public_packages "$base_dir")"
  fi
  if [[ -f "$head_dir/go.mod" ]]; then
    head_pkgs="$(public_packages "$head_dir")"
    rc=0
    leaks="$("$INTERNALCHECK" "$head_dir")" || rc=$?
    case "$rc" in
      0) ;;
      1)
        exposed+=("$m")
        annotate error "Internal types in the public API of $m" "$leaks"
        say "### $m: internal types in the public API" "" \
          "Users can't import internal packages, so these types can't be part of the API." \
          "Move each one to a public package (or keep it out of exported names):" "" \
          '```' "$leaks" '```' ""
        ;;
      *) die "internalcheck failed for $m" ;;
    esac
  fi

  report=""
  incompatible=false
  breaks=""
  while IFS= read -r pkg; do
    [[ -n "$pkg" ]] || continue
    in_base=false
    in_head=false
    grep -qxF "$pkg" <<<"$base_pkgs" && in_base=true
    grep -qxF "$pkg" <<<"$head_pkgs" && in_head=true
    if $in_base && $in_head; then
      export_file="$WORK/export"
      (cd "$base_dir" && "$APIDIFF" -w "$export_file" "$pkg") || die "apidiff -w failed for $pkg at $BASE_SHA"
      out="$(cd "$head_dir" && "$APIDIFF" "$export_file" "$pkg")" || die "apidiff failed for $pkg"
      # apidiff exits 0 even when it finds incompatible changes, so read the report.
      grep -q '^Incompatible changes:' <<<"$out" && incompatible=true
    elif $in_base; then
      out=$'Incompatible changes:\n- package removed'
      incompatible=true
    else
      out=$'Compatible changes:\n- package added'
    fi
    [[ -n "$out" ]] && report+=$'\n'"$pkg"$'\n'"$out"$'\n'
    breaks+="$(incompatible_lines "$pkg" <<<"$out")"$'\n'
  done < <(printf '%s\n%s\n' "$base_pkgs" "$head_pkgs" | sort -u)

  if $incompatible; then
    broken+=("$m")
    status="incompatible changes"
    annotate "$BREAK_LEVEL" "Breaking API change in $m" "$(grep . <<<"$breaks")"
  elif [[ -n "$report" ]]; then
    status="compatible changes only"
  else
    status="no API changes"
  fi
  say "### $m: $status"
  if [[ -n "$report" ]]; then
    report="${report#$'\n'}"
    say "" '```' "${report%$'\n'}" '```'
  fi
  say ""
done

if [[ ${#broken[@]} -gt 0 ]]; then
  echo "Incompatible API changes in: ${broken[*]}" >&2
fi
if [[ ${#exposed[@]} -gt 0 ]]; then
  echo "Internal types in the public API of: ${exposed[*]}" >&2
  exit 3
fi
if [[ ${#broken[@]} -gt 0 ]]; then
  exit 1
fi
