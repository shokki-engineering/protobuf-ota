// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"fmt"
	"strings"
)

// writer accumulates one file's text.
type writer struct{ b strings.Builder }

func (w *writer) line(indent int, format string, args ...any) {
	if format == "" {
		w.b.WriteString("\n")
		return
	}
	w.b.WriteString(strings.Repeat("  ", indent))
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteString("\n")
}

// comment writes doc as // lines wrapped to 80 columns, keeping paragraph
// breaks. Every declaration has one (AIP-192).
func (w *writer) comment(indent int, doc string) {
	width := 80 - 2*indent - 3
	for i, para := range strings.Split(strings.TrimSpace(doc), "\n\n") {
		if i > 0 {
			w.line(indent, "//")
		}
		for _, wrapped := range wrap(strings.ReplaceAll(para, "\n", " "), width) {
			w.line(indent, "// %s", wrapped)
		}
	}
}

// wrap breaks a line at spaces so no piece exceeds width, unless a single
// word does.
func wrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var out []string
	cur := words[0]
	for _, word := range words[1:] {
		if len(cur)+1+len(word) > width {
			out = append(out, cur)
			cur = word
		} else {
			cur += " " + word
		}
	}
	return append(out, cur)
}

func (w *writer) String() string { return w.b.String() }

// quote renders a string literal for a .proto option.
func quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + r.Replace(s) + `"`
}
