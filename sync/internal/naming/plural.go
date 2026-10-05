// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package naming

import "strings"

// irregularPlurals are the words English does not pluralise by adding an
// "s". The table is not speculative: each entry is a word OTA uses as the
// last word of a resource or a repeated field.
//
// A plural is not cosmetic. AIP-123 ties a resource's `plural` to the
// collection segment of its name and to its service name, and AIP-144 wants
// a repeated field's name plural.
var irregularPlurals = map[string]string{
	"multimedia": "multimedia",
	"info":       "info",
	"person":     "people",
	"criteria":   "criteria",
	"criterion":  "criteria",
	"data":       "data",
	"media":      "media",
	"series":     "series",
	"news":       "news",
}

// Pluralise renders the plural of an identifier, changing only its last
// word: `ratePlan` gives `ratePlans`, `rate_plan` gives `rate_plans`.
func Pluralise(s string) string {
	i := strings.LastIndexAny(s, "_ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if i < 0 || s[i] == '_' {
		i++
	}
	head, last := s[:i], s[i:]
	lower := strings.ToLower(last)
	if p, ok := irregularPlurals[lower]; ok {
		return head + last[:1] + p[1:]
	}
	switch {
	case strings.HasSuffix(lower, "s") && !hasAnySuffix(lower, "ss", "us", "is"):
		return s // already plural: OTA names many repeated elements so
	case strings.HasSuffix(lower, "y") && !endsVowelY(lower):
		return head + last[:len(last)-1] + "ies"
	case hasAnySuffix(lower, "s", "x", "z", "ch", "sh"):
		return head + last + "es"
	default:
		return head + last + "s"
	}
}

func endsVowelY(s string) bool {
	return len(s) >= 2 && strings.ContainsRune("aeiou", rune(s[len(s)-2]))
}

func hasAnySuffix(s string, suffixes ...string) bool {
	for _, x := range suffixes {
		if strings.HasSuffix(s, x) {
			return true
		}
	}
	return false
}
