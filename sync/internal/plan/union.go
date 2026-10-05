// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// memberNames name a calendar union's alternatives by what each holds,
// avoiding the suffixes AIP-142 reads as a claim about the type.
var memberNames = map[string]string{
	kindDate: "day", kindTimestamp: "date_time", kindTimeOfDay: "clock",
	"gMonthDay": "month_day", "gYear": "year", "gYearMonth": "year_month", kindDecimal: "number",
}

// memberName names a union alternative: by its calendar kind, by its enum
// for an enumeration, and "text" for any other string.
func memberName(ft ftype) (string, bool) {
	if n, ok := memberNames[ft.Base]; ok && ft.Kind == kindString {
		return n, true
	}
	switch ft.Kind {
	case kindEnum:
		return naming.Snake(ft.Enum.Name), true
	case kindString:
		return "text", true
	}
	n, ok := memberNames[ft.Kind]
	return n, ok
}

// union maps an XSD union (docs/decisions.md, "An XSD union is a oneof").
//
// Members are mapped first, and members that map to the same protobuf type
// collapse: DurationType's two strings are one string, and
// MoneyOrPercentageType's two decimals -- which no reader could tell apart
// anyway -- are one Decimal. Only a union whose members stay distinct
// becomes a message with a oneof.
func (s *scope) union(t *xsd.Simple, attrName, key, doc string) ftype {
	if m, ok := s.b.unions[t]; ok {
		return ftype{Kind: kindUnion, Msg: m}
	}
	// The members are mapped in a scope of their own: an anonymous enum
	// among them belongs to the union, not to whichever message first used
	// it -- nested there, it would make the two files import each other.
	holder := &Message{Doc: t.Doc, Origin: t.Name, Name: typeName(t.Name)}
	if t.Name == "" {
		holder.Name, holder.Parent, holder.Origin = naming.Pascal(attrName)+"Value", s.m, key
		if r, ok := catalog.TypeRenames[key]; ok {
			holder.Name = r
		}
	}
	inner := &scope{b: s.b, m: holder, c: s.c, oneof: map[int][]*Field{}}
	var members []ftype
	distinct := map[string]bool{}
	for _, mt := range t.Union {
		ft := inner.simple(mt, attrName, key, doc)
		if !distinct[ft.name()] {
			distinct[ft.name()] = true
			members = append(members, ft)
		}
	}
	if len(members) == 1 {
		ft := members[0]
		if len(t.Union) > 1 {
			ft.Validate = nil // each member's rule is too narrow for the union
		}
		for _, en := range holder.Enums {
			en.Parent = s.m
			s.m.Enums = append(s.m.Enums, en)
		}
		return ft
	}
	m := holder
	s.b.unions[t] = m
	if t.Name != "" {
		s.b.Top(m)
	} else {
		s.m.Nested = append(s.m.Nested, m)
	}
	if m.Doc == "" {
		m.Doc = naming.Humanise(m.Name) + "."
	}
	for i, ft := range members {
		name, ok := memberName(ft)
		if !ok {
			s.b.errorf("%s: no oneof member name for a %s alternative", key, ft.name())
			name = ft.Kind
		}
		behavior := []string{"OPTIONAL"}
		if isState(ft) {
			behavior = []string{"OUTPUT_ONLY"}
		}
		m.Fields = append(m.Fields, &Field{
			Name: name, Number: i + 1, Type: ft.Type, TypeFile: ft.File, Msg: ft.Msg, Enum: ft.Enum, OneOf: "value",
			Behavior: behavior, Validate: ft.Validate,
			Doc: "The value, when it is " + humanKind(ft) + ".",
		})
	}
	return ftype{Kind: kindUnion, Msg: m}
}

func humanKind(ft ftype) string {
	switch ft.Kind {
	case kindDate:
		return "a calendar date"
	case kindTimestamp:
		return "a date and time"
	case kindTimeOfDay:
		return "a time of day"
	case kindEnum:
		return "one of the listed values"
	case kindDecimal:
		return "a number"
	}
	if n, ok := memberNames[ft.Base]; ok {
		return "a " + strings.ReplaceAll(n, "_", " ")
	}
	return "any other text"
}

// moneyAttrs finds the attributes of a type that are amounts in the type's
// own currency: those of OTA's Money type, in a type that also carries a
// CurrencyCode. Each becomes a Money field, and CurrencyCode and
// DecimalPlaces are consumed into them (docs/conventions.md rule 13).
func moneyAttrs(c *xsd.Complex) map[string]bool {
	hasCurrency := false
	var amounts []string
	for _, a := range c.Attrs {
		switch {
		case a.Name == "CurrencyCode":
			hasCurrency = true
		case a.Simple != nil && a.Simple.Name == "Money":
			amounts = append(amounts, a.Name)
		}
	}
	out := map[string]bool{}
	if hasCurrency {
		for _, n := range amounts {
			out[n] = true
		}
	}
	return out
}
