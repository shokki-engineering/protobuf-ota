// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// ftype is what a simple type becomes in a field.
type ftype struct {
	Type     string // as written in the field
	File     string // the local file declaring it, or the well-known import
	Kind     string
	Scalar   bool // takes proto3 `optional` for presence
	Repeated bool
	Validate []string // buf.validate option bodies, `string = {...}`
	CodeList string   // the code table a code-table enum encodes
	Msg      *Message
	Enum     *Enum
	Note     string // a sentence the field's comment adds
	Ref      string // the resource type a string field names
	Base     string // the XSD built-in a string came from, for union naming
	iso3166  bool   // an ISO 3166 country code, named by AIP-143
}

const (
	kindString    = "string"
	kindBool      = "bool"
	kindInt       = "int"
	kindFloat     = "float"
	kindBytes     = "bytes"
	kindDecimal   = "decimal"
	kindDate      = "date"
	kindTimestamp = "timestamp"
	kindTimeOfDay = "time_of_day"
	kindEnum      = "enum"
	kindUnion     = "union"
	kindMessage   = "message"
)

// isoDuration is xs:duration, with OTA's N designator for nights
// (NightDurationType), which is how the two are united in DurationType.
const isoDuration = `^P(?:\d+Y)?(?:\d+M)?(?:\d+W)?(?:\d+D)?(?:\d+N)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`

// simple maps a simple type to a field type. name and key are the declaring
// attribute or element; doc is its documentation, where a code-table
// citation lives.
func (s *scope) simple(t *xsd.Simple, name, key, doc string) ftype {
	switch {
	case t.List != nil:
		ft := s.simple(t.List, name, key, doc)
		ft.Repeated, ft.Scalar = true, false
		return ft
	case t.Union != nil:
		return s.union(t, name, key, doc)
	case t.IsEnum():
		return s.enumType(t, name, key)
	case isCode(t):
		return s.code(t, key, doc)
	}
	switch t.Base {
	case "string", "normalizedString", "token", "language", "anyURI", "ID", "IDREF",
		"NMTOKEN", "NMTOKENS", "anySimpleType", "gYear", "gYearMonth", "gMonth", "gDay", "gMonthDay":
		return s.stringType(t, key)
	case "boolean":
		return ftype{Type: "bool", Kind: kindBool, Scalar: true}
	case "decimal":
		return ftype{Msg: s.b.value("Decimal"), Kind: kindDecimal}
	case "double":
		return ftype{Type: "double", Kind: kindFloat, Scalar: true}
	case "float":
		return ftype{Type: "float", Kind: kindFloat, Scalar: true}
	case "integer", "int", "short", "byte", "nonNegativeInteger", "positiveInteger",
		"unsignedInt", "unsignedShort", "unsignedByte":
		return s.intType(t, "int32")
	case "long", "unsignedLong":
		return s.intType(t, "int64")
	case "date":
		return ftype{Msg: s.b.value("Date"), Kind: kindDate}
	case "dateTime":
		return ftype{Type: "google.protobuf.Timestamp", File: "google/protobuf/timestamp.proto", Kind: kindTimestamp}
	case "time":
		return ftype{Msg: s.b.value("TimeOfDay"), Kind: kindTimeOfDay}
	case "duration":
		return ftype{Type: "string", Kind: kindString, Validate: []string{stringRule(nil, nil, isoDuration)},
			Note: "An ISO 8601 duration; OTA adds N for nights, so P3N is three nights."}
	case "base64Binary":
		return ftype{Type: "bytes", Kind: kindBytes}
	case "hexBinary":
		return ftype{Type: "string", Kind: kindString, Validate: []string{stringRule(nil, nil, "^(?:[0-9A-Fa-f]{2})*$")}}
	}
	s.b.errorf("%s: no mapping for xs:%s", key, t.Base)
	return ftype{Type: "string", Kind: kindString}
}

// stringType is a string with its length and pattern facets.
func (s *scope) stringType(t *xsd.Simple, key string) ftype {
	ft := ftype{Type: "string", Kind: kindString, Base: t.Base, iso3166: t.Name == "ISO3166"}
	var patterns []string
	for _, step := range t.Patterns {
		p, err := rePattern(step)
		if err != nil {
			s.b.errorf("%s: pattern %q: %v", key, step, err)
		}
		patterns = append(patterns, p)
	}
	pattern := ""
	if len(patterns) > 0 {
		pattern = patterns[len(patterns)-1]
	}
	if t.MinLen != nil || t.MaxLen != nil || pattern != "" {
		ft.Validate = []string{stringRule(t.MinLen, t.MaxLen, pattern)}
	}
	// A value must match every derivation step's pattern, and protovalidate
	// takes one pattern per field, so the earlier steps are CEL rules.
	for i, p := range patterns[:max(len(patterns)-1, 0)] {
		rule, err := celPattern(p, i)
		if err != nil {
			s.b.errorf("%s: %v", key, err)
		}
		ft.Validate = append(ft.Validate, rule)
	}
	return ft
}

// intType is an integer with its sign and range facets as bounds. AIP-141
// forbids unsigned types, so XSD's non-negative and unsigned types keep
// their bound here instead.
func (s *scope) intType(t *xsd.Simple, proto string) ftype {
	ft := ftype{Type: proto, Kind: kindInt, Scalar: true}
	var bounds []string
	switch {
	case t.MinIncl != nil:
		bounds = append(bounds, "gte: "+*t.MinIncl)
	case t.Base == "positiveInteger":
		bounds = append(bounds, "gt: 0")
	case t.Base == "nonNegativeInteger" || strings.HasPrefix(t.Base, "unsigned"):
		bounds = append(bounds, "gte: 0")
	}
	if t.MaxIncl != nil {
		bounds = append(bounds, "lte: "+*t.MaxIncl)
	}
	if len(bounds) > 0 {
		ft.Validate = []string{fmt.Sprintf("%s = {%s}", proto, strings.Join(bounds, " "))}
	}
	return ft
}
