#!/usr/bin/env sh
# Check that every URL in the documentation resolves.
#
# docs/references.md exists because CLAUDE.md points at it instead of working
# from memory -- AIP rule semantics have been guessed wrong before. A link
# index nobody checks decays into the thing it was meant to replace. CLAUDE.md
# and the rest of docs/ cite the same sources, so they are checked too.
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

fail=0
checked=0

# Bare URLs, one per line, deduplicated. The documents write them in prose
# (`AIP-131 <https://aip.dev/131>`), in tables and as Markdown links, so a grep
# for the scheme finds them all without a Markdown parser.
urls=$(grep -ohE 'https://[^ )>|`"]+' CLAUDE.md docs/*.md | sed 's/[.,;:]$//' | sort -u)

# status prints the HTTP status of one URL: HEAD first because most targets
# are large HTML pages, then a ranged GET because some hosts refuse HEAD.
status() {
	code=$(curl -sS -o /dev/null -w '%{http_code}' -I -L --max-time 25 "$1" 2>/dev/null || echo 000)
	if [ "$code" -lt 200 ] || [ "$code" -ge 400 ]; then
		code=$(curl -sS -o /dev/null -w '%{http_code}' -r 0-0 -L --max-time 25 "$1" 2>/dev/null || echo 000)
	fi
	echo "$code"
}

for url in $urls; do
	checked=$((checked + 1))
	# Three attempts, a pause between: a transient failure is not a dead link.
	for attempt in 1 2 3; do
		code=$(status "$url")
		if [ "$code" -ge 200 ] && [ "$code" -lt 400 ]; then
			break
		fi
		[ "$attempt" -lt 3 ] && sleep 2
	done
	if [ "$code" -lt 200 ] || [ "$code" -ge 400 ]; then
		echo "  $code  $url" >&2
		fail=$((fail + 1))
	fi
done

if [ "$fail" -ne 0 ]; then
	echo >&2
	echo "error: $fail of $checked links do not resolve." >&2
	echo "       Fix the link or drop the entry; do not leave it dangling." >&2
	exit 1
fi

echo "CLAUDE.md, docs/: $checked links, all resolve"
