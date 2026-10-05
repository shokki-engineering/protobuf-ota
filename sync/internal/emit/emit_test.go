// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"reflect"
	"strings"
	"testing"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

func TestTypeRefQualifiesShadowedNames(t *testing.T) {
	p := &model.Package{Res: &model.Resource{Spec: "hotel", Domain: "revenue", Name: "Rate"}}
	const pkg = ".protobuf.opentravel.hotel.revenue.rate.v1."

	// GuestCount with a nested GuestCount: from inside, "GuestCount" binds to
	// the nested one, so GuestCount.GuestCount would look for a third.
	gc := &plan.Message{Name: "GuestCount"}
	inner := &plan.Message{Name: "GuestCount", Parent: gc}
	gc.Nested = []*plan.Message{inner}

	// A message with a nested Rate, referring to the top-level Rate.
	plan_ := &plan.Message{Name: "RatePlan"}
	nestedRate := &plan.Message{Name: "Rate", Parent: plan_}
	plan_.Nested = []*plan.Message{nestedRate}

	for _, c := range []struct {
		from      *plan.Message
		typ, want string
	}{
		{gc, "GuestCount.GuestCount", pkg + "GuestCount.GuestCount"},
		{inner, "GuestCount.GuestCount", pkg + "GuestCount.GuestCount"},
		{plan_, "Rate", pkg + "Rate"},
		{nestedRate, "Rate", pkg + "Rate"}, // its parent's member shadows it
		{gc, "Money", "Money"},
		{gc, "string", "string"},
		{gc, "google.protobuf.Timestamp", "google.protobuf.Timestamp"},
	} {
		if got := crossRef(p, c.from, c.typ, ""); got != c.want {
			t.Errorf("crossRef(from %s, %s) = %s, want %s", c.from.FullName(), c.typ, got, c.want)
		}
	}
}

// TestCrossRefSubpackages checks references between v1 and its types,
// codes and enums subpackages.
func TestCrossRefSubpackages(t *testing.T) {
	p := &model.Package{Res: &model.Resource{Spec: "hotel", Domain: "revenue", Name: "Rate"}}
	root := &plan.Message{Name: "Rate"}
	typ := &plan.Message{Name: "RoomStay", Sub: plan.SubTypes}
	shadow := &plan.Message{Name: "Fee", Sub: plan.SubTypes}
	shadow.Nested = []*plan.Message{{Name: "Money", Parent: shadow}}
	for _, c := range []struct {
		from            *plan.Message
		typ, file, want string
	}{
		{root, "RoomStay", "types/room_stay.proto", "types.RoomStay"},
		{typ, "RoomAmenityType", "codes/room_amenity_type.proto", "codes.RoomAmenityType"},
		{typ, "Money", "types/money.proto", "Money"},
		{typ, "Rate", "rate.proto", ".protobuf.opentravel.hotel.revenue.rate.v1.Rate"},
		{shadow, "Money", "types/money.proto", ".protobuf.opentravel.hotel.revenue.rate.v1.types.Money"},
	} {
		if got := crossRef(p, c.from, c.typ, c.file); got != c.want {
			t.Errorf("crossRef(from %s, %s in %s) = %s, want %s", c.from.Name, c.typ, c.file, got, c.want)
		}
	}
}

func TestGlob(t *testing.T) {
	for in, want := range map[string]string{
		"hotels/{hotel}/ratePlans/{rate_plan}": "hotels/*/ratePlans/*",
		"hotels/{hotel}/areaInfo":              "hotels/*/areaInfo",
		"profiles/{profile}":                   "profiles/*",
	} {
		if got := model.Glob(in); got != want {
			t.Errorf("Glob(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrap(t *testing.T) {
	for _, c := range []struct {
		in    string
		width int
		want  []string
	}{
		{"a b c", 3, []string{"a b", "c"}},
		{"averylongword b", 4, []string{"averylongword", "b"}}, // a word is never split
		{"", 10, []string{""}},
	} {
		if got := wrap(c.in, c.width); !reflect.DeepEqual(got, c.want) {
			t.Errorf("wrap(%q, %d) = %q, want %q", c.in, c.width, got, c.want)
		}
	}
}

func TestCommentReflowsParagraphs(t *testing.T) {
	var w writer
	w.comment(0, "one\ntwo\n\nthree")
	if got := w.String(); got != "// one two\n//\n// three\n" {
		t.Errorf("comment = %q", got)
	}
}

func TestImportFor(t *testing.T) {
	p := &model.Package{Res: &model.Resource{Spec: "hotel", Name: "Hotel"}}
	if got := importFor(p, "money.proto"); got != "protobuf/opentravel/hotel/hotel/v1/money.proto" {
		t.Errorf("local import = %q", got)
	}
	if got := importFor(p, "google/protobuf/timestamp.proto"); got != "google/protobuf/timestamp.proto" {
		t.Errorf("well-known import = %q", got)
	}
	if got := importFor(p, ""); got != "" {
		t.Errorf("scalar import = %q", got)
	}
	if !strings.HasPrefix(quote(`a"b\c`), `"a\"b\\c`) {
		t.Errorf("quote = %s", quote(`a"b\c`))
	}
}
