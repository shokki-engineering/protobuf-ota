// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package catalog_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

var suiteDir = filepath.Join("..", "..", "..", "modules", "opentravel", "OTA1.0", "OTA2017B", "OpenTravel_2017B_XML")

// declared is every name the catalogue may key on, as the reader produces
// them: member keys, type keys, named types, and attribute names.
type declared struct {
	keys, types, attrNames map[string]bool
}

// survey walks every OTA message in the suite and records what it reaches.
// A catalogue key is only meaningful if some message reaches its
// declaration, which is exactly what this collects.
func survey(t *testing.T) declared {
	t.Helper()
	if _, err := os.Stat(suiteDir); err != nil {
		t.Skip("modules/opentravel is not checked out")
	}
	s, err := xsd.Load(suiteDir)
	if err != nil {
		t.Fatal(err)
	}
	d := declared{map[string]bool{}, map[string]bool{}, map[string]bool{}}
	seenC := map[*xsd.Complex]bool{}
	seenS := map[*xsd.Simple]bool{}
	var simple func(*xsd.Simple)
	simple = func(x *xsd.Simple) {
		if x == nil || seenS[x] {
			return
		}
		seenS[x] = true
		if x.Name != "" {
			d.types[x.Name] = true
		}
		for _, m := range x.Union {
			simple(m)
		}
		simple(x.List)
	}
	var complex func(*xsd.Complex)
	complex = func(c *xsd.Complex) {
		if c == nil || seenC[c] {
			return
		}
		seenC[c] = true
		d.keys[c.Key] = true
		if c.Name != "" {
			d.types[c.Name] = true
		}
		simple(c.Text)
		for _, a := range c.Attrs {
			d.keys[a.Key] = true
			d.attrNames[a.Name] = true
			simple(a.Simple)
		}
		for _, e := range c.Elems {
			d.keys[e.Key] = true
			simple(e.Simple)
			complex(e.Complex)
		}
	}
	for _, name := range s.MessageNames() {
		if e, err := s.Element(name); err == nil {
			complex(e.Complex)
		}
	}
	return d
}

// TestCatalogueKeysExist is CLAUDE.md rule 2's check that no entry names a
// declaration the pinned suite does not have: a stale entry resolves nothing
// and the decision it records silently stops applying.
func TestCatalogueKeysExist(t *testing.T) {
	d := survey(t)
	// Controls, so a survey that collected nothing cannot pass vacuously.
	if !d.keys["HotelReferenceGroup/@HotelName"] || !d.types["ProfileType"] || d.keys["NoSuchType/@Nothing"] {
		t.Fatalf("survey controls failed: %d keys, %d types", len(d.keys), len(d.types))
	}
	var missing []string
	check := func(table, key string, ok bool) {
		if !ok {
			missing = append(missing, table+": "+key)
		}
	}
	isKey := func(k string) bool { return d.keys[k] }
	isKeyOrType := func(k string) bool { return d.keys[k] || d.types[k] }

	for k := range catalog.FieldRenames {
		check("FieldRenames", k, isKey(k))
	}
	for k := range catalog.TypeRenames {
		check("TypeRenames", k, isKeyOrType(k))
	}
	for k := range catalog.ServerAssigned {
		check("ServerAssigned", k, isKey(k))
	}
	for k := range catalog.Dropped {
		check("Dropped", k, isKey(k))
	}
	for k := range catalog.CodeListAliases {
		check("CodeListAliases", k, isKey(k))
	}
	for k := range catalog.InlineUnions {
		check("InlineUnions", k, isKey(k))
	}
	for k := range catalog.References {
		check("References", k, d.types[k])
	}
	for name := range catalog.Actions {
		check("Actions", name, d.attrNames[name])
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d catalogue entries name nothing in the suite:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
}

// TestCodeFixKeysAreWellFormed checks the shape of the code-list tables'
// keys, which name workbook rows rather than XSD declarations.
func TestCodeFixKeysAreWellFormed(t *testing.T) {
	for k := range catalog.CodeFixes {
		abbr, value, ok := strings.Cut(k, "/")
		if !ok || len(abbr) != 3 || value == "" || strings.Trim(value, "0123456789") != "" {
			t.Errorf("CodeFixes key %q is not ABBR/code", k)
		}
	}
	for k, n := range catalog.CodeRenumbers {
		abbr, name, ok := strings.Cut(k, "/")
		if !ok || len(abbr) != 3 || name == "" || n <= 0 {
			t.Errorf("CodeRenumbers %q = %d is not ABBR/code name -> positive code", k, n)
		}
	}
}

// TestActionValuesAreTransactionActions checks every verb listed is a value
// of TransactionActionType, the only type it applies to.
func TestActionValuesAreTransactionActions(t *testing.T) {
	if _, err := os.Stat(suiteDir); err != nil {
		t.Skip("modules/opentravel is not checked out")
	}
	s, err := xsd.Load(suiteDir)
	if err != nil {
		t.Fatal(err)
	}
	e, err := s.Find("OTA_HotelResRQ/HotelReservations/HotelReservation")
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]bool{}
	for _, a := range e.Complex.Attrs {
		if a.Name != "ResStatus" {
			continue
		}
		for _, m := range a.Simple.Union {
			for _, v := range m.Enums {
				values[v.Value] = true
			}
		}
	}
	for v := range catalog.ActionValues {
		if !values[v] {
			t.Errorf("ActionValues lists %q, which ResStatus cannot carry", v)
		}
	}
}
