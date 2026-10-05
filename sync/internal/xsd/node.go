// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// node is one XSD declaration as written, before any reference is resolved.
//
// The reader keeps XML at arm's length: a file becomes a tree of nodes keyed
// by local name, and every later stage works on that tree. Nothing outside
// this file sees encoding/xml.
type node struct {
	Tag   string // local name, e.g. "complexType"
	Attr  map[string]string
	Kids  []*node
	Text  string // character data, for xs:documentation
	File  string // base name of the file it came from, for error messages
	Line  int
	Lang  string // xml:lang, for xs:documentation
	Owner *node  // the enclosing node

	// Foreign marks markup OTA embeds inside xs:documentation, which XSD
	// allows there and nowhere else: <LegacyDefaultValue> and
	// <DeprecationWarning>. Its Tag is the element's local name.
	Foreign bool
}

// xsNamespace is the only namespace a declaration may live in. Anything else
// in the tree -- a vendor extension, an embedded schema -- is refused.
const xsNamespace = "http://www.w3.org/2001/XMLSchema"

// parseFile reads one .xsd into a node tree.
func parseFile(path string) (*node, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	base := filepath.Base(path)
	dec := xml.NewDecoder(f)
	var root *node
	var stack []*node
	for {
		off := dec.InputOffset()
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			inDoc := len(stack) > 0 && (stack[len(stack)-1].Tag == "documentation" || stack[len(stack)-1].Foreign)
			n, err := newNode(t, base, off, inDoc)
			if err != nil {
				return nil, err
			}
			if len(stack) == 0 {
				root = n
			} else {
				parent := stack[len(stack)-1]
				n.Owner = parent
				parent.Kids = append(parent.Kids, n)
			}
			stack = append(stack, n)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			}
		}
	}
	if root == nil || root.Tag != "schema" {
		return nil, fmt.Errorf("%s: not an XML Schema", base)
	}
	return root, nil
}

// newNode converts a start tag, refusing anything outside the XSD namespace
// except inside documentation.
func newNode(t xml.StartElement, file string, off int64, inDoc bool) (*node, error) {
	foreign := t.Name.Space != xsNamespace
	if foreign && !inDoc {
		return nil, fmt.Errorf("%s@%d: element %s:%s is not XML Schema", file, off, t.Name.Space, t.Name.Local)
	}
	n := &node{Tag: t.Name.Local, Attr: map[string]string{}, File: file, Line: int(off), Foreign: foreign}
	for _, a := range t.Attr {
		switch {
		case a.Name.Space == "http://www.w3.org/XML/1998/namespace" && a.Name.Local == "lang":
			n.Lang = a.Value
		case a.Name.Space == "" || a.Name.Space == "xmlns":
			n.Attr[a.Name.Local] = a.Value
		}
	}
	return n, nil
}

// where names a node for an error message.
func (n *node) where() string {
	if name := n.Attr["name"]; name != "" {
		return fmt.Sprintf("%s: %s %q", n.File, n.Tag, name)
	}
	return fmt.Sprintf("%s@%d: %s", n.File, n.Line, n.Tag)
}

// kids returns the children with the given tag.
func (n *node) kids(tag string) []*node {
	var out []*node
	for _, k := range n.Kids {
		if k.Tag == tag {
			out = append(out, k)
		}
	}
	return out
}

// kid returns the single child with the given tag, or nil.
func (n *node) kid(tag string) *node {
	if ks := n.kids(tag); len(ks) > 0 {
		return ks[0]
	}
	return nil
}

// qname splits a QName reference into whether it is a built-in XSD type and
// its local name. The suite declares every type in its own default
// namespace, so an unprefixed name is always one of its own.
func qname(ref string) (builtin bool, local string) {
	prefix, local, ok := strings.Cut(ref, ":")
	if !ok {
		return false, ref
	}
	return prefix == "xs", local
}
