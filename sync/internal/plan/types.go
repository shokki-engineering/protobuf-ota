// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package plan decides what each OTA type, attribute and element becomes in
// protobuf: its name, its type, its field behavior, its validation and its
// annotation. It produces a tree of declarations and nothing else; writing
// them out is emit's job, and deciding which packages exist is model's.
package plan

import "strings"

// Message is one protobuf message: a resource, an OTA complex type, an
// anonymous element type nested in its parent, or a generated value type.
type Message struct {
	Name string
	Doc  string

	Fields []*Field
	Nested []*Message
	Enums  []*Enum

	// Origin is where the message comes from in OTA: the XSD type, or the
	// path of the element that declares an anonymous one. Empty for a
	// generated value type, which has no OTA counterpart.
	Origin string

	// Reserved is a range of field numbers held back, [from, to]; zero when
	// none is.
	Reserved [2]int

	// Resource is set on a resource message.
	Resource *Resource

	Parent *Message // the enclosing message of a nested one
	File   string   // the file a top-level message is written to, from v1/

	// Sub is the subpackage under v1 a top-level message is declared in:
	// "" for the resource and its requests, SubTypes for everything else.
	Sub string

	// anonymous marks a top-level message made from an anonymous XSD type,
	// whose name is the generator's choice and so yields to a named type's.
	anonymous bool
}

// Target returns the field's type as written in the file, and the base name
// of the local file declaring it, or the import path of a well-known type.
func (f *Field) Target() (typ, file string) {
	switch {
	case f.Msg != nil:
		return f.Msg.FullName(), f.Msg.Top().File
	case f.Enum != nil:
		if f.Enum.Parent != nil {
			return f.Enum.FullName(), f.Enum.Parent.Top().File
		}
		return f.Enum.FullName(), f.Enum.File
	}
	return f.Type, f.TypeFile
}

// FullName is the name a field elsewhere in the package uses to refer to
// the message: `RatePlan.Rate` for a nested one.
func (m *Message) FullName() string {
	if m.Parent == nil {
		return m.Name
	}
	return m.Parent.FullName() + "." + m.Name
}

// Top is the top-level message a message is, or is nested in.
func (m *Message) Top() *Message {
	for m.Parent != nil {
		m = m.Parent
	}
	return m
}

// Field is one field.
type Field struct {
	Name   string
	Number int
	Doc    string

	// Type is a scalar keyword or a well-known type, with the file to import
	// for it. A message or enum of the package is Msg or Enum instead, and
	// its name and file are read when the field is written: both can change
	// after the field is planned -- a union is named after its members, and
	// a cycle of types is regrouped into one file.
	Type     string
	TypeFile string
	Msg      *Message
	Enum     *Enum

	Repeated bool
	Optional bool   // proto3 `optional`, for presence on a scalar
	OneOf    string // the oneof the field belongs to, if any

	Behavior []string // google.api.field_behavior values, in order
	Validate []string // buf.validate option bodies, one per option
	Ref      string   // google.api.resource_reference type
	ChildRef string   // google.api.resource_reference child_type
	UUID     bool     // google.api.field_info format UUID4

	Kind string // OTA node kind: KIND_ATTRIBUTE, KIND_ELEMENT, KIND_TEXT
	Path string // OTA path of the node the field carries

	Deprecated bool
}

// Enum is one protobuf enum: a closed XSD enumeration or an OTA code table.
type Enum struct {
	Name   string
	Doc    string
	Values []*EnumValue

	Origin   string // the XSD simple type, or the attribute declaring it
	CodeList string // the code table abbreviation, for a code-table enum
	Table    string // the code table's name

	Parent *Message // the enclosing message of a nested enum
	File   string   // from v1/, for a top-level enum
	Sub    string   // SubCodes or SubEnums, for a top-level enum
}

// The subpackages under v1. Each is its own proto package and directory --
// protobuf.opentravel.hotel.revenue.rate_plan.v1.types -- so buf's
// PACKAGE_DIRECTORY_MATCH holds; api-linter reads every package under v1 as
// v1 itself, so AIP-215 holds too (docs/decisions-generator.md).
const (
	SubTypes = "types" // OTA value messages and the generated value types
	SubCodes = "codes" // OpenTravel code-table enums
	SubEnums = "enums" // closed XSD enumerations
)

// SubOf is the subpackage a file path from v1/ is in.
func SubOf(file string) string {
	if i := strings.Index(file, "/"); i >= 0 && !strings.HasPrefix(file, "google/") {
		return file[:i]
	}
	return ""
}

// FullName is the name a field uses to refer to the enum.
func (e *Enum) FullName() string {
	if e.Parent == nil {
		return e.Name
	}
	return e.Parent.FullName() + "." + e.Name
}

// EnumValue is one value. Spelling is the value as OTA writes it -- the
// enumeration literal, or a code table's code name.
type EnumValue struct {
	Name       string
	Number     int
	Doc        string
	Spelling   string
	Deprecated bool
}

// Resource is what makes a message a resource.
type Resource struct {
	Type      string // e.g. "hotel.opentravel.org/RatePlan"
	Pattern   string // e.g. "hotels/{hotel}/ratePlans/{rate_plan}"
	Singular  string // lowerCamel
	Plural    string // lowerCamel
	Singleton bool
}
