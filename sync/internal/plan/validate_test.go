// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRepeatedRules(t *testing.T) {
	for _, c := range []struct {
		in   []string
		want []string
	}{
		{ // an item rule moves under items and joins the count bound
			[]string{"string = {max_len: 8}", "ignore = IGNORE_IF_ZERO_VALUE", "repeated = {max_items: 99}"},
			[]string{"repeated = {max_items: 99 items: {string: {max_len: 8}}}"},
		},
		{[]string{"repeated = {max_items: 5}"}, []string{"repeated = {max_items: 5}"}},
		{[]string{"ignore = IGNORE_IF_ZERO_VALUE"}, nil},
		{nil, nil},
		{ // a CEL rule is an item rule too
			[]string{`cel = {id: "x"}`},
			[]string{`repeated = {items: {cel: {id: "x"}}}`},
		},
	} {
		if got := repeatedRules(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("repeatedRules(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRePattern(t *testing.T) {
	for _, c := range []struct {
		in   []string
		want string
	}{
		{[]string{"[A-Z]{3}"}, "^[A-Z]{3}$"}, // XSD is implicitly anchored
		{[]string{"a", "b"}, "^(?:a|b)$"},    // one step's patterns are alternatives
		{[]string{"a^b$"}, `^a\^b\$$`},       // ^ and $ are XSD literals
		{[]string{`[0-9]+(\.[0-9]+)?`}, `^[0-9]+(\.[0-9]+)?$`},
	} {
		got, err := rePattern(c.in)
		if err != nil || got != c.want {
			t.Errorf("rePattern(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, refused := range []string{`\i\c*`, `[a-z-[aeiou]]`} {
		if _, err := rePattern([]string{refused}); err == nil {
			t.Errorf("rePattern(%q) accepted an XSD-only construct", refused)
		}
	}
}

func TestCelPattern(t *testing.T) {
	got, err := celPattern(`^\d+$`, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{`id: "pattern.step1"`, strconv.Quote(`this.matches(r'^\d+$')`)} {
		if !strings.Contains(got, part) {
			t.Errorf("celPattern = %s, missing %s", got, part)
		}
	}
	if _, err := celPattern("a'b", 0); err == nil {
		t.Error("celPattern accepted a quote, which would end the CEL raw string")
	}
}

func TestStringRuleIntersects(t *testing.T) {
	one, zero, max := 1, 0, 64
	if got := stringRule(&one, &max, "^x$"); got != `string = {min_len: 1 max_len: 64 pattern: "^x$"}` {
		t.Errorf("stringRule = %s", got)
	}
	if got := stringRule(&zero, &max, ""); got != "string = {max_len: 64}" {
		t.Errorf("stringRule with min 0 = %s; a zero minimum says nothing", got)
	}
}

func TestSanitise(t *testing.T) {
	got := Sanitise("Wraps </Text> and <Image/>.\n# Heading")
	if strings.ContainsAny(got, "<>") || strings.Contains(got, "\n# ") {
		t.Errorf("Sanitise left markup api-linter rejects: %q", got)
	}
	if !strings.Contains(got, "Text") || !strings.Contains(got, "Image") {
		t.Errorf("Sanitise lost the words: %q", got)
	}
}
