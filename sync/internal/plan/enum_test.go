// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"testing"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

func TestNameValues(t *testing.T) {
	b := New(nil)
	e := &Enum{Name: "RoomAmenityType", Values: []*EnumValue{
		{Number: 1, Spelling: ""},
		{Number: 117, Spelling: "Video games"},
		{Number: 277, Spelling: "Video games"},
		{Number: 5, Spelling: "Unspecified"},
		{Number: 50, Spelling: "Hairdryer"},
		{Number: 9, Spelling: "Bed & breakfast"},
	}}
	b.nameValues(e, "Test")
	want := []string{
		"ROOM_AMENITY_TYPE_EMPTY",
		"ROOM_AMENITY_TYPE_VIDEO_GAMES_117",
		"ROOM_AMENITY_TYPE_VIDEO_GAMES_277",
		"ROOM_AMENITY_TYPE_UNSPECIFIED_5", // never the zero value's name
		"ROOM_AMENITY_TYPE_HAIRDRYER",
		"ROOM_AMENITY_TYPE_BED_AND_BREAKFAST",
	}
	for i, v := range e.Values {
		if v.Name != want[i] {
			t.Errorf("value %d (%q) = %s, want %s", v.Number, v.Spelling, v.Name, want[i])
		}
	}
	if errs := b.Errors(); len(errs) > 0 {
		t.Errorf("unexpected errors %v", errs)
	}
}

func TestNameValuesReportsNoIdentifier(t *testing.T) {
	b := New(nil)
	b.nameValues(&Enum{Name: "Foo", Values: []*EnumValue{{Number: 1, Spelling: "!!!"}}}, "Test")
	if len(b.Errors()) == 0 {
		t.Error("a value with no identifier characters must be reported")
	}
}

func TestWrapped(t *testing.T) {
	rate := &xsd.Elem{Name: "Rate", Max: xsd.Unbounded}
	tpa := &xsd.Elem{Name: "TPA_Extensions", Max: 1}
	single := &xsd.Elem{Name: "Rate", Max: 1}
	for _, c := range []struct {
		name string
		c    *xsd.Complex
		want *xsd.Elem
	}{
		{"pure wrapper", &xsd.Complex{Elems: []*xsd.Elem{rate}}, rate},
		{"wrapper beside TPA_Extensions", &xsd.Complex{Elems: []*xsd.Elem{rate, tpa}}, rate},
		{"wrapper with an attribute", &xsd.Complex{Attrs: []*xsd.Attr{{Name: "X"}}, Elems: []*xsd.Elem{rate}}, nil},
		{"wrapper with text", &xsd.Complex{Text: &xsd.Simple{Base: "string"}, Elems: []*xsd.Elem{rate}}, nil},
		{"single child", &xsd.Complex{Elems: []*xsd.Elem{single}}, nil},
		{"two children", &xsd.Complex{Elems: []*xsd.Elem{rate, {Name: "Other", Max: xsd.Unbounded}}}, nil},
		{"nil", nil, nil},
	} {
		if got := wrapped(c.c); got != c.want {
			t.Errorf("%s: wrapped = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestMoneyAttrs(t *testing.T) {
	money := &xsd.Simple{Name: "Money", Base: "decimal"}
	withCurrency := &xsd.Complex{Attrs: []*xsd.Attr{
		{Name: "CurrencyCode", Simple: &xsd.Simple{Base: "string"}},
		{Name: "DecimalPlaces", Simple: &xsd.Simple{Base: "nonNegativeInteger"}},
		{Name: "Amount", Simple: money},
		{Name: "Percent", Simple: &xsd.Simple{Name: "Percentage", Base: "decimal"}},
	}}
	got := moneyAttrs(withCurrency)
	if !got["Amount"] || got["Percent"] || len(got) != 1 {
		t.Errorf("moneyAttrs = %v, want only Amount", got)
	}
	without := &xsd.Complex{Attrs: []*xsd.Attr{{Name: "Amount", Simple: money}}}
	if got := moneyAttrs(without); len(got) != 0 {
		t.Errorf("moneyAttrs without a currency = %v; an amount with no currency is a Decimal", got)
	}
}

func TestCodeTypeRecognised(t *testing.T) {
	if !isCode(&xsd.Simple{Patterns: [][]string{{otaCode}}}) {
		t.Error("OTA_CodeType's pattern not recognised")
	}
	if isCode(&xsd.Simple{Patterns: [][]string{{"[A-Z]{3}"}}}) {
		t.Error("a non-code pattern taken for OTA_CodeType")
	}
}
