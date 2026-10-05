// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/codelist"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/ledger"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// enumType is a closed XSD enumeration: a top-level enum for a named type,
// one nested in the declaring message for an anonymous one.
func (s *scope) enumType(t *xsd.Simple, attrName, key string) ftype {
	e, ok := s.b.enums[t]
	if !ok {
		e = &Enum{Doc: t.Doc, Origin: t.Name}
		s.b.enums[t] = e
		if t.Name != "" {
			e.Name = typeName(t.Name)
			if s.b.claim(e.Name, "enum "+t.Name) {
				e.Sub = SubEnums
				e.File = path(e.Sub, fileFor(e.Name))
				s.b.Enums = append(s.b.Enums, e)
			}
		} else {
			e.Name, e.Parent, e.Origin = naming.Pascal(attrName), s.m, key
			if r, ok := catalog.TypeRenames[key]; ok {
				e.Name = r
			}
			s.m.Enums = append(s.m.Enums, e)
		}
		if strings.HasSuffix(e.Name, "Status") {
			s.b.errorf("%s: enum %s ends in Status, which AIP-216 reserves: add a TypeRenames entry", key, e.Name)
		}
		if e.Doc == "" {
			e.Doc = naming.Humanise(e.Name) + "."
		}
		for i, v := range t.Enums {
			e.Values = append(e.Values, &EnumValue{Number: i + 1, Doc: v.Doc, Spelling: v.Value})
		}
		s.b.nameValues(e, key)
	}
	return ftype{Enum: e, Kind: kindEnum}
}

// code is an OTA_CodeType value. When its documentation cites exactly one
// code table the workbook has, it is that table's enum; otherwise OTA did
// not say what the code means, and it stays a validated string.
func (s *scope) code(t *xsd.Simple, key, doc string) ftype {
	cited := codelist.Cited(doc)
	if a, ok := catalog.CodeListAliases[key]; ok {
		cited = []string{a}
		if a == catalog.NoTable {
			cited = nil
		}
	}
	if len(cited) == 1 && s.b.Codes.Tables[cited[0]] != nil {
		e := s.b.table(cited[0], key)
		return ftype{Enum: e, Kind: kindEnum, CodeList: cited[0]}
	}
	ft := s.stringType(t, key)
	switch len(cited) {
	case 0:
		ft.Note = "An OpenTravel code; OTA does not say which code table it is from."
	case 1:
		s.b.errorf("%s: cites code table %s, which the workbook does not have", key, cited[0])
	default:
		ft.Note = "An OpenTravel code from one of the tables " + strings.Join(cited, ", ") + "."
	}
	return ft
}

// table returns the enum for a code table, planning it on first use. Each
// value's number is its OTA code, so the wire carries exactly what OTA's
// XML does (docs/decisions.md).
func (b *Builder) table(abbr, key string) *Enum {
	if e, ok := b.tables[abbr]; ok {
		return e
	}
	t := b.Codes.Tables[abbr]
	name := naming.Pascal(tableName(t.Name))
	if strings.HasSuffix(name, "Status") {
		// AIP-216 reserves *Status for lifecycle enums; a code table is a
		// list of codes, and says so.
		name += "Code"
	}
	e := &Enum{Name: name, CodeList: abbr, Table: t.Name,
		Doc: fmt.Sprintf("The OpenTravel code table %s (%s).", t.Name, abbr)}
	b.tables[abbr] = e
	if b.claim(e.Name, "code table "+abbr) {
		e.Sub = SubCodes
		e.File = path(e.Sub, fileFor(e.Name))
		b.Enums = append(b.Enums, e)
	}
	seen := map[int]bool{}
	for _, c := range t.Codes {
		if n, ok := catalog.CodeRenumbers[abbr+"/"+c.Name]; ok {
			c.Value = n
		}
		ref := fmt.Sprintf("%s/%d", abbr, c.Value)
		name := c.Name
		if fix, ok := catalog.CodeFixes[ref]; ok {
			name = fix
		}
		switch {
		case seen[c.Value] && catalog.CodeFixes[ref] == "":
			b.errorf("%s: code table %s gives code %d twice: add a CodeFixes entry", key, abbr, c.Value)
			continue
		case seen[c.Value]:
			continue
		case strings.TrimSpace(name) == "":
			b.errorf("%s: code table %s has no name for code %d: add a CodeFixes entry", key, abbr, c.Value)
			continue
		}
		seen[c.Value] = true
		e.Values = append(e.Values, &EnumValue{Number: c.Value, Doc: c.Definition, Spelling: name})
	}
	b.nameValues(e, key)
	return e
}

var tableNoise = regexp.MustCompile(`\([A-Z]{3}\)|\*`)

// tableName drops the workbook's decorations: "Brand Category Code (BCC)*".
func tableName(n string) string { return strings.TrimSpace(tableNoise.ReplaceAllString(n, "")) }

// nameValues names an enum's values: the enum's own name as a prefix, then
// the OTA spelling. Two spellings that collide once made identifiers both
// take their number, by rule -- RMA 117 and 277 are both "Video games".
func (b *Builder) nameValues(e *Enum, key string) {
	prefix := naming.Screaming(e.Name) + "_"
	count := map[string]int{}
	for _, v := range e.Values {
		v.Name = prefix + naming.Screaming(v.Spelling)
		if v.Spelling == "" {
			v.Name = prefix + "EMPTY" // OTA declares the empty string as a value
		}
		count[v.Name]++
	}
	for _, v := range e.Values {
		switch {
		case v.Name == prefix:
			b.errorf("%s: %s value %q has no identifier characters", key, e.Name, v.Spelling)
		case count[v.Name] > 1 || v.Name == prefix+"UNSPECIFIED":
			v.Name = fmt.Sprintf("%s_%d", v.Name, v.Number)
		}
	}
	b.lock(e)
}

// lock passes an enum's values through codes.lock (sync/internal/ledger):
// a value emitted before keeps its number and identifier whatever OTA now
// calls it, and one OTA no longer lists stays, deprecated, so neither is
// ever reused.
func (b *Builder) lock(e *Enum) {
	if b.Ledger == nil {
		return
	}
	enum := "xsd:" + e.Origin
	switch {
	case e.CodeList != "":
		enum = "code:" + e.CodeList
	case e.Origin == "":
		enum = "gen:" + e.Name
	}
	seen := map[string]bool{}
	for _, v := range e.Values {
		k := v.Spelling
		if e.CodeList != "" {
			k = strconv.Itoa(v.Number) // a code's number is its identity
		}
		seen[k] = true
		got := b.Ledger.Resolve(enum, k, ledger.Entry{Number: v.Number, Name: v.Name, Spelling: v.Spelling})
		v.Number, v.Name = got.Number, got.Name
	}
	var gone []*EnumValue
	for k, ent := range b.Ledger.Values(enum) {
		if !seen[k] {
			gone = append(gone, &EnumValue{Number: ent.Number, Name: ent.Name, Spelling: ent.Spelling, Deprecated: true,
				Doc: "No longer in OTA's list. Kept, so its number and name are never reused."})
		}
	}
	sort.Slice(gone, func(i, j int) bool { return gone[i].Number < gone[j].Number })
	e.Values = append(e.Values, gone...)
}
