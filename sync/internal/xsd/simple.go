// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import (
	"fmt"
	"strconv"
	"strings"
)

// builtins are the XSD built-in types the suite uses. A built-in outside
// this list is an error rather than a guess at its protobuf mapping.
var builtins = map[string]bool{
	"string": true, "normalizedString": true, "token": true, "language": true,
	"anyURI": true, "ID": true, "IDREF": true, "NMTOKEN": true, "NMTOKENS": true,
	"boolean": true, "decimal": true, "double": true, "float": true,
	"integer": true, "int": true, "long": true, "short": true, "byte": true,
	"nonNegativeInteger": true, "positiveInteger": true,
	"unsignedInt": true, "unsignedShort": true, "unsignedByte": true, "unsignedLong": true,
	"date": true, "dateTime": true, "time": true, "duration": true,
	"gYear": true, "gYearMonth": true, "gMonth": true, "gMonthDay": true, "gDay": true,
	"base64Binary": true, "hexBinary": true, "anySimpleType": true,
}

func builtinSimple(name string) *Simple { return &Simple{Base: name} }

// simpleRef resolves a type reference to a simple type: built-in or named.
func (s *Schema) simpleRef(ref string) (*Simple, error) {
	builtin, name := qname(ref)
	if builtin {
		if !builtins[name] {
			return nil, fmt.Errorf("unsupported built-in type xs:%s", name)
		}
		return builtinSimple(name), nil
	}
	if t, ok := s.resolvedSimple[name]; ok {
		return t, nil
	}
	n, ok := s.simples[name]
	if !ok {
		return nil, fmt.Errorf("no simple type %q", name)
	}
	t, err := s.simpleNode(n, name)
	if err == nil {
		s.resolvedSimple[name] = t
	}
	return t, err
}

// simpleNode resolves a simpleType node: a restriction, a union or a list.
func (s *Schema) simpleNode(n *node, name string) (*Simple, error) {
	for _, d := range n.Kids {
		var t *Simple
		var err error
		switch d.Tag {
		case "annotation":
			continue
		case "restriction":
			t, err = s.restriction(d, name)
		case "union":
			t, err = s.union(d)
		case "list":
			t = &Simple{}
			if it := d.Attr["itemType"]; it != "" {
				t.List, err = s.simpleRef(it)
			} else if st := d.kid("simpleType"); st != nil {
				t.List, err = s.simpleNode(st, "")
			}
		default:
			err = fmt.Errorf("%s: unsupported in a simple type", d.where())
		}
		if err != nil {
			return nil, err
		}
		t.Name = name
		if t.Doc = doc(n); t.Doc == "" && t.List == nil && t.Union == nil {
			t.Doc = doc(d)
		}
		return t, nil
	}
	return nil, fmt.Errorf("%s: empty simple type", n.where())
}

// restriction derives from a base and applies the facets on d.
func (s *Schema) restriction(d *node, name string) (*Simple, error) {
	var base *Simple
	var err error
	if b := d.Attr["base"]; b != "" {
		base, err = s.simpleRef(b)
	} else if st := d.kid("simpleType"); st != nil {
		base, err = s.simpleNode(st, "")
	} else {
		err = fmt.Errorf("%s: restriction without a base", d.where())
	}
	if err != nil {
		return nil, err
	}
	return s.facets(base, d, name)
}

// union resolves memberTypes and any inline member types, in order.
func (s *Schema) union(d *node) (*Simple, error) {
	t := &Simple{}
	for _, ref := range strings.Fields(d.Attr["memberTypes"]) {
		m, err := s.simpleRef(ref)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d.where(), err)
		}
		t.Union = append(t.Union, m)
	}
	for _, st := range d.kids("simpleType") {
		m, err := s.simpleNode(st, "")
		if err != nil {
			return nil, err
		}
		t.Union = append(t.Union, m)
	}
	return t, nil
}

// facets returns a copy of base narrowed by the facets declared on d.
func (s *Schema) facets(base *Simple, d *node, name string) (*Simple, error) {
	if base == nil {
		return nil, fmt.Errorf("%s: facets on content with no simple type", d.where())
	}
	t := *base
	t.Name, t.Enums = name, append([]Enum(nil), base.Enums...)
	t.Patterns = append([][]string(nil), base.Patterns...)
	var enums []Enum
	var patterns []string
	for _, f := range d.Kids {
		v := f.Attr["value"]
		var err error
		switch f.Tag {
		case "annotation", "simpleType", "attribute", "attributeGroup", "anyAttribute":
		case "enumeration":
			enums = append(enums, Enum{Value: v, Doc: doc(f)})
		case "pattern":
			patterns = append(patterns, v)
		case "minLength":
			t.MinLen, err = intFacet(f)
		case "maxLength":
			t.MaxLen, err = intFacet(f)
		case "fractionDigits":
			t.FractionDigits, err = intFacet(f)
		case "minInclusive":
			t.MinIncl = &v
		case "maxInclusive":
			t.MaxIncl = &v
		default:
			err = fmt.Errorf("%s: unsupported facet", f.where())
		}
		if err != nil {
			return nil, err
		}
	}
	if len(enums) > 0 {
		t.Enums = enums
	}
	if len(patterns) > 0 {
		t.Patterns = append(t.Patterns, patterns)
	}
	return &t, nil
}

func intFacet(f *node) (*int, error) {
	n, err := strconv.Atoi(f.Attr["value"])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", f.where(), err)
	}
	return &n, nil
}
