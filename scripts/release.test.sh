#!/bin/sh
# Tests scripts/release.sh in a throwaway clone of this repository, so nothing
# here is ever tagged: a release follows ContractVersion, keeps a decodable
# copy of every pinned example, and a breaking change must be recorded in
# PreviousBreakingVersion.
set -eu
cd "$(dirname "$0")/.."
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
git clone -q . "$work/c"
mkdir -p "$work/c/scripts" && cp scripts/release.sh "$work/c/scripts/release.sh"
cd "$work/c"
git add scripts/release.sh
git -c user.name=t -c user.email=t@example.com commit -q -m "the script under test"
export GIT_AUTHOR_NAME=t GIT_AUTHOR_EMAIL=t@example.com GIT_COMMITTER_NAME=t GIT_COMMITTER_EMAIL=t@example.com

fail=0
expect() { # expect STATUS DESCRIPTION COMMAND...
  want=$1; what=$2; shift 2
  if "$@" >"$work/out" 2>&1; then got=0; else got=1; fi
  if [ "$got" != "$want" ]; then echo "FAIL: $what (exit $got)"; cat "$work/out"; fail=1; fi
}
set_const() { # set_const NAME VALUE, committed with the examples it changes
  sed -i.bak "s/^const $1 = \".*\"\$/const $1 = \"$2\"/" plugincontract/plugincontract.go
  rm plugincontract/plugincontract.go.bak
  GOWORK=off go test ./wire/ ./plugincontract/ ./catalog/ -update >/dev/null
  git commit -q -am "set $1"
}

expect 1 "a version other than ContractVersion" scripts/release.sh 0.1.0
set_const ContractVersion 0.1.0
expect 0 "the first release" scripts/release.sh 0.1.0
[ "$(git tag --points-at HEAD)" = v0.1.0 ] || { echo "FAIL: v0.1.0 is not on the release commit"; fail=1; }
for p in wire plugincontract catalog; do
  [ -n "$(ls "$p/testdata/v0.1.0/"*.json 2>/dev/null)" ] || { echo "FAIL: no copy of $p's examples"; fail=1; }
done
expect 1 "a version already tagged" scripts/release.sh 0.1.0

set_const ContractVersion 0.2.0
expect 1 "a breaking change PreviousBreakingVersion does not record" scripts/release.sh 0.2.0
set_const PreviousBreakingVersion 0.1
expect 0 "a breaking change recorded" scripts/release.sh 0.2.0

set_const ContractVersion 0.2.1
echo x >> README.md
expect 1 "a working tree with changes" scripts/release.sh 0.2.1
exit "$fail"
