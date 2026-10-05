// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/catalog"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
)

// prepositions are AIP-140's, as api-linter 2.3.1 lists them.
var prepositions = set("after", "at", "before", "between", "but", "by", "except",
	"for", "from", "in", "including", "into", "of", "over", "since", "to",
	"toward", "under", "upon", "with", "within", "without")

// reserved are the words api-linter rejects as field names because a
// common language reserves them.
var reserved = set("abstract", "and", "arguments", "as", "assert", "async", "await",
	"boolean", "break", "byte", "case", "catch", "char", "class", "const", "continue",
	"debugger", "def", "default", "del", "delete", "do", "double", "elif", "else",
	"enum", "eval", "except", "export", "extends", "false", "final", "finally",
	"float", "for", "from", "function", "global", "goto", "if", "implements",
	"import", "in", "instanceof", "int", "interface", "is", "lambda", "let", "long",
	"native", "new", "nonlocal", "not", "null", "or", "package", "pass", "private",
	"protected", "public", "raise", "return", "short", "static", "strictfp", "super",
	"switch", "synchronized", "this", "throw", "throws", "transient", "true", "try",
	"typeof", "var", "void", "volatile", "while", "with", "yield")

// standardCodes are AIP-143's mandated names, keyed by what they replace.
var standardCodes = map[string]string{
	"country": "region_code", "country_code": "region_code",
	"currency": "currency_code", "language": "language_code", "lang": "language_code",
	"tz": "time_zone", "timezone": "time_zone",
	"content_type": "mime_type", "mime": "mime_type", "mimetype": "mime_type",
}

// imperative is api-linter's AIP-142 mood table for timestamp names.
var imperative = map[string]string{
	"created": "create", "creation": "create", "expired": "expire",
	"modified": "update", "updated": "update", "purged": "purge",
}

// temporal are the final words api-linter's AIP-142 rules read as a claim
// that a field is a timestamp or a date.
var temporal = set("date", "datetime", "ms", "msec", "msecs", "millis", "nanos", "ns",
	"nsec", "nsecs", "sec", "secs", "seconds", "time", "timestamp", "us", "usec", "usecs")

// fieldName is the protobuf name of an OTA attribute or element, with every
// rule-based AIP fix applied (docs/conventions.md rule 11). A name a rule
// cannot fix is reported, and must be resolved in catalog.FieldRenames.
func (s *scope) fieldName(xmlName, key string, t ftype) string {
	if r, ok := s.b.ResourceRenames[key]; ok && s.m.Resource != nil {
		return r
	}
	if r, ok := catalog.FieldRenames[key]; ok {
		return r
	}
	w := naming.SplitWords(xmlName)
	for i := range w {
		switch w[i] {
		case "url":
			w[i] = "uri" // AIP-140
		case "identifier":
			w[i] = "id" // AIP-140's abbreviation table; a bare id is AIP-148's
		}
	}
	if t.Kind == kindBool {
		if len(w) > 1 && w[0] == "is" {
			w = w[1:]
		}
		if len(w) > 1 && w[len(w)-1] == "ind" {
			w = w[:len(w)-1]
		}
	}
	w = s.temporalName(w, t, key)
	if t.Repeated && len(w) > 0 && !anyPlural(w) {
		w[len(w)-1] = naming.Pluralise(w[len(w)-1])
	}
	if t.iso3166 {
		w = regionCode(w)
	}
	name := rephrase(strings.Join(w, "_"))
	if r, ok := standardCodes[name]; ok {
		name = r
	}
	switch {
	case name == "name" && s.m.Resource == nil && s.m.Parent == nil:
		// api-linter treats a top-level message with a `name` field as a
		// resource missing its annotation.
		name = "display_name"
	case strings.HasSuffix(name, "_name") && name != "display_name" && !strings.HasSuffix(name, "_display_name"):
		name = strings.TrimSuffix(name, "_name") + "_display_name"
	}
	if reserved[name] {
		name += "_value"
	}
	if name == "id" && s.m.Resource != nil {
		s.b.errorf("%s: a resource field named id (AIP-148): add a FieldRenames entry", key)
	}
	for _, word := range strings.Split(name, "_") {
		if prepositions[word] {
			s.b.errorf("%s: field %q contains the preposition %q (AIP-140): add a FieldRenames entry", key, name, word)
		}
		if word != "" && word[0] >= '0' && word[0] <= '9' {
			s.b.errorf("%s: field %q has a word beginning with a digit (AIP-140): add a FieldRenames entry", key, name)
		}
	}
	return name
}

// temporalName applies AIP-142's suffix rules: a Timestamp ends in `_time`;
// a field of the generated Date, TimeOfDay or a calendar union may not end
// in a word that claims to be one.
func (s *scope) temporalName(w []string, t ftype, key string) []string {
	if len(w) == 0 {
		return w
	}
	last := w[len(w)-1]
	switch t.Kind {
	case kindTimestamp:
		// AIP-142 wants the imperative: update_time, not updated_time.
		for i, word := range w {
			if r, ok := imperative[word]; ok {
				w[i] = r
			}
		}
		switch last {
		case "time":
		case "date", "timestamp", "datetime":
			w[len(w)-1] = "time"
		default:
			w = append(w, "time")
		}
	case kindDate:
		if temporal[last] {
			w[len(w)-1] = "day"
		}
	case kindTimeOfDay:
		// Neither "time" (AIP-142 claims it for google.type.TimeOfDay) nor
		// "time_of_day" ("of" is an AIP-140 preposition) is open to the
		// generated type, so a wall-clock time is a `clock`.
		if temporal[last] {
			w[len(w)-1] = "clock"
		} else {
			w = append(w, "clock")
		}
	case kindUnion:
		if temporal[last] {
			if w = w[:len(w)-1]; len(w) == 0 {
				w = []string{"when"}
			}
		}
	default:
		if temporal[last] {
			s.b.errorf("%s: a %s field ending in %q reads as a timestamp (AIP-142): add a FieldRenames entry", key, t.name(), last)
		}
	}
	return w
}

func set(words ...string) map[string]bool {
	m := map[string]bool{}
	for _, w := range words {
		m[w] = true
	}
	return m
}

func join(parts []string, sep string) string { return strings.Join(parts, sep) }

// regionCode names an ISO 3166 country code by AIP-143: the role the
// country plays, then region_code. BirthCountry is birth_region_code,
// Nationality nationality_region_code, CountryCode region_code.
func regionCode(w []string) []string {
	var out []string
	for _, word := range w {
		if word != "country" && word != "iso" && word != "code" {
			out = append(out, word)
		}
	}
	return append(out, "region", "code")
}
