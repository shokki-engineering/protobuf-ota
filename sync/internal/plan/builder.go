// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/codelist"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/ledger"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// Builder plans one package. Everything a package's resource reaches is
// planned into the package itself -- AIP-215 forbids a field naming a
// message in another package -- so each package has its own Builder and its
// own copy of every value type it uses (CLAUDE.md rule 4).
type Builder struct {
	Codes  *codelist.List
	Ledger *ledger.Ledger // codes.lock; nil leaves enum values as planned

	Messages []*Message // top-level, in the order first reached
	Enums    []*Enum    // top-level

	// ResourceRenames are field renames applied on the resource message
	// only, keyed by OTA path. Immutable marks OTA paths whose field may be
	// set once: the round-trip identifier ota_uid.
	ResourceRenames map[string]string
	Immutable       map[string]bool

	top      map[string]string // top-level name -> what claimed it
	topMsgs  map[string]*Message
	complexs map[*xsd.Complex]*Message
	enums    map[*xsd.Simple]*Enum
	unions   map[*xsd.Simple]*Message
	tables   map[string]*Enum
	values   map[string]*Message
	genEnums map[string]*Enum

	errs []string
}

// New returns a Builder over the code list.
func New(codes *codelist.List) *Builder {
	return &Builder{
		Codes: codes, top: map[string]string{}, topMsgs: map[string]*Message{},
		complexs: map[*xsd.Complex]*Message{}, enums: map[*xsd.Simple]*Enum{},
		unions: map[*xsd.Simple]*Message{}, tables: map[string]*Enum{},
		values: map[string]*Message{}, genEnums: map[string]*Enum{}, ResourceRenames: map[string]string{},
		Immutable: map[string]bool{},
	}
}

// Errors are the problems the plan found: each a rule-15 refusal naming the
// declaration and what to change. A package with any is not written.
func (b *Builder) Errors() []string {
	out := append([]string(nil), b.errs...)
	sort.Strings(out)
	return out
}

func (b *Builder) errorf(format string, args ...any) {
	b.errs = append(b.errs, fmt.Sprintf(format, args...))
}

// claim registers a top-level name, reporting a collision.
func (b *Builder) claim(name, by string) bool {
	if prev, taken := b.top[name]; taken {
		b.errorf("%s and %s both become top-level %q: add a TypeRenames entry", prev, by, name)
		return false
	}
	b.top[name] = by
	return true
}

// Top adds a top-level message, such as the resource message.
//
// A message from an anonymous type yields its name to a named type's,
// whichever is reached first: Address the folded wrapper's child becomes
// AddressItem when OTA's AddressType arrives. References are resolved when
// written, so renaming one already in use is safe.
func (b *Builder) Top(m *Message) {
	if prev, ok := b.topMsgs[m.Name]; ok && prev.anonymous && !m.anonymous {
		delete(b.top, prev.Name)
		delete(b.topMsgs, prev.Name)
		suffix := "Element"
		if strings.Contains(prev.Origin, "/") {
			suffix = "Item"
		}
		prev.Name += suffix
		prev.File = path(prev.Sub, fileFor(prev.Name))
		b.claim(prev.Name, "message "+prev.Name)
		b.topMsgs[prev.Name] = prev
	}
	if b.claim(m.Name, "message "+m.Name) {
		b.topMsgs[m.Name] = m
		m.Sub = SubTypes
		m.File = path(m.Sub, fileFor(m.Name))
		b.Messages = append(b.Messages, m)
	}
}

// typeName is the message or enum name for a named XSD type.
func typeName(xsdName string) string {
	if r, ok := catalog.TypeRenames[xsdName]; ok {
		return r
	}
	return naming.TypeName(xsdName)
}

// complexMessage returns the message for a complex type, planning it on
// first use. A named type is a top-level message; an anonymous one is
// nested in the message whose element declares it, which is where XSD
// scopes it too.
func (b *Builder) complexMessage(c *xsd.Complex, elemName string, parent *Message) *Message {
	if m, ok := b.complexs[c]; ok {
		return m
	}
	m := &Message{Origin: c.Key, Doc: c.Doc, anonymous: c.Name == ""}
	b.complexs[c] = m
	// An anonymous type declared on a top-level element is shared by every
	// type that references the element, so it cannot nest in whichever one
	// reached it first: the others would import that file, and a cycle
	// follows. Its key is the element name alone.
	shared := c.Name == "" && !strings.Contains(c.Key, "/")
	if c.Name != "" || parent == nil || shared {
		m.Name = typeName(c.Name)
		if c.Name == "" {
			m.Name = naming.Pascal(elemName)
		}
		b.Top(m)
	} else {
		m.Name, m.Parent = naming.Pascal(elemName), parent
		if r, ok := catalog.TypeRenames[c.Key]; ok {
			m.Name = r
		}
		parent.Nested = append(parent.Nested, m)
	}
	if m.Doc == "" {
		m.Doc = naming.Humanise(m.Name) + "."
	}
	b.Fill(m, c, nil)
	return m
}

// Fill plans c's members into m, skipping any whose key skip reports true.
// Field numbers continue from the highest m already has.
func (b *Builder) Fill(m *Message, c *xsd.Complex, skip func(key string) bool) {
	ctx := &scope{b: b, m: m, c: c, money: moneyAttrs(c), oneof: map[int][]*Field{}}
	for _, a := range c.Attrs {
		if skip != nil && skip(a.Key) {
			continue
		}
		ctx.attr(a)
	}
	if c.Text != nil {
		ctx.text(c.Text)
	}
	for _, e := range c.Elems {
		if skip != nil && skip(e.Key) {
			continue
		}
		ctx.elem(e)
	}
	ctx.finish()
}

// fileFor is the file a top-level declaration is written to. service.proto
// and messages.proto belong to the package's methods, so an OTA type named
// Service or Messages takes a suffix rather than overwrite them.
func fileFor(name string) string {
	base := naming.Snake(name)
	if base == "service" || base == "messages" {
		base += "_type"
	}
	return base + ".proto"
}

// path joins a subpackage and a base name into a file path from v1/.
func path(sub, base string) string {
	if sub == "" {
		return base
	}
	return sub + "/" + base
}

// Root moves a top-level message into the package's v1/ itself, as the
// resource and its requests are. file names its file there; empty keeps
// the base name it has.
func (b *Builder) Root(m *Message, file string) {
	if file == "" {
		file = m.File[strings.LastIndex(m.File, "/")+1:]
	}
	m.Sub, m.File = "", file
}
