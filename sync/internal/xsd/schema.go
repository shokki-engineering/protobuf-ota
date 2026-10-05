// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package xsd reads the OpenTravel XML message suite.
//
// It accepts the subset of XML Schema the pinned suite actually uses and
// refuses the rest (CLAUDE.md rule 15): complexType, simpleType, sequence,
// choice, extension and restriction of both content kinds, attributeGroup,
// union, list, the facets enumeration, pattern, minLength, maxLength,
// minInclusive, maxInclusive and fractionDigits, and xs:any where OTA uses it
// for TPA_Extensions. xs:group, substitution groups, redefine, defaults and
// fixed values do not occur in the suite and are errors if they ever do.
//
// What comes out is a typed, flattened tree: a Complex lists every attribute
// and element it really has, with its base type, attribute groups and
// anonymous children already resolved, and each member carries the
// catalogue key of the declaration it came from.
package xsd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Schema is every top-level declaration in the suite, and the types resolved
// from them so far.
type Schema struct {
	complexes  map[string]*node
	simples    map[string]*node
	attrGroups map[string]*node
	elements   map[string]*node

	resolvedComplex map[string]*Complex
	resolvedSimple  map[string]*Simple
	resolvedGroups  map[string][]*Attr
	resolvedTop     map[string]*Complex
	choices         int
}

// Load reads every .xsd in dir. The suite is one namespace spread over 329
// files that include one another; reading them all and indexing by name is
// equivalent to following the includes, and cannot miss one.
func Load(dir string) (*Schema, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.xsd"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("%s: no .xsd files", dir)
	}
	sort.Strings(paths)

	s := &Schema{
		complexes: map[string]*node{}, simples: map[string]*node{},
		attrGroups: map[string]*node{}, elements: map[string]*node{},
		resolvedComplex: map[string]*Complex{}, resolvedSimple: map[string]*Simple{},
		resolvedGroups: map[string][]*Attr{}, resolvedTop: map[string]*Complex{},
	}
	for _, p := range paths {
		root, err := parseFile(p)
		if err != nil {
			return nil, err
		}
		if err := s.index(root); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// index records one file's top-level declarations.
func (s *Schema) index(root *node) error {
	for _, n := range root.Kids {
		var table map[string]*node
		switch n.Tag {
		case "complexType":
			table = s.complexes
		case "simpleType":
			table = s.simples
		case "attributeGroup":
			table = s.attrGroups
		case "element":
			table = s.elements
		case "include", "annotation":
			continue
		default:
			return fmt.Errorf("%s: unsupported top-level declaration", n.where())
		}
		name := n.Attr["name"]
		if prev, dup := table[name]; dup {
			return fmt.Errorf("%s: also declared in %s", n.where(), prev.File)
		}
		table[name] = n
	}
	return nil
}

// Element resolves a top-level element, such as OTA_HotelRatePlanNotifRQ.
func (s *Schema) Element(name string) (*Elem, error) {
	n, ok := s.elements[name]
	if !ok {
		return nil, fmt.Errorf("no top-level element %q", name)
	}
	e := &Elem{Name: name, Key: name, Doc: doc(n), Min: 1, Max: 1}
	return e, s.elemType(e, n, name)
}

// Find walks a slash-separated path of element names from a top-level
// element: "OTA_HotelRatePlanNotifRQ/RatePlans/RatePlan".
func (s *Schema) Find(path string) (*Elem, error) {
	parts := strings.Split(path, "/")
	e, err := s.Element(parts[0])
	if err != nil {
		return nil, err
	}
	for _, p := range parts[1:] {
		if e.Complex == nil {
			return nil, fmt.Errorf("%s: %s has simple content", path, e.Name)
		}
		var next *Elem
		for _, k := range e.Complex.Elems {
			if k.Name == p {
				next = k
			}
		}
		if next == nil {
			return nil, fmt.Errorf("%s: %s has no element %q", path, e.Key, p)
		}
		e = next
	}
	return e, nil
}

// MessageNames lists the top-level elements whose name marks them as an OTA
// message, OTA_*RQ and OTA_*RS, sorted.
func (s *Schema) MessageNames() []string {
	var out []string
	for name := range s.elements {
		if strings.HasPrefix(name, "OTA_") && (strings.HasSuffix(name, "RQ") || strings.HasSuffix(name, "RS")) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// HasComplex reports whether a named complex type exists.
func (s *Schema) HasComplex(name string) bool { _, ok := s.complexes[name]; return ok }

// HasAttrGroup reports whether a named attribute group exists.
func (s *Schema) HasAttrGroup(name string) bool { _, ok := s.attrGroups[name]; return ok }
