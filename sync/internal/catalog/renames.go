// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package catalog

// FieldRenames gives a field a name other than the one its OTA name
// snake-cases to.
var FieldRenames = map[string]string{
	// AIP-122: only display_name, bare or prefixed, may end in _name.
	"HotelReferenceGroup/@HotelName": "hotel_display_name",
	"HotelReferenceGroup/@ChainName": "chain_display_name",
	"HotelReferenceGroup/@BrandName": "brand_display_name",
	// Misspelt in the source.
	"RatePlanInclusionsType/RatePlanInclusionDesciption": "description",

	// AIP-140 prepositions, one at a time. The phrases OTA repeats -- "unit
	// of measure", "amount before tax", "length of stay" -- are rewritten by
	// rule in plan.rephrase and are not listed.
	"ContactsType/Name/@OKToPublish":                                           "publishable",
	"HotelInfoType/Services/Service/@AvailableToAnyGuest":                      "available_all_guests",
	"RelativePositionType/@ToFrom":                                             "facility_direction",
	"TimeInstantType/@WindowAfter":                                             "late_window",
	"TimeInstantType/@WindowBefore":                                            "early_window",
	"CommissionType/@BillToID":                                                 "billing_id",
	"HotelRatePlanType/Supplements/Supplement/@AddToBasicRateIndicator":        "added_basic_rate_indicator",
	"OfferType/FreeUpgrade/UpgradeFrom":                                        "upgrade_source",
	"OfferType/FreeUpgrade/UpgradeTo":                                          "upgrade_target",
	"OfferType/Inventories/Inventory/@AppliesToIndicator":                      "applicable_indicator",
	"OfferType/OfferRules/OfferRule/Inventories/Inventory/@AppliesToIndicator": "applicable_indicator",
	"OfferType/OfferRules/OfferRule/@StayOverDate":                             "stayover_weekday",
	"RateUploadType/@StayOverDate":                                             "stayover_weekday",
	"RateUploadType/BaseByGuestAmts":                                           "base_guest_amounts",
	"PaymentCardType/@CountryOfIssue":                                          "issue_region_code",
	"PaymentFormType/Ticket/@ReasonForReroute":                                 "reroute_reason",

	// AIP-142: a non-timestamp whose name ends in a temporal word.
	"HotelInfoType/Descriptions/Renovation/@RenovationCompletionDate": "renovation_completion",
	"TransportationType/Transportation/@TypicalTravelTime":            "typical_travel_duration",
	"DeadlineGroup/@OffsetDropTime":                                   "offset_drop_point",
	"LengthsOfStayType/LengthOfStay/@Time":                            "length",
	"PaymentCardDateGroup/@EffectiveDate":                             "effective_month_year",
	"PaymentCardDateGroup/@ExpireDate":                                "expiry_month_year",

	"ProviderGroup/@ProvidedBy": "provider",
	"PoliciesType/Policy/TaxPolicies/TaxPolicy/@NightsForTaxExemptionQuantity":                                    "tax_exempt_nights",
	"PoliciesType/Policy/RatePolicies/RatePolicy/@SubjectToChangeInd":                                             "changeable",
	"HotelReservationIDsType/HotelReservationID/@ForGuest":                                                        "guest_facing",
	"AmountType/Discount/@AppliesTo":                                                                              "application",
	"AreaInfoType/Recreations/Recreation/MultimediaDescriptions/@RefDirectionTo":                                  "reference_direction",
	"DescriptionType/@RefDirectionTo":                                                                             "reference_direction",
	"InvBlockRoomType/RoomTypeAllocations/RoomTypeAllocation/RoomTypeAllocByGuest":                                "guest_allocations",
	"FacilityInfoType/GuestRooms/GuestRoom/Amenities/Amenity/@IncludedInRateIndicator":                            "rate_included_indicator",
	"RailPassengerCategoryType/@AccompaniedByInfantInd":                                                           "infant_accompanied",
	"SeatRequestAttributes/@SeatInRow":                                                                            "row_seat",
	"AmountType/@StayOverDate":                                                                                    "stayover_weekday",
	"EventType/Event/Sites/Site/EventDays/EventDay/EventDayFunctions/EventDayFunction/RoomSetup/@SetByTime":       "setup_deadline",
	"EventType/Event/Sites/Site/EventDays/EventDay/EventDayFunctions/EventDayFunction/RoomSetup/@DismantleByTime": "dismantle_deadline",

	// AIP-142, continued.
	"OTA_HotelAvailRQ/AvailRequestSegments/@MaximumWaitTime":               "maximum_wait",
	"OTA_HotelAvailRS/HotelStays/HotelStay/Availability/Restriction/@Time": "length",
	"CancelRuleType/@CancelByDate":                                         "cancel_deadline",
	// AIP-155 reserves request_id for a UUID the caller makes for retry
	// safety; OTA's RequestID is its own reference for the adjustment.
	"AdjustmentsType/@RequestID":                                                      "adjustment_request_id",
	"TransportInfoType/TransportInfo/@Time":                                           "transport_time",
	"TicketDistribPrefType/@TicketTime":                                               "ticketing_moment",
	"RFP_RequestSegmentsType/RFP_RequestSegment/Sites/Site/@ApproxSiteInspectionDate": "approx_site_inspection",
	"RFP_RequestSegmentsType/RFP_RequestSegment/Dates/Date":                           "date_options",
	"PostEventSiteReportType/Date":                                                    "event_dates",
	"DeliveryResponseGroup/@MinimumTime":                                              "minimum_duration",
	"DeliveryResponseGroup/@MaximumTime":                                              "maximum_duration",
	"AvailStatusMessageType/BestAvailableRates/BestAvailableRate/@LengthOfStayTime":   "stay_length",

	// Two OTA members that snake-case to one name.
	"PaymentCardType/@SignatureOnFileInd": "signature_on_file_held",
}
