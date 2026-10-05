# protobuf-ota dev tasks -- run `just` (or `just --list`) to see recipes.
#
# Common flows:
#   just sync     # regenerate protobuf/ and its READMEs from the pinned publications
#   just docs     # regenerate only the READMEs
#   just lint     # what CI checks: format, buf lint, build, api-linter, tests, caps
#   just ci       # everything CI runs
#
# Requires: go, buf, api-linter.

# List recipes (default when you run bare `just`).
_default:
    @just --list

# Regenerate protobuf/ from the pinned publications.
[doc("Regenerate every .proto from the pin. Nothing under protobuf/ is hand-written.")]
sync:
    go run ./sync/cmd/sync
    buf format -w
    @just docs

# Render the Markdown reference: one README per package, plus an index.
[doc("Regenerate the Markdown reference beside the protos.")]
docs:
    @go run ./sync/cmd/docs

# Report what the generator reads; writes nothing. Pass flags through:
# `just survey -tree OTA_HotelRatePlanNotifRQ/RatePlans/RatePlan`.
[doc("Survey the parsed suite: -tree PATH [-depth N] [-why TYPE].")]
survey *args:
    @go run ./sync/cmd/survey {{args}}

# Check sync/spec.yaml against the checkout under modules/.
[doc("Verify the pin matches modules/opentravel.")]
spec:
    ./scripts/check-spec.sh

# Check every URL in docs/references.md resolves. Needs network access.
[doc("Check the reference link index.")]
links:
    ./scripts/check-links.sh

# Format Go sources in place.
fmt:
    gofmt -w sync

# Build and test the generator.
[doc("Build and vet the generator, and run its tests.")]
test:
    go build ./sync/...
    go vet ./sync/...
    go test ./sync/...

# Everything the Lint job checks. Mutates nothing.
[doc("Format check, buf lint, buf build, api-linter, tests, and the line caps.")]
lint: spec aip test cap
    @test -z "$(gofmt -l sync)" || { echo "unformatted Go (run: just fmt):"; gofmt -l sync; exit 1; }
    buf format --diff --exit-code
    buf lint
    buf build

# Run the Google API linter over every proto. No config file, no disabled rule.
#
# api-linter exits 0 even when it reports problems -- it is a reporter -- so
# a non-empty findings table is made the failure it already was. coversa
# learned this when `just ci` passed while CI failed on the same findings.
[doc("Run api-linter over protobuf/ (zero findings, no suppressions).")]
aip:
    #!/usr/bin/env sh
    set -eu
    d=$(mktemp -d); trap 'rm -rf "$d"' EXIT
    buf build -o "$d/desc.binpb" --as-file-descriptor-set
    api-linter --descriptor-set-in "$d/desc.binpb" \
        --output-format=summary $(find protobuf -name '*.proto') | tee "$d/out"
    if grep -q '| core::' "$d/out"; then
        echo "api-linter reported violations above. Rule 1: fix the generator," >&2
        echo "never the emitted file, and never except the rule." >&2
        exit 1
    fi

# The line caps, on files a person writes: 200 for Go, 250 for Markdown.
# protobuf/ and README.md are exempt; docs/generator.md says why.
[doc("Check the line caps: 200 for Go, 250 for Markdown.")]
cap:
    #!/usr/bin/env sh
    over=0
    for f in $(find . -path ./gen -prune -o -path ./build -prune -o -path ./modules -prune \
        -o -path ./.git -prune -o -path ./protobuf -prune -o -name 'README.md' -prune \
        -o \( -name '*.md' -o -name '*.go' \) -print); do
        n=$(wc -l <"$f")
        case "$f" in
            *.go) cap=200 ;;
            *)    cap=250 ;;
        esac
        [ "$n" -gt "$cap" ] && { echo "$f: $n lines, over the $cap-line cap"; over=1; }
    done
    [ "$over" -eq 0 ] || { echo "Split the files above; do not compress them."; exit 1; }
    echo "All hand-written files within their line cap."

# Verify the committed tree matches what the generator produces.
[doc("Fail if protobuf/ is stale relative to the pin.")]
verify-sync: sync
    @git diff --exit-code -- protobuf/ codes.lock \
        || { echo "protobuf/ or codes.lock is stale -- run 'just sync' and commit"; exit 1; }

# Everything CI runs.
[doc("Everything CI checks.")]
ci: lint verify-sync links
