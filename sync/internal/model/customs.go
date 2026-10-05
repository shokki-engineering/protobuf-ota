// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

// Custom is a method beyond a resource's standard set: an AIP-136 custom
// method, an AIP-231 batch get or an AIP-234 batch update.
type Custom struct {
	Resource string
	Name     string // the RPC, e.g. "CancelReservation"
	Kind     string // KindResource, KindCollection, KindBatchGet or KindBatchUpdate
	Verb     string // the URI's ":verb", for a custom method

	// Request is the OTA message whose content, less its envelope, the
	// request carries after `name` or `parent`. Empty for none.
	Request string
	// Skip lists further members of Request the method does not carry.
	Skip []string
	// Body names the resource the request carries whole, for a method that
	// takes one: QuoteReservation prices a reservation it does not keep.
	Body bool

	// Response is the OTA message whose content, less its envelope, the
	// response carries; ResponseList returns the resources it made; empty
	// for both returns the resource.
	Response     string
	ResponseList bool
	// ResponseOne returns the resource inside a response message: a
	// collection-level method returns {Name}Response, not the resource
	// (AIP-136).
	ResponseOne bool

	// Paginated adds AIP-158 paging, and Results names the response's
	// repeated field that is paged, which AIP-158 wants first.
	Paginated bool
	Results   string

	Doc string
}

const (
	KindResource    = "resource"    // POST /v1/{name=...}:verb
	KindCollection  = "collection"  // POST /v1/{parent=...}/plural:verb
	KindBatchGet    = "batchGet"    // GET  /v1/{parent=...}/plural:batchGet
	KindBatchUpdate = "batchUpdate" // POST /v1/{parent=...}/plural:batchUpdate
)

var customs = []Custom{
	{Resource: "Hotel", Name: "BatchGetHotels", Kind: KindBatchGet,
		Doc: "Retrieves several hotels at once. OTA_HotelDescriptiveInfoRQ."},
	{Resource: "Hotel", Name: "SearchHotels", Kind: KindCollection, Verb: "search",
		Request: "OTA_HotelSearchRQ", Response: "OTA_HotelSearchRS", Paginated: true, Results: "properties",
		Doc: "Finds hotels by location, amenity and other criteria. OTA_HotelSearchRQ."},
	{Resource: "Hotel", Name: "SearchAvailability", Kind: KindCollection, Verb: "searchAvailability",
		Request: "OTA_HotelAvailRQ", Response: "OTA_HotelAvailRS", Paginated: true, Results: "hotel_stays",
		Doc: "Prices and offers rooms for a stay. Returns offers, which are not resources. OTA_HotelAvailRQ."},
	{Resource: "Reservation", Name: "CancelReservation", Kind: KindResource, Verb: "cancel",
		Request: "OTA_CancelRQ", Skip: []string{"Segment", "OriginAndDestinationSegment"},
		Doc: "Cancels a reservation: state moves to CANCELLED and the reservation is kept, with its cancellation number. OTA_CancelRQ."},
	{Resource: "Reservation", Name: "QuoteReservation", Kind: KindCollection, Verb: "quote", Body: true, ResponseOne: true,
		Doc: "Prices a reservation without booking it. OTA_HotelResRQ with ResStatus Quote."},
	{Resource: "Reservation", Name: "HoldReservation", Kind: KindResource, Verb: "hold",
		Doc: "Holds a reservation's inventory without confirming it. OTA_HotelResRQ with ResStatus Hold."},
	{Resource: "Reservation", Name: "CommitReservation", Kind: KindResource, Verb: "commit",
		Doc: "Confirms a reservation created in the pending state. OTA_HotelResRQ with ResStatus Commit."},
	{Resource: "Reservation", Name: "IgnoreReservation", Kind: KindResource, Verb: "ignore",
		Doc: "Abandons a pending reservation, releasing its inventory. OTA_HotelResRQ with ResStatus Ignore."},
	{Resource: "Reservation", Name: "ImportReservations", Kind: KindCollection, Verb: "import",
		Request: "OTA_HotelRoomListRQ", ResponseList: true,
		Doc: "Books a group or tour rooming list against the hotel's blocks. OTA_HotelRoomListRQ."},
	{Resource: "Inventory", Name: "BatchUpdateInventories", Kind: KindBatchUpdate,
		Doc: "Sets inventory counts for many days at once. OTA_HotelInvCountNotifRQ, its ranges expanded to days."},
	{Resource: "Inventory", Name: "AdjustInventories", Kind: KindCollection, Verb: "adjust",
		Request: "OTA_HotelInvAdjustRQ", ResponseList: true,
		Doc: "Adjusts inventory by the reservations that consume it. OTA_HotelInvAdjustRQ."},
	{Resource: "Inventory", Name: "SyncInventories", Kind: KindCollection, Verb: "sync",
		Request: "OTA_HotelInvSyncRQ", ResponseList: true,
		Doc: "Reconciles inventory with another system's view of it. OTA_HotelInvSyncRQ."},
	{Resource: "Rate", Name: "BatchUpdateRates", Kind: KindBatchUpdate,
		Doc: "Sets rates for many days at once. OTA_HotelRateAmountNotifRQ, its ranges expanded to days."},
	{Resource: "Restriction", Name: "BatchUpdateRestrictions", Kind: KindBatchUpdate,
		Doc: "Sets restrictions for many days at once. OTA_HotelAvailNotifRQ, its ranges expanded to days."},
	{Resource: "BookingRule", Name: "BatchUpdateBookingRules", Kind: KindBatchUpdate,
		Doc: "Sets booking rules for many days at once. OTA_HotelBookingRuleNotifRQ."},
	{Resource: "Summary", Name: "BatchUpdateSummaries", Kind: KindBatchUpdate,
		Doc: "Sets the hotel's open or closed status for many days at once. OTA_HotelSummaryNotifRQ."},
	{Resource: "Profile", Name: "MergeProfiles", Kind: KindCollection, Verb: "merge",
		Request: "OTA_ProfileMergeRQ",
		Doc:     "Merges obsolete profiles into one, which is returned. OTA_ProfileMergeRQ."},
}
