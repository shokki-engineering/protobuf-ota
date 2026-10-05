// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import (
	"fmt"
	"strconv"
)

// occurs reads minOccurs and maxOccurs, defaulting both to 1.
func occurs(n *node) (min, max int, err error) {
	min, max = 1, 1
	if v, ok := n.Attr["minOccurs"]; ok {
		if min, err = strconv.Atoi(v); err != nil {
			return 0, 0, fmt.Errorf("%s: minOccurs %q", n.where(), v)
		}
	}
	if v, ok := n.Attr["maxOccurs"]; ok {
		if v == "unbounded" {
			max = Unbounded
		} else if max, err = strconv.Atoi(v); err != nil {
			return 0, 0, fmt.Errorf("%s: maxOccurs %q", n.where(), v)
		}
	}
	return min, max, nil
}

// times multiplies two maxOccurs values, either of which may be unbounded.
func times(a, b int) int {
	if a == Unbounded || b == Unbounded {
		return Unbounded
	}
	return a * b
}

// particles flattens a sequence or choice into c, multiplying occurrence
// through: an element inside an optional sequence is optional, and one inside
// a repeated choice repeats.
//
// A choice becomes a protobuf oneof only when that says the same thing: it
// occurs at most once, and every alternative is a single element that occurs
// at most once. Anything richer -- an alternative that is itself a sequence,
// a repeated alternative -- cannot be a oneof, and its alternatives become
// ordinary optional fields.
func (s *Schema) particles(c *Complex, n *node, key string, pmin, pmax, choice int, oneof bool) error {
	min, max, err := occurs(n)
	if err != nil {
		return err
	}
	min, max = pmin*min, times(pmax, max)
	if n.Tag == "choice" {
		s.choices++
		choice, oneof = s.choices, max == 1 && simpleChoice(n)
		min = 0 // only one alternative is present, so none is required
	}
	for _, k := range n.Kids {
		switch k.Tag {
		case "annotation":
		case "element":
			if err := s.element(c, k, key, min, max, choice, oneof); err != nil {
				return err
			}
		case "sequence", "choice":
			if err := s.particles(c, k, key, min, max, choice, false); err != nil {
				return err
			}
		case "any":
			c.Any = true
		default:
			return fmt.Errorf("%s: unsupported in a model group", k.where())
		}
	}
	return nil
}

// simpleChoice reports whether every alternative of a choice is one element
// occurring at most once.
func simpleChoice(n *node) bool {
	for _, k := range n.Kids {
		switch k.Tag {
		case "annotation":
		case "element":
			if _, max, err := occurs(k); err != nil || max != 1 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// element adds one child element to c.
func (s *Schema) element(c *Complex, n *node, key string, pmin, pmax, choice int, oneof bool) error {
	min, max, err := occurs(n)
	if err != nil {
		return err
	}
	e := &Elem{Min: pmin * min, Max: times(pmax, max), Choice: choice, OneOf: oneof}
	decl := n
	if ref := n.Attr["ref"]; ref != "" {
		top, ok := s.elements[ref]
		if !ok {
			return fmt.Errorf("%s: no top-level element %q", n.where(), ref)
		}
		e.Name, decl = ref, top
	} else {
		e.Name = n.Attr["name"]
	}
	e.Key = key + "/" + e.Name
	if e.Doc = doc(n); e.Doc == "" {
		e.Doc = doc(decl)
	}
	if e.Notes = notes(n); e.Notes == (Notes{}) {
		e.Notes = notes(decl)
	}
	if decl != n {
		// A referenced element's anonymous type belongs to that element, so
		// it is keyed and resolved once under the element's own name.
		err = s.elemType(e, decl, decl.Attr["name"])
	} else {
		err = s.elemType(e, decl, e.Key)
	}
	if err != nil {
		return err
	}
	c.Elems = append(c.Elems, e)
	return nil
}

// elemType sets an element's type from its declaration node: a type
// reference, or an anonymous complexType or simpleType child.
func (s *Schema) elemType(e *Elem, decl *node, key string) error {
	if t := decl.Attr["type"]; t != "" {
		builtin, name := qname(t)
		if _, ok := s.complexes[name]; ok && !builtin {
			c, err := s.complexNamed(name)
			e.Complex = c
			return err
		}
		st, err := s.simpleRef(t)
		e.Simple = st
		if err != nil {
			return fmt.Errorf("%s: %w", decl.where(), err)
		}
		return nil
	}
	if ct := decl.kid("complexType"); ct != nil {
		if top := decl.Owner != nil && decl.Owner.Tag == "schema"; top {
			if c, ok := s.resolvedTop[key]; ok {
				e.Complex = c
				return nil
			}
			c := &Complex{Key: key, Doc: doc(decl)}
			s.resolvedTop[key] = c
			e.Complex = c
			return s.fill(c, ct, key)
		}
		c, err := s.complexAnon(ct, key, e.Doc)
		e.Complex = c
		return err
	}
	if st := decl.kid("simpleType"); st != nil {
		t, err := s.simpleNode(st, "")
		e.Simple = t
		return err
	}
	return fmt.Errorf("%s: element has no type", decl.where())
}
