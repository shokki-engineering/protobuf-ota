# Decisions

What was deliberately not done, and why. Each of these was the obvious choice
and was rejected for a reason worth recording. Entries inherited unchanged
from coversa-protobuf are not repeated; this file records where OTA forced a
different answer, or posed a question VSS never did. How OTA's XSD shapes
become protobuf -- nesting, cycles, wrappers, references, workbook defects --
is in [`decisions-generator.md`](decisions-generator.md).

## OTA 1.0, not OTA 2.0

OpenTravel publishes two specifications. OTA 2.0 is the newer, and it is
resource-oriented -- the obvious fit for AIP. It was rejected as the source
because its hospitality libraries cover **descriptive content only**:
`HospitalityContent_HospitalityResources_1_0_0.xsd` is 108 lines. Availability,
rates, inventory, reservations, blocks and RFPs exist only in the 1.0 XML
message suite, and those are what "hotel booking" means.

2017B is the newest 1.0 publication in OpenTravel's own repository. The 2.0
2020A object suite is read for reference -- where it models a hotel's content,
it confirms the resource shape chosen here -- but nothing is generated from it.

## Messages become resources, not messages

The faithful translation is one protobuf message per OTA message:
`HotelResRQ`, `HotelResRS`, and a service with one RPC per pair. It was
rejected. AIP-121 designs an API around the nouns it manages, and OTA's
nouns -- a hotel, a rate plan, a reservation -- are buried two and three
levels inside envelopes named for verbs.

The cost is real. One OTA message can become several calls:
`OTA_HotelDescriptiveContentNotifRQ` carries a hotel together with its rooms,
policies and media, which are separate resources here, and writing them is
not atomic. A codec that accepts the OTA message must fan it out, and must
report a partial failure as one.

## The envelope is dropped

Every OTA message carries `OTA_PayloadStdAttributes` -- `EchoToken`,
`TimeStamp`, `Target`, `Version`, `SequenceNmbr`, `TransactionIdentifier`,
`PrimaryLangID` -- a `POS` naming the requesting system, and in its response
`Success`, `Warnings` and `Errors`. None of it is about a hotel. Each job
belongs to the transport, and AIP already names where:

| OTA | Here |
| --- | --- |
| `Errors`, `Warnings` | `google.rpc.Status` and its details, returned by gRPC |
| `EchoToken`, `TransactionIdentifier` | request metadata |
| retry safety | AIP-155 `request_id` on the methods that create |
| `POS/Source/RequestorID` | the caller's authenticated identity |

The `POS` *inside* a `HotelReservation` is different and is kept: it records
which channel made the booking, which is a fact about the reservation.

## `TPA_Extensions` is dropped

It is `xs:any` -- an extension point for trading partners, with by definition
no schema. There is nothing to generate a type from. Carrying it as a
`google.protobuf.Struct` or as bytes would make every resource accept
arbitrary, unvalidated content at that point, and rule 14 exists to prevent
that. A partner that needs an extension wraps these types in a package of its
own.

## Code tables are enums, numbered by their code

OpenTravel publishes 152 code tables in a workbook released separately from
the schema -- the code list pinned here is 2024-09-04, the schema 2017B -- and
the 2017B XSDs cite 129 of them. Each becomes an enum: `RoomAmenityType`,
`HotelAmenityCode`, `MealPlanType`. The first design made them strings, on
coversa's `CountryCode` argument -- the list is someone else's and moves on
its own schedule. Surveying the workbook removed both halves of it:

