# Conventions

The second half of the working rules. `CLAUDE.md` holds rules 1-9 (structure,
layout and the spec pin) and imports this file; these are the ones about what
goes *inside* a file. Same standing as the others -- none is advisory.

## 10. Two identifiers, not one -- and a code that is neither

`uid` is ours: OUTPUT_ONLY, server-assigned, `(google.api.field_info).format
= UUID4` because AIP-148 requires it.

OTA's own `ID` attribute -- `GuestRoom/@ID` beside the room's `@Code`,
`HotelDescriptiveContent/@ID` beside its `@HotelCode` -- gets a separate
field, `ota_uid`, marked OPTIONAL and IMMUTABLE, unless the `resources` table
names it as the resource's identity, as it does for `MeetingRoom` and
`Restaurant`, which have no code. It arrives in another system's data, is
whatever that system chose, and must survive a round trip unchanged or every
re-import duplicates the record. Do not conflate the two.

The **identifying code** is neither. `HotelCode`, `RatePlanCode`, a guest
room's `Code` are the last segment of the resource name (rule 7) and appear
nowhere else in the message. A `UniqueID` element carrying a reservation's
confirmation numbers is data, kept as a repeated field, because OTA allows
several per reservation from several systems.

The generator enforces the boundary. A source field whose name lands on one
of a resource's own AIP fields -- `name`, `uid`, `etag`, the four timestamps
-- is a build error naming both, rather than a silent overwrite: OTA's
`Event/@Name` was caught this way, and is the event's `display_name`.

## 11. AIP naming traps

OTA's XML names are PascalCase and are snake-cased by rule. Many are still
unusable as field names. Rename the field, keep the source term in the
comment, and add a row to the catalogue at the foot of this file.

The recurring ones, checked against the api-linter 2.3.1 source rather than
from memory:

- AIP-122 flags **every** `_name` suffix except `display_name` (bare or
  prefixed), `given_name`, `family_name`, `full_resource_name` and
  `crypto_key_name`. OTA has dozens:
  `HotelName`, `RoomTypeName`, `RestaurantName`, `InvBlockName`, `ChainName`.
  The resource's own becomes `display_name`, others `<prefix>_display_name`.
- AIP-140 bans the prepositions `after at before between but by except for
  from in including into of over since to toward under upon with within
  without`, and the reserved words of common languages -- `default`, `class`,
  `package`, `new`, `delete`, `final`, `is`, `in` among them.
- AIP-142 reads a final `date`, `time`, `datetime`, `timestamp`, `seconds`,
  `nanos` and similar token as a claim that the field is a `Timestamp` or a
  `google.type` date. Rule 4 rules out `google.type`, so a field of the
  generated `Date` ends in `_day`, and one of `TimeOfDay` in `_clock`:
  `time_of_day` would carry the preposition `of`. A `Timestamp` ends in
  `_time`, in the imperative -- `update_time`, never `updated_time`. A
  calendar union drops its temporal word: OTA's `Start` and `End` are `start`
  and `end`. The generated value types obey the same rule:
  `Money.nano_units`, and `TimeOfDay`'s `hour`, `minute`, `second`,
  `nanosecond`, where `google.type` has `nanos` and `seconds`.
- AIP-143 mandates `region_code`, `currency_code`, `language_code` and
  `time_zone` for the values OTA calls `CountryCode`, `CurrencyCode`,
  `LanguageCode` and `TimeZone`.
- AIP-148 reserves a bare `id` (rule 10).
- AIP-216 reserves `state` for server-owned lifecycle and prefers an enum
  named `*State` over `*Status`. OTA uses `Status` for a dozen unrelated
  things -- only the lifecycle one becomes `state` (rule 8). By rule, any
  field typed by a `*State` enum is `OUTPUT_ONLY`; a code-table enum whose
  name ends in `Status` takes `Code` (`EmployeeStatusCode`); and the
  reservation's state enum is nested and named `Reservation.State`.

These are applied **by rule rather than by list**, so a later OTA
publication cannot introduce one unnoticed (`sync/internal/plan/names.go`):

