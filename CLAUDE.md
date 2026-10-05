# protobuf-ota — working conditions

Protobuf types for the OpenTravel Alliance hotel specification -- the OTA 1.0
XML message suite and the OpenTravel code lists -- reshaped from XML
request/response pairs into resource-oriented APIs under Google's AIPs. Every
rule below is a hard constraint, not a preference. `.github/workflows/`
enforces the ones that can be enforced mechanically, one workflow per concern,
so a failure names what broke. The rest are on you.

> **Where this stands.** `sync` and `docs` generate `protobuf/` -- 28
> resources, their standard, custom and batch methods, and a README each --
> and api-linter reports 0 findings over 1504 files (2026-10-05). CI is in
> `.github/workflows/`; `codes.lock` and `model.CheckCatalogue` are enforced.
> Open questions are in `docs/resources.md`. Never fill a gap by hand.

The conventions are
[coversa-protobuf](https://github.com/oh-tarnished/coversa-protobuf)'s, which
adapt [protobuf-rfc](https://github.com/the-protobuf-project/protobuf-rfc),
and the numbering is kept. Where a rule differs from coversa it says why --
almost always because OTA is a message specification and VSS a data model.

## 1. No lint rule may be disabled

Not with an `except` list in `buf.yaml`, not with a `disabled_rules` entry in
an api-linter config, not with an inline `(-- api-linter: ... =disabled --)`
comment. There is no api-linter config file in this repo and there must not
be one.

If a linter objects, the code is wrong. Fix the code -- which here means
fixing **the generator**, never the emitted file.

When two linters contradict each other, narrow one tool's scope rather than
except a rule. So `buf.yaml` uses `BASIC`: buf STANDARD wants `HotelsService`
and `GetHotelResponse`, AIP-131 wants `Hotels` returning `Hotel`. buf checks
structure, api-linter API design, every rule of each enforced.

## 2. Nothing under `protobuf/` is written by hand

Both the `.proto` files and the `README.md` beside them are generated -- the
schema by `sync`, the reference by `docs`, from the same model rather than by
parsing the emitted protos back in.

Every `.proto` is emitted by `sync` from the publications pinned in
`sync/spec.yaml`: message structure from the OTA 2017B XSDs, code tables from
the 2024-09-04 code list workbook. Every file carries a `DO NOT EDIT` banner
naming both and the repository revision they were read at. Editing one is not
a small shortcut, it is a change the next `just sync` silently reverts.

A fix belongs in the generator, and usually in a named table there:

| Change | Where |
| --- | --- |
| a field name a linter rejects | `FieldRenames`, `sync/internal/catalog/renames.go` |
| a field name a linter flags but that is right as it stands | `AcceptedTraps`, same file |
| a message name a linter rejects | `TypeRenames`, `catalog.go` |
| an attribute the responding system assigns | `ServerAssigned`, same file |
| an attribute that is an operation, by name | `Actions`, same file |
| whether an OTA enum value is an action or a state | `ActionValues`, same file |
| a union placed in the message as one `oneof` | `InlineUnions`, same file |
| a type that is a resource, wherever OTA embeds it | `References`, same file |
| a code-table reference the XSD prose spells wrongly | `CodeListAliases`, same file |
| a defect in the code list workbook | `CodeFixes`, `CodeRenumbers`, same file |
| a resource: its root, id, parent, shape, domain, carve-outs | `resources*.go` in `sync/internal/model` |
| an OTA message and the RPCs that do its job | `methods`, `sync/internal/model/methods.go` |
| a custom or batch method: its kind, verb, request, response | `customs`, `customs.go`, same package |
| a fact two merged OTA types both carry | `Merge.Drop` on the resource |
| a plural English gets wrong | `irregularPlurals` in `sync/internal/naming` |
| which publications to read | `sync/spec.yaml`, then `just sync` |

Every entry is keyed by **OTA path**: a named XSD type or group and the path
inside it, attributes marked `@` -- `RatePlanGroup/@RatePlanCode`. A named
type is OTA's unit of reuse, so the entry applies wherever it is used. An
entry naming a path the suite no longer has fails the build:
`model.CheckCatalogue`.

How the generator itself is laid out -- one package per stage, the line caps,
**one message per file**, and why there is a `go.work` -- is in
`docs/generator.md`.

@docs/generator.md

## 3. AIP is the convention

api-linter must report zero violations across every file. In practice:

- `option java_package`, `java_outer_classname`, `java_multiple_files` on
  every file -- api-linter mandates all three (AIP-191)
- **no `go_package`**: the Go import path depends on where a consumer
  generates, so consumers set it through managed mode. Nothing forces it, so
  nothing is committed
- a comment on every message, field, enum, enum value, service and RPC: OTA's
  `xs:documentation`, or a sentence made from the declaration's name
- `(google.api.field_behavior)` on every field
- `(google.api.resource)` on every resource, `(google.api.resource_reference)`
  on every field naming one
- `(google.api.http)` on every RPC; a `(google.api.method_signature)`
  naming every required request field (AIP-4232)
- `int32`, never `uint32` (AIP-141). XSD's `nonNegativeInteger` and
  `positiveInteger` become `int32`, and the lost bound is restored as a
  `buf.validate` rule
- messages before top-level enums; services before messages
- a `uid` field carries `(google.api.field_info).format = UUID4` (AIP-148)

## 4. No cross-package message references

AIP-215 forbids a field from referencing a message in another proto package,
and exempts only `google.*`. A package must therefore contain everything it
references. Two consequences, and neither is a workaround:

- A **child resource** is not a field. `Hotel` has no `room_types` field;
  each is at `hotels/{hotel}/roomTypes/{room_type}`, fetched from `RoomTypes`.
- A **genuine association** is a resource name: where OTA embeds a guest
  profile, the field is a `string` holding `profiles/{profile}`, with a
  `(google.api.resource_reference)` (`catalog.References`).

`google.*` is the one exemption, and this repository **does not take it**, as
coversa does not; why, with no FlatBuffers layer here yet, is in
`docs/decisions.md`. `Money`, `Date` and `Decimal` are **generated into each
package's `types/`**. `Timestamp`, `FieldMask` and `Empty` are protobuf's.

## 5. Custom annotations are allowed, and there is exactly one

protobuf-rfc bans them outright. This repository has one vocabulary,
`protobuf.opentravel.annotations.v1`, carrying two facts protobuf has no
place for:

- **which OpenTravel code table an enum encodes** (`RMA`), and each value's
  code name as OTA spells it; the value's *number* is already its OTA code.
- **where the value lives in OTA's XML** -- element or attribute, and its path
  -- without which a codec cannot write the OTA message back.

coversa's licence was the *unit*; OTA sends units as data (rule 13). Otherwise
the same: **provenance still goes in comments**, and no second vocabulary.

## 6. Layout

```text
protobuf/opentravel/<spec>/<domain>/<resource>/v1/   ...<resource>.v1
    <resource>.proto  messages.proto  service.proto  <custom_method>.proto
    types/   ...<resource>.v1.types   OTA value messages, Money, Date
    codes/   ...<resource>.v1.codes   OpenTravel code-table enums
    enums/   ...<resource>.v1.enums   closed XSD enumerations
```

Each subfolder is its own package, so `PACKAGE_DIRECTORY_MATCH` holds, and
api-linter reads a package by its versioned prefix, so `...v1.types` is
`...v1` to AIP-215 (`rules/aip0215/foreign_type_reference.go`, 2.3.1). buf
refuses package import cycles: a `types/` message that reaches `v1/` is
lifted into it (`plan.liftCycles`; `docs/decisions-generator.md`).

`<spec>` is the OTA vertical: `hotel` for the hotel message suite, `common`
for what OTA defines across verticals -- guest profiles. Air, vehicle, rail
and cruise would each be a further `<spec>`; none is in scope. `<domain>` is
the functional area each resource row names: `property`, `revenue`,
`booking`, `sales`, `finance`, `reporting`. The `hotel` root and the `common`
packages have no domain, since a root belongs to no area and the common set is
small. The map is in `docs/resources.md`.

The annotation vocabulary is at `protobuf/opentravel/annotations/v1`,
**named** and outside either `<spec>`: a bare `opentravel/v1` would prefix
every other package. The buf module is rooted at the **repository root**, so
`protobuf/` is the first package segment; `build/` and `modules/` are
excluded.

## 7. Every resource has full CRUD, and the shape follows the source model

A message with `google.api.resource` gets the standard methods in its
package's `service.proto`, and the XSD decides which:

- An element its parent holds **many** of (`maxOccurs` > 1) -- room types,
  rate plans, reservations -- is a collection: `Get`, `List`, `Create`,
  `Update`, `Delete`, plus AIP-164 `Undelete`.
- One it holds **one** of -- area info, affiliation info -- is a singleton,
  AIP-156: `Get` and `Update` only.

**An OTA type with an identifying code is a resource.** `HotelCode`,
`RatePlanCode`, a room type's `InvTypeCode`, `InvBlockCode`, a reservation's
`ResID_Value` are OTA stating that the thing has identity. The code is the
resource id, the last segment of its name, never also a field.

**A fact keyed by a date is a daily resource.** OTA sends availability, rates
and inventory as date ranges, and a range has no identity. Each (room type,
rate plan, day) is a resource whose id composes its axes, and a range push is
an AIP-234 `BatchUpdate`. See `docs/decisions.md`.

Everything OTA does beyond CRUD is an AIP-136 custom or AIP-231/234 batch
method (`customs`). **Every message in scope is in the `methods` table
exactly once** -- every `OTA_Hotel*RQ`, and `OTA_Cancel`, `OTA_Read`,
`OTA_Delete`, `OTA_Profile*` -- mapped to RPCs or marked not taken with the
reason; `model.checkMethods` fails the build otherwise. `Package.RPCs()` is
the one method list `emit` and `docs` both read.

## 8. An OTA action is a method; an OTA status is output

coversa reads writability off each VSS signal's kind. OTA labels nothing that
way. Instead it puts the *operation* in the payload: `TransactionActionType`
(`Book`, `Quote`, `Hold`, `Modify`, `Commit`, `Cancel`, ...),
`RatePlanNotifType` (`New`, `Delta`, `Overlay`, `Remove`), and the `Removal`
and `Overwrite` flags. None of these becomes a field. Each is an RPC.
`TransactionActionType` mixes verbs (`Book`) with states (`Committed`), so the
`ActionValues` table in `sync/internal/catalog` says which each value is.

What an action produces -- `ResStatus` and its kin -- is a `state` field,
`OUTPUT_ONLY` per AIP-216, moved only by those methods, and sharing a `oneof`
with a GDS code where the XSD allows one (`docs/decisions.md`). Also
`OUTPUT_ONLY`: OTA's `DateTimeStampGroup`, absorbed into `create_time` and
`update_time`, and every path in `ServerAssigned`, such as a confirmation
number the hotel issues. An `update_mask` naming any of them is rejected
rather than silently ignored.

## 9. The specification revision is pinned and dated

All three publications -- the 2017B message suite, the 2024-09-04 code list,
and the OTA 2.0 object suite read for reference -- live in one upstream,
[OpenTravel-Specifications](https://github.com/OpenTravel/OpenTravel-Specifications),
the submodule `modules/opentravel`. `sync/spec.yaml` names its commit, that
commit's date (`YYYY-MM-DD`, as MCP versions its specification) and each
publication's path; all go into every banner.

Bumping a pin is a deliberate act, and `just spec` checks the pin against its
working tree; both are in `docs/spec.md`.

@docs/spec.md

## Rules 10-15

Identifiers, naming traps and the catalogue, money and units, and the two
annotation systems are in `docs/conventions.md`, with the same weight.

@docs/conventions.md

## Where to look things up

`docs/references.md` is a checked link index. **Follow the link, or the
api-linter source, rather than memory** -- coversa guessed wrong before.

@docs/references.md

## Before you finish

Run what CI runs:

```sh
just ci
```

which runs the pin check, `buf format`, `buf lint`, `buf build`, api-linter,
`go test`, the line cap, a regenerate-and-diff and the link check. The
workflows `lint`, `sync`, `breaking` and `links` are the authority.