- **Every code is an integer**, all 5,293 of them, and **none is 0**. So the
  enum value's number *is* the OTA code -- `ROOM_AMENITY_TYPE_HAIRDRYER =
  50` is RMA 50 -- and AIP-126's `_UNSPECIFIED = 0` collides with nothing. The
  wire carries exactly the number OTA's XML carries.
- **proto3 enums are open.** A code a later list adds, arriving before this
  schema is regenerated, is kept as its number by every mainstream runtime
  and written back unchanged. Unlike `CountryCode`, a missing value is not a
  value the API cannot express.

What remains of the cost is naming. Values are named from OTA's code name, so
OTA renaming a code would rename the enum value -- harmless on the wire,
breaking in proto3 JSON, which writes enum names. **A name, once emitted, is
frozen**: `codes.lock` records every table, code and identifier, the way
coversa's `buffers.lock` records field slots, and an OTA rename changes the
comment and the `source_spelling`, never the identifier. A code OTA removes
stays, marked `deprecated`; its number is OTA's and cannot be reused anyway.

The workbook is not clean, and rule 15 applies: the `FAR` table gives code
39 two names, and `RLT` leaves code 1 unnamed. Both are cited in scope --
`RLT` by `RoomGroup/@RoomLocationCode` in a rate plan, `FAR` through a
guest profile's air preferences --
and both stopped the build until fixed from OpenTravel's own older
workbooks (`decisions-generator.md`). Eight tables have two codes whose
names collide once made identifiers (`HAC` and `RMA` twice each); a
collision takes its code as a suffix, by rule.

Three kinds of code-valued field stay `string`:

- one whose XSD cites no table -- OTA did not say what the code means, and
  choosing for it would be a guess;
- one using OTA's extended form, `101.EQP` -- vehicle-only, and outside scope;
- the **identifying codes**, below.

A closed XSD enumeration -- `DayOfWeekType` -- was always an enum.

## Resource ids keep OTA's codes verbatim, and are not enums

`DBL` and `BAR` are not from a code table. A room type code and a rate plan
code are **chosen by each hotel** -- `RatePlanCode` is a `StringLength1to64`
the property fills in -- so no list of them exists for an enum to hold. What
OTA does standardise is the *kind*: Rate Plan Type (`RPT`), Bed Type (`BED`),
Guest Room Info (`GRI`). Those are enums, as fields on the resource. The code
is its id.

The id keeps OTA's spelling: `hotels/H1234/ratePlans/BAR`, not `bar`.
AIP-122 asks for RFC 1034 ids, which are lower case, but it is a SHOULD, and
OTA codes are case-sensitive: lower-casing would merge two rate plans a hotel
keeps apart, and a reply could not restore the spelling the hotel sent.

## An XSD union is a `oneof`

XSD lets a value be any of several types; protobuf says the same with a
`oneof`, one member per member type. That is faithful -- the source accepts
every member -- where picking one would refuse valid OTA. One refinement:
members that are **enumerations merge into a single enum**, because two enums
answering one question would make every reader check both.

`ResStatus` on a reservation is the case that settled it. Its type is a union
of `PMS_ResStatusType` (`Reserved`, `In-house`, `No-show`, ...),
`TransactionActionType`, and `UpperCaseAlphaLength1to2` -- the free one- or
two-letter status codes of a GDS (`HK`, `UC`). It becomes:

```proto
oneof lifecycle {
  State state = 20 [(google.api.field_behavior) = OUTPUT_ONLY];
  string gds_status_code = 21 [(google.api.field_behavior) = OUTPUT_ONLY];
}
```

`Reservation.State` -- nested and named `State`, as AIP-216 asks -- merges
the PMS statuses with the *states* among
`TransactionActionType`'s values -- `Pending`, `Committed`, `OnHold`. The
*verbs* among them -- `Book`, `Commit`, `Cancel` -- are methods (rule 8), and
`catalog.ActionValues` says which value is which. The GDS code is a string,
validated by the union member's pattern: GDS status codes are a list no OTA
table holds.

The `oneof` is named `lifecycle` rather than `status`, the word AIP-216
steers away from. api-linter checks each member rather than the `oneof`, and
`state-field-output-only` holds for `state` inside it.

## Availability, rates and inventory are daily resources

OTA sends ARI as ranges: `Start`, `End`, a weekday mask, a room type and a
rate plan, and the values that apply to all of it. The first design kept the
range as the resource. It was rejected because a range has no identity. A
second push over 1-15 March after a first over 10-20 March either splits a
resource, merges two, or creates a third overlapping one, and none of those
is something a client can address afterwards.

A day can be addressed. Each (room type, rate plan, day) is a resource whose
id composes those axes, as coversa composes VSS instance axes into one id. A
range push becomes one AIP-234 `BatchUpdate` across the days it covers, which
the codec expands from the range and the mask. A read of a range is an
AIP-160 filtered `List`. The number of resources is large and that is the
point: every one of them is a price or a count a channel manager actually
sells against.

## Cancelling is not deleting

`OTA_CancelRQ` becomes `CancelReservation`, an AIP-136 custom method that
moves `state` to `CANCELLED` and returns the reservation with its
cancellation number. It is not `DeleteReservation`. A cancelled reservation
is a record the hotel keeps, reports on and may be asked about; deleting it
would make cancellation indistinguishable from a reservation that never
existed. `Delete` exists too, per rule 7, and means purge.

## Reservations belong to a hotel

`hotels/{hotel}/reservations/{reservation}`, not a root collection. OTA
nearly allows otherwise: each `RoomStay` carries its own
`BasicPropertyInfo`. But `ResGlobalInfo/BasicPropertyInfo` occurs once, every
implementation surveyed -- AlpineBits, Booking.com's connectivity API -- is
per hotel, and a confirmation number is issued by one. A multi-property
itinerary is several reservations.

## Calendar values keep their calendar

- `xs:date` becomes a generated `Date`. A check-in date is a date *at the
  hotel*; a `Timestamp` would invent an instant and a time zone OTA never
  sent, and the stay would shift a day for any reader west of it.
- `xs:dateTime` becomes `google.protobuf.Timestamp`.
- OTA's date unions -- `DateOrDateTimeType`, `TimeOrDateTimeType`,
  `DateOrTimeOrDateTimeType` -- become a generated message with a `oneof`,
  as every XSD union does (above).
- `xs:duration` and OTA's `DurationType` stay ISO 8601 strings, validated by
  pattern. They are calendar periods: `P1M` has no fixed length, and OTA adds
  `N` for nights, so `P3N` means three nights whatever the clock says.
  `google.protobuf.Duration` can hold neither.

## Integers are `int32`

XSD's `integer`, `nonNegativeInteger` and `positiveInteger` are unbounded.
OTA uses them for counts, nights and occupancies. `int64` is the lossless
choice and was rejected: AIP-141 asks for `int32` unless the range demands
more, and proto3's JSON mapping writes every `int64` as a string, which would
make `"quantity": "2"` the wire form of a room count. The sign bound XSD
implies is restored with `buf.validate`. A value too large for `int32` is an
error in the codec, never a silent wrap.

## Optional scalars use `optional`

proto3 cannot tell an unset `int32` from zero, and OTA can. An absent
`MinLOS` is no minimum stay; `MinLOS="0"` is a value. Every optional numeric
or boolean attribute is a proto3 `optional` field, as AIP-149 recommends.

## `google.type` is still not taken

AIP-215 exempts `google.*`, and `google.type.Money`, `Date` and `Decimal` are
the correct protobuf answers for three of OTA's value types. coversa adopted
`Date` and `PostalAddress` and withdrew both, because its FlatBuffers and
Cap'n Proto layer silently dropped them -- a field vanished from the emitted
table with no diagnostic.

This repository has no such layer yet, so the reason does not bite today. The
line is held anyway: adopting `google.type` now and withdrawing it when that
layer arrives would be a breaking change to every package that uses money,
which is nearly all of them. The value types are generated into each package
instead, with `google.type`'s shapes, so a consumer converting between the
two has nothing to map.

## The annotation vocabulary sits outside both specs

`protobuf/opentravel/annotations/v1`, not under `hotel/`. `common` uses it
too. Its extension numbers are **54100-54103**, clear of coversa's
54000-54003 and buffers' 53000-53007, so a consumer can load this schema and
coversa's into one descriptor pool. Extension numbers on
`google.protobuf.FieldOptions` share a single space across everything loaded.
