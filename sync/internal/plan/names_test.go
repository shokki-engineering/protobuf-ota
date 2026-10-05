// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"strings"
	"testing"
)

// testScope is a scope over an empty package, planning into a top-level
// value message unless the caller says otherwise.
func testScope() *scope {
	return &scope{b: New(nil), m: &Message{Name: "Value"}, oneof: map[int][]*Field{}}
}

func TestFieldNameRules(t *testing.T) {
	str, boolean := ftype{Kind: kindString}, ftype{Kind: kindBool}
	for _, c := range []struct {
		xml  string
		t    ftype
		want string
	}{
		{"URL", str, "uri"},                                         // AIP-140 uri
		{"WingIdentifier", str, "wing_id"},                          // AIP-140 abbreviation
		{"IsRoom", boolean, "room"},                                 // boolean Is prefix
		{"FeaturedInd", boolean, "featured"},                        // boolean Ind suffix
		{"Ind", boolean, "ind"},                                     // a lone word is kept
		{"Rate", ftype{Kind: kindMessage, Repeated: true}, "rates"}, // AIP-144 plural
		{"GuaranteesAccepted", ftype{Kind: kindMessage, Repeated: true}, "guarantees_accepted"},
		{"RoomTypeName", str, "room_type_display_name"},         // AIP-122 _name
		{"Name", str, "display_name"},                           // top-level, not a resource
		{"CountryCode", str, "region_code"},                     // AIP-143
		{"Language", str, "language_code"},                      // AIP-143
		{"TimeZone", str, "time_zone"},                          // already AIP-143's
		{"Default", str, "default_value"},                       // AIP-140 reserved word
		{"UnitOfMeasureCode", str, "unit_code"},                 // phrase rule
		{"NumberOfNights", ftype{Kind: kindInt}, "night_count"}, // phrase rule
		{"AmountBeforeTax", ftype{Kind: kindMessage}, "pre_tax_amount"},
	} {
		s := testScope()
		if got := s.fieldName(c.xml, "Test/@"+c.xml, c.t); got != c.want {
			t.Errorf("fieldName(%q) = %q, want %q", c.xml, got, c.want)
		}
		if errs := s.b.Errors(); len(errs) > 0 {
			t.Errorf("fieldName(%q): unexpected errors %v", c.xml, errs)
		}
	}
}

func TestFieldNameNestedKeepsName(t *testing.T) {
	s := testScope()
	s.m = &Message{Name: "Inner", Parent: &Message{Name: "Outer"}}
	if got := s.fieldName("Name", "Test/@Name", ftype{Kind: kindString}); got != "name" {
		t.Errorf("nested Name = %q; api-linter only reads a top-level message's name field", got)
	}
}

func TestFieldNameReportsWhatNoRuleFixes(t *testing.T) {
	for _, c := range []struct {
		xml, want string
	}{
		{"WindowAfter", "preposition"},
		{"24Hour", "digit"},
		{"StayTime", "timestamp"}, // an int ending in time (AIP-142)
	} {
		s := testScope()
		kind := kindString
		if c.xml == "StayTime" {
			kind = kindInt
		}
		s.fieldName(c.xml, "Test/@"+c.xml, ftype{Kind: kind, Type: "int32"})
		errs := strings.Join(s.b.Errors(), "\n")
		if !strings.Contains(errs, c.want) {
			t.Errorf("fieldName(%q): errors %q, want one mentioning %q", c.xml, errs, c.want)
		}
	}
}

func TestResourceIDIsReported(t *testing.T) {
	s := testScope()
	s.m.Resource = &Resource{}
	s.fieldName("ID", "Test/@ID", ftype{Kind: kindString})
	if len(s.b.Errors()) == 0 {
		t.Error("a resource field named id must be reported (AIP-148)")
	}
}

func TestTemporalNames(t *testing.T) {
	for _, c := range []struct {
		xml  string
		kind string
		want string
	}{
		{"LastUpdated", kindTimestamp, "last_update_time"}, // imperative mood
		{"CreateDateTime", kindTimestamp, "create_date_time"},
		{"ExpireDate", kindTimestamp, "expire_time"},
		{"Start", kindTimestamp, "start_time"},
		{"PurgeDate", kindDate, "purge_day"},
		{"ArrivalTime", kindTimeOfDay, "arrival_clock"},
		{"Hold", kindTimeOfDay, "hold_clock"},
		{"CheckInTime", kindTimeOfDay, "checkin_clock"},
		{"ArrivalDate", kindUnion, "arrival"},
		{"Date", kindUnion, "when"},
		{"End", kindUnion, "end"},
	} {
		s := testScope()
		if got := s.fieldName(c.xml, "Test/@"+c.xml, ftype{Kind: c.kind}); got != c.want {
			t.Errorf("fieldName(%q, %s) = %q, want %q", c.xml, c.kind, got, c.want)
		}
		if errs := s.b.Errors(); len(errs) > 0 {
			t.Errorf("fieldName(%q): unexpected errors %v", c.xml, errs)
		}
	}
}

func TestRephrase(t *testing.T) {
	for in, want := range map[string]string{
		"unit_of_measure_code":            "unit_code",
		"area_unit_of_measure_code":       "area_unit_code",
		"number_of_nights":                "night_count",
		"total_number_of_units":           "total_unit_count",
		"nbr_of_bedrooms":                 "bedroom_count",
		"amount_before_tax":               "pre_tax_amount",
		"amount_after_tax":                "post_tax_amount",
		"amount_including_markup":         "marked_up_amount",
		"first_stay_day_of_week":          "first_stay_weekday",
		"check_in":                        "checkin",
		"walk_in_indicator":               "walkin_indicator",
		"in_house_time_span":              "inhouse_time_span",
		"lengths_of_stay":                 "stay_lengths",
		"source_of_business":              "business_source",
		"percent_of_renovation_completed": "renovation_completed_percent",
		"rate_plan_code":                  "rate_plan_code",
	} {
		if got := rephrase(in); got != want {
			t.Errorf("rephrase(%q) = %q, want %q", in, got, want)
		}
	}
}