| Rule | Example |
| --- | --- |
| AIP-140 reserved words get a suffix naming their role | `Default` -> `default_value` |
| AIP-140's abbreviation table, and `identifier` -> `id` | `Information` -> `info`, `SupplierIdentifier` -> `supplier_id` |
| AIP-140 `url` -> `uri` | `URL` -> `uri` |
| OTA's boolean `Ind` suffix and `Is` prefix are dropped (AIP-140) | `FeaturedInd` -> `featured`, `IsRoom` -> `room` |
| the seven weekday booleans (`DOW_PatternGroup`) collapse into one repeated enum | `Mon`..`Sun` -> `repeated Weekday weekdays` |
| a repeated field's last word is pluralised, unless a word is already plural | `Rate` -> `rates`; `GuaranteesAccepted` stays |
| the phrases OTA repeats, rewritten (`sync/internal/plan/phrases.go`) | `unit_of_measure` -> `unit`; `number_of_nights` -> `night_count`; `day_of_week` -> `weekday`; `amount_before_tax` -> `pre_tax_amount`; `length_of_stay` -> `stay_length`; `check_in` -> `checkin`; `source_of_business` -> `business_source` |
| an empty XSD enumeration literal | `""` -> `..._EMPTY` |
| a code-table enum value is its table's prefix plus the code name | RMA 50 `Hairdryer` -> `ROOM_AMENITY_TYPE_HAIRDRYER = 50` |
| two code names that collide in one table both take their code | RMA 117 and 277, both `Video games` -> `..._VIDEO_GAMES_117`, `..._VIDEO_GAMES_277` |

Resource **patterns** trap differently. AIP-123 requires a name to alternate
collection and identifier, and a singleton's name ends in a literal, so a
child of a singleton cannot hang beneath it. A meeting room is therefore
`hotels/{hotel}/meetingRooms/{meeting_room}`, not
`.../facilityInfo/meetingRooms/...`; the OTA path is in the annotation's
`path` either way.

## 12. Every citation carries its link

These comments become the generated documentation in every target language,
so a citation without a URL is a lookup the reader does by hand. AIP mentions
get `AIP-131 <https://aip.dev/131>`. OTA references name the XSD type and
path, and link the file on GitHub **at the pinned commit**, not at `master`,
so the link cannot drift from what the field was generated from.

## 13. Units are data, as OTA has them; money is a value type

This rule **inverts coversa's**. VSS pins a unit per signal and sends bare
numbers, so coversa records the unit in an annotation. OTA does the opposite:
it sends the unit beside the value -- `UnitOfMeasureCode`,
`DistanceUnitOfMeasureCode`, `AreaUnitOfMeasureCode` drawn from the Unit of
Measure Code (UOM) table. The schema follows the source: the unit is a field,
typed by the generated `UnitOfMeasureCode` enum like any other code.

Money follows `google.type.Money`'s shape without importing it (rule 4).
OTA's `CurrencyAmountGroup` -- `Amount`, `CurrencyCode`, `DecimalPlaces` --
becomes a generated `Money { currency_code, units, nano_units }` in each package
that needs one. `DecimalPlaces` is **consumed, not kept**: OTA defines it as
the ISO 4217 minor unit by which an integral `Amount` is scaled -- `8500` with
`DecimalPlaces="2"` is 85.00 -- so the codec divides on the way in, and
`units` plus `nanos` are exact without it. A bare `xs:decimal` that is not
money becomes a generated `Decimal { value }`, after `google.type.Decimal`.

Calendar values keep their calendar: see `docs/decisions.md` for `xs:date`,
the date-or-datetime unions, and `xs:duration`, which is not a `Duration`.

## 14. Two annotations, two jobs

**`google.api.field_behavior`** declares intent, and api-linter requires it on
every field. `IDENTIFIER` on a resource `name`; `REQUIRED` where the XSD says
`use="required"` or `minOccurs` of one or more; `OUTPUT_ONLY` per rule 8;
`IMMUTABLE` where a value may be set once, such as `ota_uid`; `OPTIONAL`
otherwise.

**`buf.validate`** (protovalidate) enforces. Constraints come from the XSD
facets, not from taste: `StringLength1to64` becomes `min_len: 1, max_len:
64`, an `xs:pattern` becomes `string.pattern`, `minInclusive` and
`maxInclusive` become numeric bounds, and a widened `nonNegativeInteger`
gains `gte: 0`. Where several apply they are **intersected** into one rule,
because protobuf accepts a single rule per field. An XSD pattern is implicitly
anchored and RE2 is not, so the generator adds `^...$`; XSD-only classes such
as `\i` and `\c` are a build error, not a best-effort translation. A type
restricted twice, each step with its own pattern, must match both, and
protovalidate takes one pattern per field: the last step is the `pattern`,
and each earlier one a CEL rule, `this.matches(r'...')`.

On a repeated field the item's rules sit under `repeated.items`, merged with
any `max_items` from `maxOccurs`. A bare `string = {...}` on a repeated field
compiles, and then fails every validation at run time.

One trap: a constraint on a possibly-unset field must carry
`(buf.validate.field).ignore = IGNORE_IF_ZERO_VALUE`, or an empty string fails
its pattern and every create request is rejected.

Optional XSD attributes of numeric or boolean type are proto3 `optional`,
because OTA distinguishes absent from zero: an absent `MinLOS` is no
restriction, `MinLOS="0"` is a value. AIP-149 asks for exactly this.

