// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// field adds a field of type t to the message.
func (s *scope) field(xmlName, key, doc string, notes xsd.Notes, t ftype, required bool) *Field {
	f := &Field{
		Name: s.fieldName(xmlName, key, t), Number: s.number(), Type: t.Type, TypeFile: t.File,
		Msg: t.Msg, Enum: t.Enum,
		Repeated: t.Repeated, Validate: t.Validate, Path: key, Deprecated: notes.Deprecation != "",
		Ref: t.Ref,
	}
	switch {
	case catalog.ServerAssigned[key] || isState(t):
		f.Behavior = []string{"OUTPUT_ONLY"}
	case s.b.Immutable[key] && s.m.Resource != nil:
		f.Behavior = []string{"OPTIONAL", "IMMUTABLE"}
	case required:
		f.Behavior = []string{"REQUIRED"}
	default:
		f.Behavior = []string{"OPTIONAL"}
		f.Optional = t.Scalar && !t.Repeated
		if len(t.Validate) > 0 && !f.Optional && !t.Repeated && t.Kind != kindMessage {
			f.Validate = append(f.Validate, "ignore = IGNORE_IF_ZERO_VALUE")
		}
	}
	if f.Repeated {
		f.Validate = repeatedRules(f.Validate)
	}
	f.Doc = fieldDoc(xmlName, doc, notes, t, key)
	s.m.Fields = append(s.m.Fields, f)
	return f
}

// repeatedRules moves an item type's constraints under `repeated.items`,
// which is where protovalidate reads them on a repeated field, and merges
// them with any bound on the count. A bare `string = {...}` on a repeated
// field compiles and then fails every validation at run time.
func repeatedRules(rules []string) []string {
	var items, count []string
	for _, r := range rules {
		kind, body, _ := strings.Cut(r, " = ")
		switch {
		case kind == "repeated":
			count = append(count, strings.Trim(body, "{}"))
		case kind != "ignore":
			items = append(items, kind+": "+body)
		}
	}
	if len(items) > 0 {
		count = append(count, "items: {"+strings.Join(items, " ")+"}")
	}
	if len(count) == 0 {
		return nil
	}
	return []string{"repeated = {" + strings.Join(count, " ") + "}"}
}

// finish names the oneofs and checks the message's field names are unique.
func (s *scope) finish() {
	for _, alts := range s.oneof {
		if len(alts) == 1 {
			continue // its alternatives carved away, a choice of one is a field
		}
		var names []string
		for _, f := range alts {
			names = append(names, f.Name)
		}
		name := names[0] + "_choice"
		if len(names) <= 3 {
			name = join(names, "_or_")
		}
		for _, f := range alts {
			f.OneOf = name
		}
	}
	seen := map[string]string{}
	for _, f := range s.m.Fields {
		if prev, dup := seen[f.Name]; dup {
			s.b.errorf("%s: %s and %s both become field %q: add a FieldRenames entry", s.m.FullName(), prev, f.Path, f.Name)
		}
		seen[f.Name] = f.Path
	}
}

// isState reports whether a field holds a lifecycle state: an enum named
// *State. AIP-216 makes it OUTPUT_ONLY, and CLAUDE.md rule 8 says why --
// a state is what a method produces, not what a client writes.
func isState(t ftype) bool { return t.Enum != nil && strings.HasSuffix(t.Enum.Name, "State") }
