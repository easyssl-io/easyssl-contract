#!/bin/sh
# Releases this module as version VERSION (MAJOR.MINOR.PATCH, no `v`):
#
#   scripts/release.sh 0.1.0
#
# 1. Refuses unless the working tree is clean, VERSION is not tagged yet,
#    CHANGELOG.md has a `## VERSION` entry, and
#    VERSION is ContractVersion in plugincontract/plugincontract.go: the
#    constant is the version every build of this module reports, so the tag
#    follows it, never the reverse.
# 2. Refuses a breaking-version change that PreviousBreakingVersion does not
#    record: when VERSION's breaking version differs from the last release's,
#    PreviousBreakingVersion must name the last release's.
# 3. Copies the current examples of every package whose test pins them with
#    contracttest.Messages to `<package>/testdata/v<VERSION>/`. Those copies
#    never change afterwards: each later build must still decode them.
# 4. Runs the tests, commits, and tags `v<VERSION>`. Pushing is left to the
#    caller: `git push origin HEAD v<VERSION>`.
set -eu
cd "$(dirname "$0")/.."

[ $# -eq 1 ] || { echo "usage: $0 VERSION (e.g. 0.1.0)" >&2; exit 2; }
version=$1
echo "$version" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' ||
  { echo "$version is not MAJOR.MINOR.PATCH" >&2; exit 2; }

[ -z "$(git status --porcelain)" ] || { echo "the working tree has changes" >&2; exit 1; }
if git rev-parse -q --verify "refs/tags/v$version" >/dev/null; then
  echo "v$version is already tagged" >&2; exit 1
fi
grep -qx "## $version" CHANGELOG.md 2>/dev/null ||
  { echo "CHANGELOG.md has no \"## $version\" entry: write what this release changes first, in its own commit" >&2; exit 1; }

const() { sed -n "s/^const $1 = \"\(.*\)\"\$/\1/p" plugincontract/plugincontract.go; }
current=$(const ContractVersion)
previous_breaking=$(const PreviousBreakingVersion)
[ "$current" = "$version" ] ||
  { echo "ContractVersion is \"$current\", not $version: set it first, in its own commit with the examples it changes (go test -update in each package that pins them)" >&2; exit 1; }

# breaking VERSION: 0.<minor> while the major number is 0, the major after.
breaking() { echo "$1" | awk -F. '{ print ($1 == "0") ? "0." $2 : $1 }'; }
last=$(git tag --list 'v[0-9]*' | sed 's/^v//' | grep -Ev -- '-' | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1)
if [ -n "$last" ] && [ "$(breaking "$last")" != "$(breaking "$version")" ] &&
  [ "$previous_breaking" != "$(breaking "$last")" ]; then
  echo "v$version breaks v$last, so PreviousBreakingVersion must be \"$(breaking "$last")\", not \"$previous_breaking\"" >&2
  exit 1
fi

for test in $(grep -l 'contracttest.Messages(t, "testdata"' $(git ls-files '*_test.go')); do
  dir=$(dirname "$test")/testdata
  mkdir -p "$dir/v$version"
  cp "$dir"/*.json "$dir/v$version/"
  git add "$dir/v$version"
done

GOWORK=off go test ./...
git commit -q -m "Release v$version" -m "Keeps this version's examples under testdata/v$version/, which every later build must still decode."
git tag "v$version"
echo "Tagged v$version. Push with: git push origin HEAD v$version"
