# The generator

`sync/` is one Go module and one program, in two commands: `sync` writes the
protobuf, `docs` writes the Markdown reference beside it. Both read the same
publications and build the same model. A third, `survey`, reports what was
parsed and every AIP trap, and writes nothing.

It has **no dependencies**, and the `require` block should stay empty. XSD is
XML and the code list workbook is a zip of XML, both readable with
`encoding/xml` and `archive/zip`; the pin is flat YAML that a twenty-line
reader handles. A generator with nothing to keep current is one that still
builds when someone returns to it in two years. Reaching for an XSD library
is the tempting exception, and it is refused: the 2017B suite uses a known
subset of XSD (rule 15), and a library that accepts all of XSD accepts
exactly the unfamiliar shapes that rule exists to stop.

## Laid out by stage

```text
sync/cmd/sync/          the CLI that regenerates protobuf/
sync/cmd/docs/          the CLI that regenerates the Markdown reference
sync/cmd/survey/        the CLI that reports parse results and AIP traps
sync/internal/spec/     the revision pin
sync/internal/xsd/      the XSD reader: types, groups, inheritance, facets
sync/internal/codelist/ the code list workbook reader
sync/internal/naming/   identifier case and pluralisation
sync/internal/catalog/  the AIP rename tables -- data, not logic
sync/internal/model/    what becomes a resource and a package, and its name
sync/internal/plan/     what one attribute or element becomes
sync/internal/emit/     the protobuf text
sync/internal/docs/     the Markdown reference
```

Each package is one stage, and the dependencies point one way: `emit` and
`docs` read `model`, `model` drives `plan`, `plan` reads `catalog`,
`naming`, `codelist` and `xsd`, and nothing knows about `emit` or `docs`.
`xsd` and `codelist` know about nothing but the standard library; they turn
files into a typed tree and stop there. A change to how a field is *named*
has one place to go; a change to how it is *written* has another.

`model.Package.RPCs()` is the one list of a package's methods -- the standard
set for its shape, then its custom and batch methods -- and both `emit`
(`service.proto`) and `docs` (the README's methods table) render it, so the
reference cannot describe a method the schema does not have.

`xsd` resolves what makes OTA's schema hard to read by eye -- `extension` and
`restriction` bases, `group` and `attributeGroup` references, anonymous
types, and unions -- so that every stage after it sees a type as the flat set
of elements and attributes it actually has. It keeps where each one came
from, because a rename keyed on `DateTimeStampGroup/@CreateDateTime` must
apply everywhere that group is pulled in.

`codelist` reads the workbook's index and code sheets into tables keyed by
their three-letter abbreviation, and extracts each XSD citation --
`Refer to OpenTravel Code List Room Amenity Type (RMA)` -- by that
abbreviation. The citations are prose and their spelling varies; the
abbreviation is what is matched, and a citation without one goes through
`catalog.CodeListAliases` or fails the build.

Each table a field in scope cites becomes an enum, and `sync` reads and
rewrites **`codes.lock`** (`sync/internal/ledger`), the committed ledger of
every enum value it has emitted. A name in the ledger never changes, only is
added to: that keeps an OTA rename of a code from renaming an enum value and
breaking every JSON consumer. The sync workflow fails on any change to it
that was not committed with the regenerated tree.

`catalog` is the odd one, and deliberately so: it is data, not logic. Every
entry is a place OTA vocabulary and an AIP rule disagree, and every one is
also a row in the catalogue at the foot of [`conventions.md`](conventions.md).
It is what a specification bump most often needs edited, so it is kept where
it can be read without reading the code around it. The `resources`,
`methods` and `customs` tables in `model` are the same kind of thing and
[`resources.md`](resources.md) is their rendering.

## The line caps

**Every Go file is capped at 200 lines**, tighter than the 250 prose gets, and
CI checks both. The cap is a prompt: the generator is decomposed by stage, so
a file over it usually means two stages have merged, and the fix is to find
the seam rather than to compress. The same holds for prose -- move a passage
to the document that owns it, do not squeeze it.

Exempt, each because the remedy the cap prescribes is unavailable rather than
merely inconvenient:

| Exempt | Why |
| --- | --- |
| everything under `protobuf/` | generated. Protobuf cannot continue a message across files, and the READMEs beside them are written by `docs`. A reservation's `RoomStay` alone has dozens of fields, each with a comment and an annotation |
| `README.md` | its length is diagrams. Moving half of it behind a link produces a worse entry document, not a shorter one |

The exemption replaces protobuf-rfc's 250-line cap on a `.proto`, which
cannot apply to a generated file. What that cap was asking for is enforced
directly instead: **one top-level message or enum per file**. The
exceptions are where a request and its response are meaningless apart --
`messages.proto`, and one file per custom method -- and a cycle of types,
which protobuf forces into one file (`decisions-generator.md`).

## Where the files go

`plan` assigns each top-level declaration a subpackage of `v1`: the resource,
its requests and its custom methods stay in `v1/`; OTA value messages and
the generated value types go to `types/`, code-table enums to `codes/`,
closed XSD enumerations to `enums/`. `plan.Group` then puts each cycle of
types in one file and `plan.liftCycles` lifts into `v1/` any `types/`
message that reaches back into it. `emit.crossRef` writes each reference so
it resolves from wherever it is written; why the split is lint-legal is in
`decisions-generator.md`.

Every other hand-written file is capped, Markdown and Go alike. The list is in
the `cap` recipe in the justfile, which is what CI runs.

## Why there is a workspace

`go.work` names one member, `./sync`. Any other `go.mod` in the tree -- a
template for compiling generated Go in CI, or the codec below -- is either a
member with packages of its own or not a member at all; a workspace member
with nothing to build would be a lie. CI and the justfile set `GOWORK=off` for
any build inside generated output.

## The codec is not the generator

Several rules speak of a codec: it expands an ARI range into days, divides an
amount by `DecimalPlaces`, fans a descriptive-content message out into
resources, and writes OTA XML back. It is a separate module, `codec/`, which
reads the annotations (rule 5) and the generated types; it is not part of
`sync` and does not exist yet. Nothing in the schema may assume it -- a field
that only makes sense once a codec has run is a field designed wrong.

## Tests

`go test ./sync/...` covers the decisions, not the whitespace: which XSD
constructs the reader accepts and refuses, how a type's inheritance and
groups flatten, how identifiers are cased and pluralised, which types become
resources and what their names are, which code table a citation resolves to,
and -- in `emit` -- the handful of facts the emitted text has to carry. A
code-valued field reaching both the annotation and the comment, an XSD length
facet and a pattern being intersected into one rule rather than two, the held
field numbers being reserved rather than skipped.

The strongest check is not in the test suite: CI regenerates the whole tree
and diffs it. A change that alters any generated file shows up there whether
or not a test covers it.
