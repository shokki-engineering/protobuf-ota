// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// stringRule renders one buf.validate string rule: protobuf accepts a
// single rule per field, so length and pattern are intersected into it.
func stringRule(min, max *int, pattern string) string {
	var parts []string
	if min != nil && *min > 0 {
		parts = append(parts, fmt.Sprintf("min_len: %d", *min))
	}
	if max != nil {
		parts = append(parts, fmt.Sprintf("max_len: %d", *max))
	}
	if pattern != "" {
		parts = append(parts, "pattern: "+strconv.Quote(pattern))
	}
	return "string = {" + strings.Join(parts, " ") + "}"
}

// rePattern turns one derivation step's XSD patterns into one anchored RE2
// pattern. XSD patterns are implicitly anchored and RE2's are not; XSD
// treats ^ and $ as literals; and XSD has classes -- \i, \c -- and class
// subtraction that RE2 lacks, which are refused rather than approximated.
func rePattern(alts []string) (string, error) {
	var out []string
	for _, p := range alts {
		if strings.Contains(p, `\i`) || strings.Contains(p, `\c`) || strings.Contains(p, "-[") {
			return "", fmt.Errorf("uses an XSD-only construct")
		}
		p = strings.NewReplacer("^", `\^`, "$", `\$`).Replace(p)
		out = append(out, p)
	}
	re := "^(?:" + strings.Join(out, "|") + ")$"
	if len(out) == 1 {
		re = "^" + out[0] + "$"
	}
	_, err := regexp.Compile(re)
	return re, err
}

// otaCode is OTA_CodeType's pattern, by which a code-valued type is known
// whatever restriction it was reached through.
const otaCode = `[0-9A-Z]{1,3}(\.[A-Z]{3}(\.X){0,1}){0,1}`

func isCode(t *xsd.Simple) bool {
	for _, step := range t.Patterns {
		for _, p := range step {
			if p == otaCode {
				return true
			}
		}
	}
	return false
}

// name is the type as an error message or a union member name sees it.
func (t ftype) name() string {
	switch {
	case t.Msg != nil:
		return t.Msg.Name
	case t.Enum != nil:
		return t.Enum.Name
	}
	return t.Type
}

// celPattern is a pattern as a protovalidate CEL rule. The regular
// expression is a CEL raw string, so its backslashes are not CEL escapes.
func celPattern(re string, i int) (string, error) {
	if strings.Contains(re, "'") {
		return "", fmt.Errorf("pattern %q cannot be a CEL raw string", re)
	}
	expr := "this.matches(r'" + re + "')"
	return fmt.Sprintf("cel = {id: %s message: %s expression: %s}",
		strconv.Quote(fmt.Sprintf("pattern.step%d", i+1)),
		strconv.Quote("must match the pattern OTA's base type declares"),
		strconv.Quote(expr)), nil
}
