// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// weekdayGroup is OTA's seven weekday booleans, Mon to Sun.
const weekdayGroup = "DOW_PatternGroup/@"

// weekdays are the group's attributes in order, as OTA spells them.
var weekdays = []string{"Mon", "Tue", "Weds", "Thur", "Fri", "Sat", "Sun"}

// weekdayField folds OTA's seven weekday booleans into one repeated enum
// (docs/conventions.md rule 11): `Mon="true" Weds="true"` becomes weekdays:
// [WEEKDAY_MON, WEEKDAY_WEDS]. It reports whether a was one of them; the
// field is added at the first and the other six add nothing.
func (s *scope) weekdayField(a *xsd.Attr) bool {
	if !strings.HasPrefix(a.Key, weekdayGroup) {
		return false
	}
	for _, f := range s.m.Fields {
		if f.Path == strings.TrimSuffix(weekdayGroup, "/@") {
			return true
		}
	}
	s.m.Fields = append(s.m.Fields, &Field{
		Name: "weekdays", Number: s.number(), Enum: s.b.weekdayEnum(), Repeated: true,
		Behavior: []string{"OPTIONAL"}, Kind: "KIND_ATTRIBUTE", Path: strings.TrimSuffix(weekdayGroup, "/@"),
		Doc: "The days of the week this applies on. OTA's seven boolean attributes, Mon to\n" +
			"Sun, folded into one list: a day is listed when OTA's attribute is true. Each\n" +
			"value's source_spelling is the attribute's name. An empty list means OTA sent\n" +
			"none of them.\n\nOTA: DOW_PatternGroup.",
		Validate: []string{"repeated = {unique: true max_items: 7}"},
	})
	return true
}

// weekdayEnum is the generated enum of OTA's weekday attributes. It is not
// OTA's DayOfWeekType, whose spellings differ (Wed, Thu), so it has its own
// name.
func (b *Builder) weekdayEnum() *Enum {
	if e, ok := b.genEnums["Weekday"]; ok {
		return e
	}
	e := &Enum{Name: "Weekday", Doc: "A day of the week, as OTA's weekday attributes name it."}
	for i, d := range weekdays {
		e.Values = append(e.Values, &EnumValue{Number: i + 1, Spelling: d,
			Doc: "OTA's " + d + " attribute."})
	}
	b.nameValues(e, weekdayGroup)
	b.genEnums["Weekday"] = e
	if b.claim(e.Name, "generated enum Weekday") {
		e.Sub = SubEnums
		e.File = path(e.Sub, fileFor(e.Name))
		b.Enums = append(b.Enums, e)
	}
	return e
}
