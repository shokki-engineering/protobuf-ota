// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import "fmt"

// complexNamed resolves a named complex type, once.
func (s *Schema) complexNamed(name string) (*Complex, error) {
	if c, ok := s.resolvedComplex[name]; ok {
		return c, nil
	}
	n, ok := s.complexes[name]
	if !ok {
		return nil, fmt.Errorf("no complex type %q", name)
	}
	c := &Complex{Name: name, Key: name, Doc: doc(n)}
	s.resolvedComplex[name] = c
	return c, s.fill(c, n, name)
}

// complexAnon resolves an anonymous complexType node declared at key.
func (s *Schema) complexAnon(n *node, key, docText string) (*Complex, error) {
	c := &Complex{Key: key, Doc: docText}
	return c, s.fill(c, n, key)
}

// fill flattens a complexType node's content into c.
func (s *Schema) fill(c *Complex, ct *node, key string) error {
	c.resolving = true
	defer func() { c.resolving = false }()
	if ct.Attr["mixed"] == "true" {
		return fmt.Errorf("%s: mixed content is not supported", ct.where())
	}
	for _, k := range ct.Kids {
		var err error
		switch k.Tag {
		case "annotation":
		case "sequence", "choice":
			err = s.particles(c, k, key, 1, 1, 0, false)
		case "attribute", "attributeGroup":
			err = s.attribute(c, k, key)
		case "anyAttribute":
			c.Any = true
		case "complexContent":
			err = s.complexContent(c, k, key)
		case "simpleContent":
			err = s.simpleContent(c, k, key)
		default:
			err = fmt.Errorf("%s: unsupported in a complex type", k.where())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// base resolves the named complex type a derivation starts from.
func (s *Schema) base(n *node) (*Complex, error) {
	builtin, name := qname(n.Attr["base"])
	if builtin {
		return nil, fmt.Errorf("%s: complex content cannot derive from xs:%s", n.where(), name)
	}
	b, err := s.complexNamed(name)
	if err == nil && b.resolving {
		err = fmt.Errorf("%s: %s derives from itself", n.where(), name)
	}
	return b, err
}

// complexContent handles extension and restriction of a complex type.
//
// An extension appends to everything the base has. A restriction restates
// the elements it keeps, but inherits the attributes unless it prohibits
// them -- XSD's rule, and the reason the two cases differ here.
func (s *Schema) complexContent(c *Complex, cc *node, key string) error {
	for _, d := range cc.Kids {
		if d.Tag == "annotation" {
			continue
		}
		if builtin, name := qname(d.Attr["base"]); builtin && name == "anyType" {
			// Deriving from xs:anyType is deriving from nothing: the
			// declaration's own content is all there is.
			if err := s.fill(c, d, key); err != nil {
				return err
			}
			continue
		}
		b, err := s.base(d)
		if err != nil {
			return err
		}
		c.Attrs = append(c.Attrs, b.Attrs...)
		c.Text, c.Any = b.Text, c.Any || b.Any
		switch d.Tag {
		case "extension":
			c.Elems = append(c.Elems, b.Elems...)
		case "restriction":
		default:
			return fmt.Errorf("%s: unsupported in complexContent", d.where())
		}
		if err := s.fill(c, d, key); err != nil {
			return err
		}
	}
	return nil
}

// simpleContent handles a type whose element carries text and attributes:
// the base supplies the text's type, and the derivation adds attributes or,
// for a restriction, facets on the text.
func (s *Schema) simpleContent(c *Complex, sc *node, key string) error {
	for _, d := range sc.Kids {
		if d.Tag == "annotation" {
			continue
		}
		if d.Tag != "extension" && d.Tag != "restriction" {
			return fmt.Errorf("%s: unsupported in simpleContent", d.where())
		}
		builtin, name := qname(d.Attr["base"])
		switch _, isComplex := s.complexes[name]; {
		case !builtin && isComplex:
			b, err := s.base(d)
			if err != nil {
				return err
			}
			c.Attrs, c.Text = append(c.Attrs, b.Attrs...), b.Text
		default:
			t, err := s.simpleRef(d.Attr["base"])
			if err != nil {
				return fmt.Errorf("%s: %w", d.where(), err)
			}
			c.Text = t
		}
		if d.Tag == "restriction" {
			t, err := s.facets(c.Text, d, "")
			if err != nil {
				return err
			}
			c.Text = t
		}
		for _, k := range d.Kids {
			if k.Tag == "attribute" || k.Tag == "attributeGroup" {
				if err := s.attribute(c, k, key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