## 15. The generator fails loudly or not at all

`sync` parses only the XSD subset the pinned suite actually uses --
`complexType`, `simpleType`, `sequence`, `choice`, `all`, `extension`,
`restriction`, `group`, `attributeGroup`, `union`, `list`, and `any` where
OTA uses it for `TPA_Extensions` -- and errors on anything else. The input is
a published, regular corpus; a parser that quietly accepts an unfamiliar shape
produces a schema that is wrong in a way no linter catches.

The same holds downstream. An unknown XSD type is an error, not a skipped
field. A code-table reference that does not resolve against the workbook is
an error unless `CodeListAliases` names it, and so is a value a table
repeats, unless `CodeFixes` or `CodeRenumbers` does. Two OTA members that
become one field name, or two types one message name, are an error until a
rename separates them. Also enforced: `codes.lock` keeps an enum value's
name once emitted (`docs/decisions.md`); `model.checkMethods` rejects a
message in scope with no row; `model.CheckCatalogue` rejects an entry naming
a path the suite does not have. Run `just survey` after a
specification bump: it reports what was parsed and every name that trips a
known AIP rule, which is how the catalogue below is built and rechecked.

## The catalogue

OTA vocabulary and AIP rules disagree in the places below. The authority is
the code -- `sync/internal/catalog/catalog.go` and `renames.go` -- and every
entry there carries its reason; this table groups them. Once the survey
reports traps, `just survey` checks the two agree.

| Source (OTA path) | Becomes | Why |
| --- | --- | --- |
| `HotelReferenceGroup/@HotelName` | `display_name` on `Hotel`, `hotel_display_name` elsewhere | AIP-122 `_name` suffix |
| `HotelReferenceGroup/@ChainName`, `@BrandName` | `chain_display_name`, `brand_display_name` | same |
| `@ID` on a resource whose identity is a code | `ota_uid` | AIP-148; rule 10's round-trip identifier |
| `CountryNameType/@Code`, `@Country`, `@CountryCode` | `region_code` | AIP-143. The other `ISO3166` attributes keep their role and drop "country": `BirthCountry` -> `birth_region_code`, `Nationality` -> `nationality_region_code` |
| `PaymentCardType/@CountryOfIssue` | `issue_region_code` | AIP-140 and AIP-143 |
| `HotelReservationType/@ResStatus` | `state` or `gds_status_code` in `oneof lifecycle` | `InlineUnions`; verbs per `ActionValues` are methods |
| `RoomStayType/@RoomStayStatus` | `PmsReservationState`, OUTPUT_ONLY | `ServerAssigned`; AIP-216 |
| `List_OfferPaymentStatus`, `List_OfferReservationStatus` | `OfferPaymentState`, `OfferReservationState` | `TypeRenames`; an offer's lifecycle in a search response |
| `AdjustmentsType/@RequestID` | `adjustment_request_id` | AIP-155 reserves `request_id` for a caller's UUID |
| `RatePlanInclusionsType/RatePlanInclusionDesciption` | `description` | misspelt in the source |
| about 40 attributes with a preposition or a temporal suffix (`@OKToPublish`, `@ToFrom`, `@WindowBefore`, `@SetByTime`, `@TicketTime`, ...) | one rename each | AIP-140, AIP-142; `renames.go` |
| `AvailabilityStatusType`, `PMS_ResStatusType` and four anonymous `*Status` enums | `Availability`, `PmsReservationState`, `BookingAvailability`, ... | `TypeRenames`; AIP-216 |
| `MeetingRoomCodeType`, `CommissionType` | `MeetingRoomCodeItem`, `CommissionDetail` | `TypeRenames`; collide with a code table and a resource |
| `ProfileType`, wherever embedded | a `common.opentravel.org/Profile` name | `References`; rule 4 |
| citations of `BCC`, and `ErrorWarningAttributeGroup/@Code` | validated strings | `CodeListAliases` = `NoTable`: BCC is withdrawn, ERR is not a code table |
| `FAR` "Zone fare", numbered 39 | code 30 | `CodeRenumbers`; a workbook typo |
| `RLT` code 1, unnamed | "Away from the elevator" | `CodeFixes`, from the 2016 and 2017 workbooks |
| `@Removal`, `@Overwrite`, `@TransactionAction`, `@*NotifType`, by name | removed; each is a method | `Actions`; rule 8 |
| `DateTimeStampGroup/@CreateDateTime`, `@LastModifyDateTime`, `@LastUpdated` on a resource | absorbed into `create_time`, `update_time` | rule 8 |
| `TPA_Extensions`; the envelope | dropped by rule | `docs/decisions.md` |
