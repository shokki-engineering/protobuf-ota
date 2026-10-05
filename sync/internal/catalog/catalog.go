// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package catalog is the generator's data: every place OTA's vocabulary and
// an AIP rule disagree, and how each is resolved. It is data, not logic, and
// every entry is also a row in the catalogue at the foot of
// docs/conventions.md.
//
// Every key is an OTA path: a named XSD type or attribute group, then the
// path inside it, attributes marked `@` -- "HotelReferenceGroup/@HotelName".
// That is the reader's Key for the declaration, so an entry applies wherever
// the type or group is used. An entry naming a path the suite does not have
// is a build error (see model.CheckCatalogue).
package catalog

// AcceptedTraps are names a rule flags but that are right as they stand,
// with the reason. The survey stops reporting them.
var AcceptedTraps = map[string]string{}

// TypeRenames gives a message or enum a name other than the one its XSD
// name produces. Keyed by XSD type name, or by declaring path for an
// anonymous type.
var TypeRenames = map[string]string{
	// AIP-216 reserves *Status for lifecycle state, and a *State enum must be
	// OUTPUT_ONLY. Availability is set by the hotel, not by a method.
	"AvailabilityStatusType": "Availability",
	// A room stay's PMS status -- in-house, checked out, no-show -- is
	// lifecycle the hotel reports: a State, OUTPUT_ONLY (CLAUDE.md rule 8).
	"PMS_ResStatusType":                      "PmsReservationState",
	"InvBlockGroup/@BookingStatus":           "BookingAvailability",
	"CustomerType/Address/@ValidationStatus": "AddressValidation",
	"CustomerType/@MaritalStatus":            "MaritalCondition",
	"ShareAllGroup/@OptInStatus":             "OptInChoice",
	// Collides with the Meeting Room Code (MRC) code table's enum.
	"MeetingRoomCodeType": "MeetingRoomCodeItem",
	// An offer's payment and reservation status in a search response are
	// lifecycle the supplier reports: States, OUTPUT_ONLY.
	"List_OfferPaymentStatus":     "OfferPaymentState",
	"List_OfferReservationStatus": "OfferReservationState",
	// Collides with the Commission resource.
	"CommissionType": "CommissionDetail",
}

// ServerAssigned are the attributes and elements the responding system
// sets: OUTPUT_ONLY (CLAUDE.md rule 8).
var ServerAssigned = map[string]bool{
	"RoomStayType/@RoomStayStatus": true,
}

// Actions are attributes that carry an operation rather than data, by
// attribute name, wherever they occur. None becomes a field; each is a
// method (CLAUDE.md rule 8). The value says what the operation becomes.
var Actions = map[string]string{
	"Removal":           "Delete of the item, or omitting it from an Update",
	"Overwrite":         "Update with an update_mask naming the replaced fields",
	"TransactionAction": "the Create, Update, Delete or custom method that performs it",
	"RatePlanNotifType": "CreateRatePlan, UpdateRatePlan or DeleteRatePlan",
	"ProductNotifType":  "CreateHotelProduct, UpdateHotelProduct or DeleteHotelProduct",
	"InvNotifType":      "CreateRoomType, UpdateRoomType or DeleteRoomType",
}

// Dropped are declarations that become nothing, with the reason. Rules
// drop more than this -- TPA_Extensions everywhere, the envelope -- and
// those are not listed.
var Dropped = map[string]string{}

// References are XSD types that are resources in this schema. Wherever OTA
// embeds one, the field holds the resource's name instead of a copy
// (CLAUDE.md rule 4: a genuine association is a resource name). OTA embeds
// a whole customer profile -- airline seat and rail preferences included --
// wherever it means "this person", even to say who may view a hotel's
// content.
var References = map[string]string{
	"ProfileType": "common.opentravel.org/Profile",
}

// CodeListAliases resolves a code-table citation that does not name a table
// the workbook has, keyed by the citing declaration. NoTable keeps the
// value a validated string.
var CodeListAliases = map[string]string{
	// OTA's error codes are a separate sheet of the workbook, for the
	// envelope's Errors, not a code table.
	"ErrorWarningAttributeGroup/@Code": NoTable,
	// Brand Category Code (BCC) is withdrawn; OTA points to Segment
	// Category Code (SEG), whose codes are not BCC's. No table is safer
	// than the wrong one.
	"PropertyClassPrefType/@PropertyClassType":       NoTable,
	"AffiliationInfoType/Brands/Brand/@CategoryCode": NoTable,
}

// NoTable is a CodeListAliases value: the citation names no usable table.
const NoTable = "-"

// CodeRenumbers corrects a code the workbook numbers wrongly, keyed
// "ABBR/code name".
var CodeRenumbers = map[string]int{
	// Listed between 29 and 31 as 39, which One way fare also is; 30 is
	// otherwise missing.
	"FAR/Zone fare (list of zones fare can be used for)": 30,
}

// CodeFixes resolves a defect in the code list workbook for a table some
// field in scope cites: a code with no name, or a value given twice.
// Keyed "ABBR/value".
var CodeFixes = map[string]string{
	// Unnamed in the 2024-09-04 workbook; "Away from the elevator" in the
	// 2016-02-24 and 2017-10-25 workbooks, the latter shipped with 2017B.
	"RLT/1": "Away from the elevator",
}

// InlineUnion places an XSD union's alternatives directly in the message as
// one oneof, rather than in a wrapper message, with its enumerations merged
// into one enum. It is how ResStatus becomes the reservation's lifecycle
// (docs/decisions.md, "An XSD union is a oneof").
type InlineUnion struct {
	OneOf      string // the oneof's name
	Enum       string // the merged enum's name
	EnumField  string // the field holding the merged enum
	TextField  string // the field holding the free-text alternative
	OutputOnly bool
}

// InlineUnions are keyed by the attribute's OTA path.
var InlineUnions = map[string]InlineUnion{
	"HotelReservationType/@ResStatus": {
		OneOf: "lifecycle", Enum: "State",
		EnumField: "state", TextField: "gds_status_code", OutputOnly: true,
	},
}

// ActionValues are the values of OTA's TransactionActionType that are
// verbs. Each is a method (CLAUDE.md rule 8); the type's other values are
// states, and are kept.
var ActionValues = map[string]bool{
	"Book": true, "Quote": true, "Hold": true, "Initiate": true, "Ignore": true,
	"Modify": true, "Commit": true, "Cancel": true, "CommitOverrideEdits": true,
	"VerifyPrice": true, "Ticket": true,
}
