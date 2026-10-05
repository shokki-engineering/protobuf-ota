// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"strings"
	"testing"
)

func TestPackageNames(t *testing.T) {
	hotel := &Package{Res: &Resource{Spec: "hotel", Name: "Hotel"}}
	plan := &Package{Res: &Resource{Spec: "hotel", Domain: "revenue", Name: "RatePlan"}, Parent: hotel}
	area := &Package{Res: &Resource{Spec: "hotel", Domain: "property", Name: "AreaInfo", Singleton: true}, Parent: hotel}
	media := &Package{Res: &Resource{Spec: "hotel", Domain: "property", Name: "Multimedia"}, Parent: hotel}

	for _, c := range []struct {
		p                                         *Package
		pattern, re, proto, service, sing, plural string
	}{
		{hotel, "hotels/{hotel}", "^hotels/[^/]+$", "protobuf.opentravel.hotel.hotel.v1", "Hotels", "hotel", "hotels"},
		{plan, "hotels/{hotel}/ratePlans/{rate_plan}", "^hotels/[^/]+/ratePlans/[^/]+$",
			"protobuf.opentravel.hotel.revenue.rate_plan.v1", "RatePlans", "ratePlan", "ratePlans"},
		// A singleton ends in a literal (AIP-156).
		{area, "hotels/{hotel}/areaInfo", "^hotels/[^/]+/areaInfo$",
			"protobuf.opentravel.hotel.property.area_info.v1", "AreaInfoService", "areaInfo", "areaInfo"},
		// A plural that is its singular cannot name the service too.
		{media, "hotels/{hotel}/multimedia/{multimedia}", "^hotels/[^/]+/multimedia/[^/]+$",
			"protobuf.opentravel.hotel.property.multimedia.v1", "MultimediaService", "multimedia", "multimedia"},
	} {
		got := []string{c.p.Pattern(), Regexp(c.p.Pattern()), c.p.Proto(), c.p.Service(), c.p.Singular(), c.p.Plural()}
		want := []string{c.pattern, c.re, c.proto, c.service, c.sing, c.plural}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("%s: got %q, want %q", c.p.Res.Name, got[i], want[i])
			}
		}
	}
	if hotel.ParentPattern() != "" || plan.ParentPattern() != "hotels/{hotel}" {
		t.Errorf("ParentPattern: %q, %q", hotel.ParentPattern(), plan.ParentPattern())
	}
	if plan.Type() != "hotel.opentravel.org/RatePlan" {
		t.Errorf("Type = %q", plan.Type())
	}
}

func TestCheckUniqueCatchesSharedType(t *testing.T) {
	// Different packages under different parents, so only the type -- flat
	// across the API (AIP-123) -- is shared.
	hotel := &Package{Res: &Resource{Spec: "hotel", Name: "Hotel"}}
	a := &Package{Res: &Resource{Spec: "hotel", Domain: "revenue", Name: "Rate"}, Parent: hotel}
	b := &Package{Res: &Resource{Spec: "hotel", Domain: "booking", Name: "Rate"}}
	err := checkUnique(&Model{Packages: []*Package{a, b}})
	if err == nil || !strings.Contains(err.Error(), "type hotel.opentravel.org/Rate") {
		t.Errorf("checkUnique = %v, want the shared resource type reported", err)
	}
}

// TestTableIsCoherent checks the resources table itself: every parent
// exists, and no two resources share a package, pattern, type or a service
// name with their own message.
func TestTableIsCoherent(t *testing.T) {
	byName := map[string]*Package{}
	m := &Model{}
	for _, r := range resources {
		p := &Package{Res: r}
		if _, dup := byName[r.Name]; dup {
			t.Errorf("resource %s listed twice", r.Name)
		}
		byName[r.Name] = p
		m.Packages = append(m.Packages, p)
	}
	for _, p := range m.Packages {
		if p.Res.Parent != "" {
			if p.Parent = byName[p.Res.Parent]; p.Parent == nil {
				t.Errorf("%s: no parent %s", p.Res.Name, p.Res.Parent)
			}
		}
		if p.Service() == p.Res.Name {
			t.Errorf("%s: service and message share a name", p.Res.Name)
		}
		if p.Res.Spec != "hotel" && p.Res.Spec != "common" {
			t.Errorf("%s: spec %q", p.Res.Name, p.Res.Spec)
		}
	}
	if err := checkUnique(m); err != nil {
		t.Error(err)
	}
}

func TestCarved(t *testing.T) {
	p := &Package{Res: &Resource{Carve: []string{"HotelRef", "HotelReferenceGroup/*"}}}
	for _, c := range []struct {
		name, key string
		want      bool
	}{
		{"HotelRef", "HotelRatePlanType/HotelRef", true},
		{"@HotelCode", "HotelReferenceGroup/@HotelCode", true}, // by group
		{"@RatePlanCode", "RatePlanGroup/@RatePlanCode", false},
		{"@HotelCode", "HotelReferenceGroupX/@HotelCode", false}, // prefix is a whole segment
	} {
		if got := p.carved(c.name, c.key); got != c.want {
			t.Errorf("carved(%s, %s) = %v, want %v", c.name, c.key, got, c.want)
		}
	}
}
