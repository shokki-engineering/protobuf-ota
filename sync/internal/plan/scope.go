// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"fmt"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// scope plans the members of one complex type into one message.
type scope struct {
	b     *Builder
	m     *Message
	c     *xsd.Complex
	money map[string]bool // Money attributes sharing the type's CurrencyCode
	next  int
	oneof map[int][]*Field // choice number -> its alternatives
}

// number hands out the next field number, after any the message already
// has and after its reserved range.
func (s *scope) number() int {
	if s.next == 0 {
		s.next = s.m.Reserved[1] + 1
		for _, f := range s.m.Fields {
			if f.Number >= s.next {
				s.next = f.Number + 1
			}
		}
	}
	s.next++
	return s.next - 1
}

// attr plans one attribute.
func (s *scope) attr(a *xsd.Attr) {
	if _, ok := catalog.Actions[a.Name]; ok {
		return
	}
	if _, ok := catalog.Dropped[a.Key]; ok {
		return
	}
	if s.weekdayField(a) {
		return
	}
	if iu, ok := catalog.InlineUnions[a.Key]; ok && a.Simple.Union != nil {
		s.inlineUnion(a, iu)
		return
	}
	if len(s.money) > 0 && (a.Name == "CurrencyCode" || a.Name == "DecimalPlaces") {
		return // consumed into each Money field (conventions rule 13)
	}
	var t ftype
	if s.money[a.Name] {
		t = ftype{Msg: s.b.value("Money"), Kind: kindMessage}
	} else {
		t = s.simple(a.Simple, a.Name, a.Key, a.Doc)
	}
	f := s.field(a.Name, a.Key, a.Doc, a.Notes, t, a.Required)
	f.Kind = "KIND_ATTRIBUTE"
}

// text plans the character content of a simple-content type.
func (s *scope) text(t *xsd.Simple) {
	name := "Value"
	for _, a := range s.c.Attrs {
		if a.Name == "Value" {
			name = "Text" // the element also has a Value attribute
		}
	}
	ft := s.simple(t, name, s.c.Key+"/#text", "")
	f := s.field(name, s.c.Key+"/#text", "The element's text.", xsd.Notes{}, ft, false)
	f.Kind, f.Path = "KIND_TEXT", s.c.Key
}

// elem plans one child element.
func (s *scope) elem(e *xsd.Elem) {
	if e.Name == "TPA_Extensions" || (e.Complex != nil && e.Complex.Any && len(e.Complex.Elems) == 0 && len(e.Complex.Attrs) == 0) {
		return // decisions.md: TPA_Extensions is dropped
	}
	if _, ok := catalog.Dropped[e.Key]; ok {
		return
	}
	var t ftype
	repeated, max, path, required := e.Repeated(), e.Max, e.Key, e.Min >= 1
	switch {
	case e.Simple != nil:
		t = s.simple(e.Simple, e.Name, e.Key, e.Doc)
	case !e.OneOf && !repeated && wrapped(e.Complex) != nil:
		// A wrapper -- <Rates><Rate/><Rate/></Rates> -- is the repeated
		// child itself. The path keeps both names, so a codec can put the
		// wrapper back.
		k := wrapped(e.Complex)
		if e.Complex.Name != "" && k.Complex != nil && k.Complex.Name == "" {
			// The child's anonymous type belongs to a named wrapper that is
			// folded away, so no message is its scope: it is top-level.
			name := k.Name
			if _, taken := s.b.top[naming.Pascal(name)]; taken {
				name = typeName(e.Complex.Name) + "Item"
			}
			s.b.complexMessage(k.Complex, name, nil)
		}
		t, repeated, max, path = s.elemType(k), true, k.Max, e.Key+"/"+k.Name
		required = required && k.Min >= 1
	default:
		t = s.elemType(e)
	}
	if t.Repeated {
		repeated = true
	}
	t.Repeated = repeated
	if repeated && max > 1 {
		t.Validate = append(t.Validate, fmt.Sprintf("repeated = {max_items: %d}", max))
	}
	f := s.field(e.Name, e.Key, e.Doc, e.Notes, t, required && !e.OneOf)
	f.Kind, f.Path = "KIND_ELEMENT", path
	if e.OneOf {
		s.oneof[e.Choice] = append(s.oneof[e.Choice], f)
	}
}

// elemType is the type of an element with complex content: a message, or
// the scalar a text-only element carries.
func (s *scope) elemType(e *xsd.Elem) ftype {
	if e.Simple != nil {
		return s.simple(e.Simple, e.Name, e.Key, e.Doc)
	}
	c := e.Complex
	if ref, ok := catalog.References[c.Name]; ok {
		return ftype{Type: "string", Kind: kindString, Ref: ref,
			Note: "The name of the " + ref + " resource; OTA embeds a copy here."}
	}
	if c.Text != nil && len(c.Attrs) == 0 && len(c.Elems) == 0 {
		return s.simple(c.Text, e.Name, e.Key, e.Doc)
	}
	return ftype{Msg: s.b.complexMessage(c, e.Name, s.m), Kind: kindMessage}
}

// wrapped returns the one repeated child of a wrapper element, or nil when
// the element is not a pure wrapper: no attributes, no text, and exactly one
// child element once TPA_Extensions is set aside.
func wrapped(c *xsd.Complex) *xsd.Elem {
	if c == nil || len(c.Attrs) > 0 || c.Text != nil {
		return nil
	}
	var only *xsd.Elem
	for _, k := range c.Elems {
		if k.Name == "TPA_Extensions" {
			continue
		}
		if only != nil {
			return nil
		}
		only = k
	}
	if only == nil || !only.Repeated() {
		return nil
	}
	return only
}
