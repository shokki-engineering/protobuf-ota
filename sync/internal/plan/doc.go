// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"regexp"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// fieldDoc is a field's comment: OTA's documentation, or a sentence made
// from the name when OTA has none, then what the schema adds -- the code
// table, a former default, the OTA path the field carries (CLAUDE.md rule 3,
// docs/conventions.md rule 12).
func fieldDoc(xmlName, doc string, notes xsd.Notes, t ftype, key string) string {
	var paras []string
	if notes.Deprecation != "" {
		// AIP-192: a deprecated element's comment opens with "Deprecated:".
		paras = append(paras, "Deprecated: "+notes.Deprecation)
	}
	if doc == "" {
		doc = naming.Humanise(xmlName) + "."
	}
	paras = append(paras, doc)
	if t.CodeList != "" {
		paras = append(paras, "Codes from the OpenTravel code table "+t.CodeList+"; each value's number is its OTA code.")
	}
	if t.Note != "" {
		paras = append(paras, t.Note)
	}
	if notes.LegacyDefault != "" {
		paras = append(paras, "When absent, OTA formerly assumed \""+notes.LegacyDefault+"\".")
	}
	paras = append(paras, "OTA: "+key+".")
	return Sanitise(strings.Join(paras, "\n\n"))
}

var (
	htmlTag = regexp.MustCompile(`</ *([a-zA-Z-]+)>|<([a-zA-Z-]+) */>`)
	heading = regexp.MustCompile(`(?m)^\s?(#+) `)
)

// Sanitise makes OTA's prose acceptable to AIP-192's comment rules, which
// read element names written as tags as HTML and a leading # as a markdown
// heading. The words survive; only the markup that trips the rule goes.
func Sanitise(s string) string {
	s = htmlTag.ReplaceAllString(s, "$1$2")
	s = heading.ReplaceAllString(s, "$1")
	return strings.TrimSpace(s)
}
