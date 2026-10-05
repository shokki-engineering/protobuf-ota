#!/usr/bin/env sh
# Check sync/spec.yaml against the OpenTravel checkout under modules/.
#
# The pin is what every generated file's banner claims. If the checkout has
# moved and the pin has not, every banner names a revision the files were not
# generated from -- a claim no linter and no compiler can see is false.
#
# One upstream, OpenTravel-Specifications, checked four ways: the commit is
# the one checked out, the version is that commit's date, the tree is clean
# so the commit describes it fully, and both publications the generator reads
# -- the message suite and the code list -- exist in it. See docs/spec.md.
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

checkout=modules/opentravel

# field prints one top-level key of the pin. Read with awk rather than a YAML
# parser so this script depends on nothing: it is the first thing CI runs.
field() {
	awk -v key="$1:" '$1 == key { sub(/^[^:]*:[[:space:]]*/, ""); print; exit }' sync/spec.yaml
}

pin_commit=$(field commit)
pin_version=$(field version)
pin_suite=$(field suite)
pin_codelist=$(field codelist)

fail=0
for k in commit version suite codelist; do
	if [ -z "$(field "$k")" ]; then
		echo "error: sync/spec.yaml has no '$k'." >&2
		fail=1
	fi
done
[ "$fail" -eq 0 ] || exit 1

# A submodule's .git is a file holding a gitlink, not a directory.
if [ ! -e "$checkout/.git" ]; then
	echo "error: $checkout is not checked out." >&2
	echo "       Run: git submodule update --init" >&2
	exit 1
fi

have_commit=$(git -C "$checkout" rev-parse HEAD)
have_version=$(git -C "$checkout" log -1 --format=%cd --date=short)

if [ "$pin_commit" != "$have_commit" ]; then
	echo "error: sync/spec.yaml pins commit $pin_commit" >&2
	echo "       but $checkout is at       $have_commit" >&2
	fail=1
fi

if [ "$pin_version" != "$have_version" ]; then
	echo "error: sync/spec.yaml says version $pin_version," >&2
	echo "       but that commit is dated $have_version" >&2
	fail=1
fi

if [ -n "$(git -C "$checkout" status --porcelain)" ]; then
	echo "error: $checkout has uncommitted changes, so the pinned commit" >&2
	echo "       does not describe what was read." >&2
	fail=1
fi

if [ ! -d "$checkout/$pin_suite" ]; then
	echo "error: suite $pin_suite does not exist in $checkout." >&2
	fail=1
fi

if [ ! -f "$checkout/$pin_codelist" ]; then
	echo "error: codelist $pin_codelist does not exist in $checkout." >&2
	fail=1
fi

if [ "$fail" -ne 0 ]; then
	echo >&2
	echo "Update the pin, run 'just sync', and record the change in docs/spec.md." >&2
	exit 1
fi

echo "spec pin:"
echo "  opentravel $pin_version ($(echo "$pin_commit" | cut -c1-7)) <- $checkout"
echo "  suite    $(field suite_label)  $pin_suite"
echo "  codelist $(field codelist_label)  $pin_codelist"
