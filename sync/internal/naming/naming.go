// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package naming converts OTA's XML names and code names into protobuf ones.
//
// Every conversion runs through [SplitWords], so a rule applied there -- the
// AIP-140 abbreviation table, digit handling, acronym boundaries -- reaches
// field names, message names, enum names and enum values alike. That matters
// because api-linter checks all four.
package naming

import "strings"

// SplitWords breaks an identifier or a code name into lowercase words.
//
//	ResID_Value       ->  [res id value]
//	PMS_InvBlockID    ->  [pms inv block id]
//	Bed & breakfast   ->  [bed and breakfast]
//	Children's menu   ->  [childrens menu]
//
// OTA's PascalCase, its underscores and the prose of its code names all
// arrive here. Punctuation that carries meaning is spelt out; an apostrophe
// is dropped rather than splitting a word; anything else separates words.
func SplitWords(s string) []string {
	var (
		words []string
		cur   []rune
	)
	flush := func() {
		if len(cur) > 0 {
			words = append(words, abbreviate(strings.ToLower(string(cur))))
			cur = nil
		}
	}
	rs := []rune(s)
	for i, r := range rs {
		switch {
		case r == '\'' || r == '’':
		case spelt[r] != "":
			flush()
			words = append(words, spelt[r])
		case isUpper(r):
			if startsWord(rs, i) {
				flush()
			}
			cur = append(cur, r)
		case isLower(r) || isDigit(r):
			cur = append(cur, r)
		default:
			flush()
		}
	}
	flush()
	return words
}

// spelt is punctuation that means a word. "Bed & breakfast" and "Bed or
// breakfast" must not both become BED_BREAKFAST.
var spelt = map[rune]string{'&': "and", '+': "plus", '%': "percent", '#': "number", '@': "at"}

// startsWord reports whether the upper-case rune at i opens a new word: a
// lower-to-upper transition (`fooBar`), or the end of an acronym run, where
// the last capital belongs to the following word (`ISOCountry`).
func startsWord(rs []rune, i int) bool {
	if i == 0 {
		return false
	}
	prev := rs[i-1]
	if isLower(prev) || isDigit(prev) {
		return true
	}
	if !isUpper(prev) || i+1 >= len(rs) || !isLower(rs[i+1]) {
		return false
	}
	// "RPHs" is the plural of RPH, not RP then Hs: a lone trailing s after
	// an acronym stays with it.
	return !(rs[i+1] == 's' && (i+2 == len(rs) || !isLower(rs[i+2])))
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// abbreviations is AIP-140's table of words that must be written short.
//
// `identifier` is deliberately absent: shortening it gives `id`, which then
// fails AIP-148. Where OTA uses it, the catalogue renames the field.
var abbreviations = map[string]string{
	"configuration": "config",
	"information":   "info",
	"specification": "spec",
	"statistics":    "stats",
}

func abbreviate(w string) string {
	if a, ok := abbreviations[w]; ok {
		return a
	}
	return w
}

// Snake renders field case: `RatePlanCode` gives `rate_plan_code`.
func Snake(s string) string { return strings.Join(SplitWords(s), "_") }

// Screaming renders enum value case.
func Screaming(s string) string { return strings.ToUpper(Snake(s)) }

// Pascal renders message case.
func Pascal(s string) string {
	var b strings.Builder
	for _, w := range SplitWords(s) {
		b.WriteString(strings.ToUpper(w[:1]))
		b.WriteString(w[1:])
	}
	return b.String()
}

// LowerCamel renders a resource-name segment, which AIP-122 writes in
// camelCase: `RatePlan` gives `ratePlan`.
func LowerCamel(s string) string {
	p := Pascal(s)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// TypeName renders an XSD type name as a message or enum name, dropping
// OTA's `Type` suffix: `RatePlanType` and `UniqueID_Type` give `RatePlan`
// and `UniqueId`.
func TypeName(xsdName string) string {
	words := SplitWords(xsdName)
	if len(words) > 1 && words[len(words)-1] == "type" {
		words = words[:len(words)-1]
	}
	return Pascal(strings.Join(words, " "))
}

// Humanise turns an identifier into a sentence fragment for a comment.
func Humanise(v string) string {
	words := SplitWords(v)
	if len(words) == 0 {
		return v
	}
	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]
	return strings.Join(words, " ")
}
