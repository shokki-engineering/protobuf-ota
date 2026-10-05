// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import (
	"os"
	"path/filepath"
	"testing"
)

// suiteDir is the pinned message suite, read from the submodule. Tests that
// need it skip when the checkout is absent rather than fail, so the package's
// own unit tests still run in a shallow clone.
var suiteDir = filepath.Join("..", "..", "..", "modules", "opentravel", "OTA1.0", "OTA2017B", "OpenTravel_2017B_XML")

func loadSuite(t *testing.T) *Schema {
	t.Helper()
	if _, err := os.Stat(suiteDir); err != nil {
		t.Skip("modules/opentravel is not checked out")
	}
	s, err := Load(suiteDir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestEveryMessageResolves is the reader's acceptance test. Every message in
// scope -- the hotel suite and the generic messages a hotel uses -- must
// resolve. Outside scope, the reader refuses exactly the constructs listed
// here and nothing else; a new refusal, or one that disappears, fails.
func TestEveryMessageResolves(t *testing.T) {
	s := loadSuite(t)
	names := s.MessageNames()
	// 299 at the pinned revision. A count that drops means a message was
	// lost in reading, which is the failure this test exists to catch.
	if len(names) != 299 {
		t.Fatalf("found %d messages, want the suite's 299", len(names))
	}
	refused := map[string]bool{
		"OTA_AirBaggageRS":  true, // FilingPricing: an element with no type
		"OTA_AirGetOfferRS": true, // CabinClassDetailType: mixed content
	}
	for _, name := range names {
		_, err := s.Element(name)
		switch {
		case err != nil && !refused[name]:
			t.Errorf("%s: %v", name, err)
		case err == nil && refused[name]:
			t.Errorf("%s: resolved, but is listed as refused", name)
		}
	}
}

func TestFind(t *testing.T) {
	s := loadSuite(t)
	e, err := s.Find("OTA_HotelRatePlanNotifRQ/RatePlans/RatePlan")
	if err != nil {
		t.Fatal(err)
	}
	if e.Complex == nil || !e.Repeated() {
		t.Fatalf("RatePlan: complex=%v repeated=%v", e.Complex != nil, e.Repeated())
	}
	var code *Attr
	for _, a := range e.Complex.Attrs {
		if a.Name == "RatePlanCode" {
			code = a
		}
	}
	if code == nil {
		t.Fatal("RatePlan has no @RatePlanCode")
	}
	if code.Key != "RatePlanGroup/@RatePlanCode" {
		t.Errorf("RatePlanCode key = %q, want it keyed by its attribute group", code.Key)
	}
}
