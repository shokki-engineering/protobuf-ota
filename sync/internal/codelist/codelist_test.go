// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package codelist

import (
	"os"
	"path/filepath"
	"testing"
)

var workbookFile = filepath.Join("..", "..", "..", "modules", "opentravel", "CodeLists",
	"OTA_CodeLists_2024_09_04", "OpenTravel_CodeList_2024_9_4.xlsm")

// TestPinnedWorkbook checks the facts the enum design rests on, at the
// pinned revision: 152 tables, 5,293 codes, every one a positive integer.
func TestPinnedWorkbook(t *testing.T) {
	if _, err := os.Stat(workbookFile); err != nil {
		t.Skip("modules/opentravel is not checked out")
	}
	l, err := Load(workbookFile)
	if err != nil {
		t.Fatal(err)
	}
	codes := 0
	for _, tb := range l.Tables {
		codes += len(tb.Codes)
	}
	if len(l.Tables) != 152 || codes != 5293 {
		t.Errorf("tables, codes = %d, %d; want 152, 5293", len(l.Tables), codes)
	}
	rma := l.Tables["RMA"]
	if rma == nil || rma.Name != "Room Amenity Type" {
		t.Fatalf("RMA = %+v", rma)
	}
	for _, c := range rma.Codes {
		if c.Value == 50 && c.Name != "Hairdryer" {
			t.Errorf("RMA 50 = %q, want Hairdryer", c.Name)
		}
	}
}

func TestCited(t *testing.T) {
	for doc, want := range map[string]string{
		"Refer to OpenTravel Code List Room Amenity Type (RMA).":                        "RMA",
		"Refer to OpenTravel Code list Business Srvc Type (BUS).":                       "BUS",
		"Refer to OpenTravel Code list OpenTravel Code List Property Class Type (PCT).": "PCT",
		"Refer to OpenTravel Code List TDB":                                             "",
	} {
		got := Cited(doc)
		if (want == "" && len(got) != 0) || (want != "" && (len(got) != 1 || got[0] != want)) {
			t.Errorf("Cited(%q) = %v, want %q", doc, got, want)
		}
	}
}
