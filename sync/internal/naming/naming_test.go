// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package naming

import "testing"

func TestSnake(t *testing.T) {
	for in, want := range map[string]string{
		"RatePlanCode":     "rate_plan_code",
		"ResID_Value":      "res_id_value",
		"PMS_InvBlockID":   "pms_inv_block_id",
		"ISOCountry":       "iso_country",
		"TPA_Extensions":   "tpa_extensions",
		"Bed & breakfast":  "bed_and_breakfast",
		"Children's menu":  "childrens_menu",
		"No-show":          "no_show",
		"24 hour desk":     "24_hour_desk",
		"HotelInformation": "hotel_info",
		"CouponRPHs":       "coupon_rphs",
		"RPHs":             "rphs",
	} {
		if got := Snake(in); got != want {
			t.Errorf("Snake(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTypeName(t *testing.T) {
	for in, want := range map[string]string{
		"RatePlanType":  "RatePlan",
		"UniqueID_Type": "UniqueId",
		"POS_Type":      "Pos",
		"Type":          "Type",
		"RoomStays":     "RoomStays",
	} {
		if got := TypeName(in); got != want {
			t.Errorf("TypeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPluralise(t *testing.T) {
	for in, want := range map[string]string{
		"ratePlan":   "ratePlans",
		"rate_plan":  "rate_plans",
		"policy":     "policies",
		"summary":    "summaries",
		"multimedia": "multimedia",
		"inventory":  "inventories",
		"area_info":  "area_info",
		"tax":        "taxes",
		"day":        "days",
	} {
		if got := Pluralise(in); got != want {
			t.Errorf("Pluralise(%q) = %q, want %q", in, got, want)
		}
	}
}
