// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package xsd

// Complex is a complex type with every reference resolved: its base type's
// members, the attribute groups it pulls in and its anonymous children are
// all flattened into Attrs and Elems, in document order.
//
// A named type is resolved once and shared, so a recursive type -- one that
// contains itself, directly or not -- is a cycle of pointers, not an
// infinite tree. Walk it with a visited set.
type Complex struct {
	Name string // the XSD name; empty for an anonymous type
	Key  string // the catalogue key: Name, or the path to the declaring element
	Doc  string

	Attrs []*Attr
	Elems []*Elem

	// Text is the type of the element's character content, for a type with
	// simple content -- `<Text Formatted="true">hello</Text>`. Nil otherwise.
	Text *Simple

	// Any marks a type whose content is xs:any: a trading-partner extension
	// point with, by definition, no schema to generate from.
	Any bool

	resolving bool
}

// Elem is a child element, with its occurrence already multiplied through
// every enclosing sequence and choice.
type Elem struct {
	Name string
	Key  string // e.g. "RatePlanType/Guarantee"
	Doc  string
	Min  int
	Max  int // Unbounded for maxOccurs="unbounded"

	// Exactly one of Complex and Simple is set.
	Complex *Complex
	Simple  *Simple

	// Choice is non-zero for an element that is one alternative of an
	// xs:choice; alternatives of the same choice share the number. OneOf
	// reports whether that choice can be a protobuf oneof.
	Choice int
	OneOf  bool

	Notes
}

// Unbounded is Elem.Max for maxOccurs="unbounded".
const Unbounded = -1

// Repeated reports whether the element may occur more than once.
func (e *Elem) Repeated() bool { return e.Max == Unbounded || e.Max > 1 }

// Attr is an attribute.
type Attr struct {
	Name     string
	Key      string // e.g. "HotelReferenceGroup/@HotelCode"
	Doc      string
	Required bool
	Simple   *Simple

	Notes
}

// Notes is what OTA records about a declaration in markup inside its
// documentation rather than in the schema proper.
type Notes struct {
	// LegacyDefault is a default value OTA once declared and has since
	// removed: an absent attribute used to mean this.
	LegacyDefault string
	// Deprecation is OTA's warning that the declaration is deprecated, or
	// empty.
	Deprecation string
}

// Simple is a simple type: a built-in base and the facets every restriction
// between it and the declaration added.
type Simple struct {
	Name string // the XSD name; empty for an anonymous or built-in type
	Doc  string

	// Base is the local name of the built-in type the chain ends in:
	// "string", "decimal", "date", "boolean" ... Empty for a union or list.
	Base string

	Enums []Enum // a closed enumeration, when non-empty

	// Patterns holds one entry per derivation step that declared patterns.
	// Within a step the patterns are alternatives; across steps a value must
	// match every step. Most types have at most one step.
	Patterns [][]string

	MinLen, MaxLen   *int
	MinIncl, MaxIncl *string
	FractionDigits   *int

	Union []*Simple // the member types of a union
	List  *Simple   // the item type of a list
}

// Enum is one value of a closed enumeration.
type Enum struct {
	Value string
	Doc   string
}

// IsEnum reports whether the type is a closed enumeration.
func (s *Simple) IsEnum() bool { return len(s.Enums) > 0 }
