// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

// Mapping is one row of the methods table: an OTA message and the RPCs that
// do its job, as "Resource.Method" (CLAUDE.md rule 7). Every message in
// scope has exactly one row; a message the schema does not take says why.
type Mapping struct {
	Message  string
	RPCs     []string
	NotTaken string
	Note     string
}

// methods is the table. CheckMethods fails the build on a message in scope
// with no row, a row for a message the suite does not have, or an RPC that
// no package declares.
var methods = []Mapping{
	{Message: "OTA_HotelDescriptiveContentNotifRQ", RPCs: []string{"Hotel.UpdateHotel", "Hotel.CreateHotel"},
		Note: "fanned out to each child resource; not atomic (docs/decisions.md)"},
	{Message: "OTA_HotelDescriptiveInfoRQ", RPCs: []string{"Hotel.BatchGetHotels"}},
	{Message: "OTA_HotelSearchRQ", RPCs: []string{"Hotel.SearchHotels"}},
	{Message: "OTA_HotelAvailRQ", RPCs: []string{"Hotel.SearchAvailability"}},
	{Message: "OTA_HotelAvailGetRQ", RPCs: []string{"Restriction.ListRestrictions"}},
	{Message: "OTA_HotelAvailNotifRQ", RPCs: []string{"Restriction.BatchUpdateRestrictions"}},
	{Message: "OTA_HotelInvCountNotifRQ", RPCs: []string{"Inventory.BatchUpdateInventories"}},
	{Message: "OTA_HotelInvCountRQ", RPCs: []string{"Inventory.ListInventories"}},
	{Message: "OTA_HotelInvAdjustRQ", RPCs: []string{"Inventory.AdjustInventories"}},
	{Message: "OTA_HotelInvSyncRQ", RPCs: []string{"Inventory.SyncInventories"}},
	{Message: "OTA_HotelRateAmountNotifRQ", RPCs: []string{"Rate.BatchUpdateRates"}},
	{Message: "OTA_HotelRatePlanNotifRQ", RPCs: []string{"RatePlan.CreateRatePlan", "RatePlan.UpdateRatePlan", "RatePlan.DeleteRatePlan"},
		Note: "by RatePlanNotifType"},
	{Message: "OTA_HotelRatePlanRQ", RPCs: []string{"RatePlan.ListRatePlans"}},
	{Message: "OTA_HotelBookingRuleNotifRQ", RPCs: []string{"BookingRule.BatchUpdateBookingRules"}},
	{Message: "OTA_HotelBookingRuleRQ", RPCs: []string{"BookingRule.ListBookingRules"}},
	{Message: "OTA_HotelSummaryNotifRQ", RPCs: []string{"Summary.BatchUpdateSummaries"}},
	{Message: "OTA_HotelInvNotifRQ", RPCs: []string{"RoomType.CreateRoomType", "RoomType.UpdateRoomType", "MeetingRoom.UpdateMeetingRoom"},
		Note: "by what each SellableProduct carries"},
	{Message: "OTA_HotelProductNotifRQ", RPCs: []string{"HotelProduct.CreateHotelProduct", "HotelProduct.UpdateHotelProduct", "HotelProduct.DeleteHotelProduct"},
		Note: "by ProductNotifType"},
	{Message: "OTA_HotelProductRQ", RPCs: []string{"HotelProduct.ListHotelProducts"}},
	{Message: "OTA_HotelResRQ", RPCs: []string{"Reservation.CreateReservation", "Reservation.QuoteReservation",
		"Reservation.HoldReservation", "Reservation.CommitReservation", "Reservation.IgnoreReservation"},
		Note: "by ResStatus: Book creates; Quote, Hold, Commit and Ignore are custom methods"},
	{Message: "OTA_HotelResNotifRQ", RPCs: []string{"Reservation.CreateReservation"}, Note: "with a caller-supplied reservation_id"},
	{Message: "OTA_HotelResModifyRQ", RPCs: []string{"Reservation.UpdateReservation"}},
	{Message: "OTA_HotelResModifyNotifRQ", RPCs: []string{"Reservation.UpdateReservation"}},
	{Message: "OTA_CancelRQ", RPCs: []string{"Reservation.CancelReservation"}, Note: "not DeleteReservation; a cancelled reservation persists"},
	{Message: "OTA_ReadRQ", RPCs: []string{"Reservation.GetReservation", "Reservation.ListReservations", "Profile.GetProfile"}},
	{Message: "OTA_DeleteRQ", RPCs: []string{"Profile.DeleteProfile"}},
	{Message: "OTA_HotelStayInfoNotifRQ", RPCs: []string{"Stay.UpdateStay"}},
	{Message: "OTA_HotelRoomListRQ", RPCs: []string{"Reservation.ImportReservations"}},
	{Message: "OTA_HotelInvBlockNotifRQ", RPCs: []string{"InventoryBlock.CreateInventoryBlock", "InventoryBlock.UpdateInventoryBlock",
		"InventoryBlock.DeleteInventoryBlock"}, Note: "by TransactionAction"},
	{Message: "OTA_HotelInvBlockRQ", RPCs: []string{"InventoryBlock.ListInventoryBlocks"}},
	{Message: "OTA_HotelEventRQ", RPCs: []string{"Event.CreateEvent"}},
	{Message: "OTA_HotelPostEventNotifRQ", RPCs: []string{"EventReport.CreateEventReport"}},
	{Message: "OTA_HotelPostEventRQ", RPCs: []string{"EventReport.ListEventReports"}},
	{Message: "OTA_HotelRFP_MeetingRQ", RPCs: []string{"MeetingRfp.CreateMeetingRfp"}},
	{Message: "OTA_HotelRFP_MeetingNotifRQ", RPCs: []string{"MeetingRfp.UpdateMeetingRfp"}},
	{Message: "OTA_HotelRFP_TransientRQ", RPCs: []string{"TransientRfp.CreateTransientRfp"}},
	{Message: "OTA_HotelRFP_TransientNotifRQ", RPCs: []string{"TransientRfp.UpdateTransientRfp"}},
	{Message: "OTA_HotelCommNotifRQ", RPCs: []string{"Commission.CreateCommission"}},
	{Message: "OTA_HotelStatsNotifRQ", RPCs: []string{"StatisticReport.CreateStatisticReport"}},
	{Message: "OTA_HotelStatsRQ", RPCs: []string{"StatisticReport.ListStatisticReports"}},
	{Message: "OTA_HotelCacheChangeRQ", RPCs: []string{"Rate.ListRates", "Restriction.ListRestrictions", "Inventory.ListInventories"},
		Note: "with an update_time filter, AIP-160"},
	{Message: "OTA_HotelCacheChangeNotifRQ", NotTaken: "a push of change events; AIP has no push method"},
	{Message: "OTA_HotelGetMsgRQ", NotTaken: "re-sends an earlier message, a transport concern Get and List make unnecessary"},
	{Message: "OTA_ProfileCreateRQ", RPCs: []string{"Profile.CreateProfile"}},
	{Message: "OTA_ProfileModifyRQ", RPCs: []string{"Profile.UpdateProfile"}},
	{Message: "OTA_ProfileMergeRQ", RPCs: []string{"Profile.MergeProfiles"}},
}
