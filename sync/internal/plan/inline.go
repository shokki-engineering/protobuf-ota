// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// inlineUnion plans a union the catalog places directly in the message:
// its enumerations merged into one enum, minus the values that are verbs,
// and its free-text alternative beside it, both in one oneof.
//
// ResStatus is the case: a union of PMS statuses, OTA's
// TransactionActionType -- verbs and states mixed -- and the free one- or
// two-letter status codes of a GDS. The verbs are methods (CLAUDE.md rule
// 8); what remains is the reservation's lifecycle (docs/decisions.md).
func (s *scope) inlineUnion(a *xsd.Attr, iu catalog.InlineUnion) {
	e := &Enum{Name: iu.Enum, Parent: s.m, Origin: a.Key,
		Doc: "The " + iu.EnumField + " of the " + s.m.Name + ": OTA's " + a.Name +
			" values that are states, merged from the enumerations its type unites."}
	seen := map[string]bool{}
	var text *xsd.Simple
	for _, m := range a.Simple.Union {
		if !m.IsEnum() {
			if text == nil {
				text = m
			}
			continue
		}
		for _, v := range m.Enums {
			if catalog.ActionValues[v.Value] || seen[v.Value] {
				continue
			}
			seen[v.Value] = true
			e.Values = append(e.Values, &EnumValue{Number: len(e.Values) + 1, Doc: v.Doc, Spelling: v.Value})
		}
	}
	s.b.nameValues(e, a.Key)
	s.m.Enums = append(s.m.Enums, e)

	behavior := []string{"OPTIONAL"}
	if iu.OutputOnly {
		behavior = []string{"OUTPUT_ONLY"}
	}
	s.m.Fields = append(s.m.Fields, &Field{
		Name: iu.EnumField, Number: s.number(), Enum: e, OneOf: iu.OneOf, Behavior: behavior,
		Kind: "KIND_ATTRIBUTE", Path: a.Key,
		Doc: Sanitise(a.Doc + "\n\nThe states among OTA's " + a.Name + " values. The verbs among them -- Book,\n" +
			"Commit, Cancel -- are methods, and move this field (CLAUDE.md rule 8).\n\nOTA: " + a.Key + "."),
	})
	if text == nil {
		return
	}
	ft := s.stringType(text, a.Key)
	s.m.Fields = append(s.m.Fields, &Field{
		Name: iu.TextField, Number: s.number(), Type: "string", OneOf: iu.OneOf, Behavior: behavior,
		Validate: ft.Validate, Kind: "KIND_ATTRIBUTE", Path: a.Key,
		Doc: "A status code OTA carries verbatim when no state applies, such as a GDS's\n" +
			"two-letter status (HK, UC).\n\nOTA: " + a.Key + ".",
	})
}
