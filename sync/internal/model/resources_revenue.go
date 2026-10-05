// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

// revenueResources are what a hotel sells, and when and for how much.
//
// The four daily resources -- Inventory, Rate, Restriction, BookingRule --
// are OTA's availability, rate and inventory messages. OTA keys each by a
// StatusApplicationControl: a date range, a weekday mask, a room type and a
// rate plan. A range has no identity, so each (room type, rate plan, day)
// is a resource and the StatusApplicationControl is its name, not a field
// (docs/decisions.md).
var revenueResources = []*Resource{
	{
		Spec: "hotel", Domain: "revenue", Name: "RatePlan", Parent: "Hotel",
		Root: "OTA_HotelRatePlanNotifRQ/RatePlans/RatePlan", Identity: "RatePlanCode",
		Carve: []string{"HotelRef"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "HotelProduct", Parent: "Hotel",
		Root:  "OTA_HotelProductNotifRQ/HotelProducts/HotelProduct",
		Carve: []string{"HotelReferenceGroup/*"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "Inventory", Parent: "RoomType",
		Root:  "OTA_HotelInvCountNotifRQ/Inventories/Inventory",
		Carve: []string{"StatusApplicationControl"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "Rate", Parent: "RatePlan",
		Root:  "OTA_HotelRateAmountNotifRQ/RateAmountMessages/RateAmountMessage",
		Carve: []string{"StatusApplicationControl"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "Restriction", Parent: "RatePlan",
		Root:  "OTA_HotelAvailNotifRQ/AvailStatusMessages/AvailStatusMessage",
		Carve: []string{"StatusApplicationControl"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "BookingRule", Parent: "RatePlan",
		Root:  "OTA_HotelBookingRuleNotifRQ/RuleMessages/RuleMessage",
		Carve: []string{"StatusApplicationControl"},
	},
	{
		Spec: "hotel", Domain: "revenue", Name: "Summary", Parent: "Hotel",
		Root:  "OTA_HotelSummaryNotifRQ/HotelSummaryMessages/HotelSummaryMessage",
		Carve: []string{"DateTimeSpanGroup/*", "DOW_PatternGroup/*"},
	},
}
