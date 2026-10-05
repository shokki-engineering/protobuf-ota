// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import (
	"fmt"
	"regexp"
	"strings"
)

// attribute adds an attribute, or every attribute of a referenced group.
//
// A name already present is replaced rather than duplicated: that is how a
// complexContent restriction restates an inherited attribute, and
// use="prohibited" is how it removes one.
func (s *Schema) attribute(c *Complex, n *node, key string) error {
	var add []*Attr
	if n.Tag == "attributeGroup" {
		g, err := s.attrGroup(n.Attr["ref"])
		if err != nil {
			return fmt.Errorf("%s: %w", n.where(), err)
		}
		add = g
	} else {
		a, err := s.attr(n, key)
		if err != nil {
			return err
		}
		if a == nil {
			removeAttr(c, n.Attr["name"])
			return nil
		}
		add = []*Attr{a}
	}
	for _, a := range add {
		removeAttr(c, a.Name)
		c.Attrs = append(c.Attrs, a)
	}
	return nil
}

func removeAttr(c *Complex, name string) {
	out := c.Attrs[:0]
	for _, a := range c.Attrs {
		if a.Name != name {
			out = append(out, a)
		}
	}
	c.Attrs = out
}

// attr resolves one xs:attribute declared at key. A prohibited attribute
// resolves to nil.
func (s *Schema) attr(n *node, key string) (*Attr, error) {
	for _, k := range []string{"ref", "default", "fixed", "form"} {
		if _, ok := n.Attr[k]; ok {
			return nil, fmt.Errorf("%s: attribute %s= is not supported", n.where(), k)
		}
	}
	name := n.Attr["name"]
	a := &Attr{Name: name, Key: key + "/@" + name, Doc: doc(n), Notes: notes(n)}
	switch n.Attr["use"] {
	case "required":
		a.Required = true
	case "prohibited":
		return nil, nil
	case "", "optional":
	default:
		return nil, fmt.Errorf("%s: use=%q", n.where(), n.Attr["use"])
	}
	var err error
	switch t, st := n.Attr["type"], n.kid("simpleType"); {
	case t != "":
		a.Simple, err = s.simpleRef(t)
	case st != nil:
		a.Simple, err = s.simpleNode(st, "")
	default:
		// No type is xs:anySimpleType: any text. The suite never relies on
		// it for structure, and a string carries it without loss.
		a.Simple = builtinSimple("string")
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", n.where(), err)
	}
	return a, nil
}

// attrGroup resolves a named attribute group, once. Its attributes are keyed
// under the group's own name, so a catalogue entry for
// "HotelReferenceGroup/@HotelName" applies everywhere the group is used.
func (s *Schema) attrGroup(name string) ([]*Attr, error) {
	if g, ok := s.resolvedGroups[name]; ok {
		return g, nil
	}
	n, ok := s.attrGroups[name]
	if !ok {
		return nil, fmt.Errorf("no attribute group %q", name)
	}
	holder := &Complex{}
	for _, k := range n.Kids {
		switch k.Tag {
		case "annotation":
		case "attribute", "attributeGroup":
			if err := s.attribute(holder, k, name); err != nil {
				return nil, err
			}
		case "anyAttribute":
		default:
			return nil, fmt.Errorf("%s: unsupported in an attribute group", k.where())
		}
	}
	s.resolvedGroups[name] = holder.Attrs
	return holder.Attrs, nil
}

var spaces = regexp.MustCompile(`\s+`)

// doc is a declaration's English documentation, as one paragraph per
// xs:documentation. OTA writes a Description and, often, an Implementer
// note; both are kept, because the Implementer note is frequently where the
// semantics are -- DecimalPlaces is defined there.
func doc(n *node) string {
	ann := n.kid("annotation")
	if ann == nil {
		return ""
	}
	var paras []string
	for _, d := range ann.kids("documentation") {
		if d.Lang != "" && d.Lang != "en" {
			continue
		}
		if len(d.kids("LegacyDefaultValue"))+len(d.kids("DeprecationWarning")) > 0 {
			continue // markup, read by notes
		}
		t := strings.TrimSpace(spaces.ReplaceAllString(d.Text, " "))
		if t != "" {
			paras = append(paras, t)
		}
	}
	return strings.Join(paras, "\n\n")
}

// notes reads the markup OTA puts inside a declaration's documentation.
func notes(n *node) Notes {
	var out Notes
	ann := n.kid("annotation")
	if ann == nil {
		return out
	}
	for _, d := range ann.kids("documentation") {
		for _, k := range d.Kids {
			t := strings.TrimSpace(spaces.ReplaceAllString(k.Text, " "))
			switch k.Tag {
			case "LegacyDefaultValue":
				out.LegacyDefault = t
			case "DeprecationWarning":
				out.Deprecation = t
			}
		}
	}
	return out
}
