# Decisions: how the XSD becomes protobuf

The second half of [`decisions.md`](decisions.md): what the generator does
with the shapes OTA's XSD takes, where the obvious translation was tried and
failed. Each is a rule in `sync/internal/plan`, not a list.

## Anonymous types nest where XSD scopes them

An anonymous complex type -- `<xs:element name="Rates"><xs:complexType>` --
is a message nested in the message whose element declares it, which is
where XSD scopes it too: `RatePlan.Rate`. Two exceptions, both because the
type has no single scope:

- the anonymous type of a **top-level element**, which every type that
  references the element shares. Nested in whichever reached it first, the
  others would import that file, and a cycle follows. Its key is the element
  name alone, and it is a top-level message.
- the anonymous child of a **named wrapper that is folded away** (below).
  The wrapper was its scope and is no longer a message. It is top-level.

A top-level message from an anonymous type has a name the generator chose,
so it yields to a named type's, whichever is reached first: the folded
child `Address` becomes `AddressItem` when OTA's `AddressType` arrives, and
a top-level element's type takes `Element` instead (`plan.Builder.Top`).
Fields point at messages and are named when written, so renaming one already
in use is safe.

Nesting shadows: inside `GuestCount`, a reference to `GuestCount.GuestCount`
binds to the nested message and resolves to nothing, since protobuf walks
outward to the first scope with a member of that name and never backtracks.
Where any enclosing scope has a member named like a reference's first
component, the emitter writes the reference fully qualified.

## A cycle of types shares one file

OTA's types are recursive -- a type can reach itself through others -- and
protobuf forbids files that import one another in a cycle. One message per
file cannot hold for those, and nothing else can declare them at all. So
the messages of each strongly connected component of the package's type
graph share the file of the member the package reached first (`plan.Group`,
Tarjan's algorithm). It is the only exception to one message per file
that types make; the others are requests and their responses.

Field types are therefore resolved when written, not when planned: a field
points at its message or enum, and reads its name and file at emit time,
after a union is named and a cycle regrouped.

## Each package splits into types, codes and enums

The user asked for a package's files in subfolders. A flat `v1/` held the
resource beside a hundred value messages and as many enums; reservation's had
four hundred files. `v1/` now keeps the resource, `messages.proto`,
`service.proto` and one file per custom method, and three subpackages hold
the rest: `types/` (OTA value messages, `Money`, `Date`), `codes/`
(code-table enums), `enums/` (closed XSD enumerations).

A subfolder must be its own package -- buf's `PACKAGE_DIRECTORY_MATCH` --
and AIP-215 forbids a field naming a message of another package. It does
not forbid this: api-linter's `foreign-type-reference` compares packages
truncated at their version segment, so `...reservation.v1.types` is
`...reservation.v1` to it (`rules/aip0215/foreign_type_reference.go`,
api-linter 2.3.1). The split changes where a declaration sits, not what any
package contains.

buf `BASIC` also refuses a cycle of package imports, which files alone can
avoid and packages cannot: `v1/` imports `types/`, so a `types/` message
that reaches back into `v1/` -- through an anonymous type first planned
there -- would close one. Such a message is lifted into `v1/`
(`plan.liftCycles`), and a file cycle with a `v1/` member in it stays in
`v1/`. References are written to resolve from where they stand: to a
sibling subpackage relative, `codes.RoomAmenityType`; into `v1/` fully
qualified (`emit.crossRef`).

## Wrapper elements fold into a repeated field

OTA wraps lists: `<Rates><Rate/><Rate/></Rates>`. Translated literally that
is a `Rates` message holding `repeated Rate rate`, and every reader walks
through it. An element with no attributes, no text and exactly one child --
repeated, and once `TPA_Extensions` is set aside -- is that child:
`repeated Rate rates`. The annotation's `path` keeps both names,
`HotelRatePlanType/Rates/Rate`, so a codec can put the wrapper back.

A wrapper inside a `oneof` is not folded, because a `oneof` member cannot
be repeated. A wrapper with any attribute is a message, because the
attribute would have nowhere to go.

## A resource is referenced, not embedded

OTA embeds a whole customer profile -- airline seat and rail preferences
included -- wherever it means "this person", even to say who may view a
hotel's content (`Viewerships/Viewership/Profiles/Profile`). Copied into
every hotel package, by rule 4, it pulled air and rail types into rate
plans. A type that is a resource elsewhere in the schema is a resource name
wherever OTA embeds it: `ProfileType` is a `string` holding
`profiles/{profile}`, with a `resource_reference` to
`common.opentravel.org/Profile` (`catalog.References`).

Not every embedded type is a reference. A rate plan inside a reservation's
room stay is the terms as booked, which must not change when the rate plan
later does; it stays a copy.

## Workbook defects are fixed from OpenTravel's own workbooks

The 2024-09-04 workbook has two defects a hotel field reaches. Each is
fixed from an older OpenTravel workbook in the pinned checkout, never by
guessing, and the source is cited beside the entry:

- `RLT` (Room Location Type) code 1 has no name. The 2016-02-24 workbook and
  the 2017-10-25 one shipped with 2017B both name it "Away from the
  elevator" (`catalog.CodeFixes`).
- `FAR` (Fare Restriction) lists "Zone fare" as 39, between 29 and 31, and
  "One way fare" is also 39; 30 is otherwise missing. Zone fare is 30
  (`catalog.CodeRenumbers`).

A citation of a table the workbook does not have is not guessed either.
Brand Category Code (`BCC`) is withdrawn, and OTA points to Segment Category
Code (`SEG`), whose codes are not BCC's; OTA's error codes (`ERR`) are a
separate sheet for the envelope. Both stay validated strings
(`catalog.CodeListAliases`, `NoTable`).
