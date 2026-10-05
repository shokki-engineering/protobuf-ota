// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"sort"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// declared is every name a catalogue entry may key on, as the reader
// produces them: member keys, type names and attribute names.
type declared struct{ keys, types, attrs map[string]bool }

// survey walks every message in the suite and records what it reaches. An
// entry is only meaningful if some message reaches its declaration.
func survey(s *xsd.Schema) declared {
	d := declared{map[string]bool{}, map[string]bool{}, map[string]bool{}}
	seenC, seenS := map[*xsd.Complex]bool{}, map[*xsd.Simple]bool{}
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
			d.keys[a.Key], d.attrs[a.Name] = true, true
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

// CheckCatalogue is CLAUDE.md rule 2's check: every catalogue entry names a
// declaration the pinned suite has. A stale entry is not harmless -- it
// resolves nothing, and the decision it records silently stops applying.
func CheckCatalogue(s *xsd.Schema) []string {
	d := survey(s)
	var errs []string
	check := func(table string, keys map[string]bool, ok func(string) bool) {
		for k := range keys {
			if !ok(k) {
				errs = append(errs, fmt.Sprintf("catalog.%s: %q names nothing in the suite", table, k))
			}
		}
	}
	key := func(k string) bool { return d.keys[k] }
	keyOrType := func(k string) bool { return d.keys[k] || d.types[k] }
	check("FieldRenames", keysOf(catalog.FieldRenames), key)
	check("TypeRenames", keysOf(catalog.TypeRenames), keyOrType)
	check("ServerAssigned", catalog.ServerAssigned, key)
	check("Dropped", keysOf(catalog.Dropped), key)
	check("CodeListAliases", keysOf(catalog.CodeListAliases), key)
	check("InlineUnions", keysOf(catalog.InlineUnions), key)
	check("References", keysOf(catalog.References), func(k string) bool { return d.types[k] })
	check("Actions", keysOf(catalog.Actions), func(k string) bool { return d.attrs[k] })
	for _, r := range resources {
		check(r.Name+".Renames", keysOf(r.Renames), key)
	}
	sort.Strings(errs)
	return errs
}

func keysOf[V any](m map[string]V) map[string]bool {
	out := map[string]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}
