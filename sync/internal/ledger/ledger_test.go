// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package ledger

import (
	"path/filepath"
	"testing"
)

// TestRenameKeepsIdentifier is the reason the ledger exists: OTA renaming a
// code must not rename the enum value, which proto3 JSON writes by name.
func TestRenameKeepsIdentifier(t *testing.T) {
	path := filepath.Join(t.TempDir(), "codes.lock")
	l, _ := Load(path)
	l.Resolve("code:RMA", "50", Entry{Number: 50, Name: "ROOM_AMENITY_TYPE_HAIRDRYER", Spelling: "Hairdryer"})
	if err := l.Save(); err != nil {
		t.Fatal(err)
	}
	l2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := l2.Resolve("code:RMA", "50", Entry{Number: 50, Name: "ROOM_AMENITY_TYPE_HAIR_DRYER", Spelling: "Hair dryer"})
	if got.Name != "ROOM_AMENITY_TYPE_HAIRDRYER" {
		t.Errorf("renamed code emitted as %s, want the locked identifier", got.Name)
	}
}

// TestInsertedSpellingTakesNextNumber: an XSD enumeration is keyed by
// spelling, so a value OTA inserts in the middle does not renumber the rest.
func TestInsertedSpellingTakesNextNumber(t *testing.T) {
	l, _ := Load(filepath.Join(t.TempDir(), "codes.lock"))
	l.Resolve("xsd:T", "Open", Entry{Number: 1, Name: "T_OPEN"})
	l.Resolve("xsd:T", "Close", Entry{Number: 2, Name: "T_CLOSE"})
	got := l.Resolve("xsd:T", "Hold", Entry{Number: 2, Name: "T_HOLD"}) // inserted second
	if got.Number != 3 {
		t.Errorf("inserted value numbered %d, want 3", got.Number)
	}
	if again := l.Resolve("xsd:T", "Close", Entry{Number: 3, Name: "T_CLOSE"}); again.Number != 2 {
		t.Errorf("existing value moved to %d", again.Number)
	}
}

// TestNameCollisionTakesNumber: a new value whose identifier another value
// already holds takes its number as a suffix.
func TestNameCollisionTakesNumber(t *testing.T) {
	l, _ := Load(filepath.Join(t.TempDir(), "codes.lock"))
	l.Resolve("code:X", "1", Entry{Number: 1, Name: "X_A"})
	if got := l.Resolve("code:X", "7", Entry{Number: 7, Name: "X_A"}); got.Name != "X_A_7" {
		t.Errorf("colliding identifier = %s, want X_A_7", got.Name)
	}
}
