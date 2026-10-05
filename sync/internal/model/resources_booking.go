// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

// bookingResources are what a guest bought, and the group, event, finance
// and reporting business around it.
//
// Events, event reports and RFPs are root collections: each is addressed
// to several hotels at once, so no one hotel is its parent.
var bookingResources = []*Resource{
	{
		Spec: "hotel", Domain: "booking", Name: "Reservation", Parent: "Hotel",
		Root: "OTA_HotelResRQ/HotelReservations/HotelReservation",
	},
	{
		Spec: "hotel", Domain: "booking", Name: "Stay", Parent: "Reservation", Singleton: true,
		Root:  "OTA_HotelStayInfoNotifRQ/StayInfos/StayInfo",
		Carve: []string{"ReservationID", "HotelReservation"},
	},
	{
		Spec: "hotel", Domain: "sales", Name: "InventoryBlock", Parent: "Hotel",
		Root: "OTA_HotelInvBlockNotifRQ/InvBlocks/InvBlock", Identity: "InvBlockCode",
		Carve: []string{"HotelRef"},
	},
	{
		Spec: "hotel", Domain: "sales", Name: "Event", Root: "OTA_HotelEventRQ/Events/Event",
		Renames: map[string]string{"EventType/Event/@Name": "display_name"},
	},
	{Spec: "hotel", Domain: "sales", Name: "EventReport", Root: "OTA_HotelPostEventNotifRQ/EventReports/EventReport"},
	{Spec: "hotel", Domain: "sales", Name: "MeetingRfp", Root: "OTA_HotelRFP_MeetingRQ/RFP_RequestSegments/RFP_RequestSegment"},
	{
		Spec: "hotel", Domain: "sales", Name: "TransientRfp",
		Root: "OTA_HotelRFP_TransientRQ/TransientRFP_RequestSegments/TransientRFP_RequestSegment",
	},
	{
		Spec: "hotel", Domain: "finance", Name: "Commission", Parent: "Hotel",
		Root:  "OTA_HotelCommNotifRQ/CommissionEvents/CommissionEvent",
		Carve: []string{"HotelReference"},
	},
	{
		Spec: "hotel", Domain: "reporting", Name: "StatisticReport", Parent: "Hotel",
		Root:  "OTA_HotelStatsNotifRQ/Statistics/Statistic",
		Carve: []string{"HotelReferenceGroup/*"},
	},
}

// commonResources are what OTA defines across verticals.
var commonResources = []*Resource{
	{Spec: "common", Name: "Profile", Root: "OTA_ProfileCreateRQ/Profile"},
}
