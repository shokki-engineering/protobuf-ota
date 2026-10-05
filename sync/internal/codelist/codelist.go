// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package codelist reads the OpenTravel code list workbook.
//
// The workbook is SpreadsheetML -- a zip of XML parts -- and is read with the
// standard library alone. Columns are found by their header text, not their
// letter, and a header this reader needs and cannot find is an error: a
// workbook reorganised between releases must fail here rather than yield
// codes read from the wrong column.
package codelist

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// List is every code table in the workbook, by abbreviation.
type List struct {
	Tables map[string]*Table
}

// Table is one code table: Room Amenity Type, abbreviated RMA.
type Table struct {
	Abbr  string
	Name  string
	Codes []Code // in workbook order
}

// Code is one value of a table. Every code in the pinned workbook is a
// positive integer, and the reader insists on it.
type Code struct {
	Value      int
	Name       string
	Definition string
	Changed    string // the workbook's create/change date
	Hotel      bool   // marked as applying to hotels
}

// columns are the headers the reader needs, normalised.
var columns = []string{
	"CODE LIST NAME", "CODE VALUE", "CODE LIST NAME DETAIL", "CODE NAME",
	"HOTEL/ HOSTEL", "DEFINITION", "CREATE/ CHANGE DATE",
}

// Load reads the workbook at file.
func Load(file string) (*List, error) {
	w, err := openWorkbook(file)
	if err != nil {
		return nil, err
	}
	defer w.Close()
	rows, err := w.rows("Codelists")
	if err != nil {
		return nil, err
	}

	col, start, err := header(rows)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	l := &List{Tables: map[string]*Table{}}
	for i, r := range rows[start:] {
		get := func(h string) string { return clean(r[col[h]]) }
		abbr, detail := get("CODE VALUE"), get("CODE LIST NAME DETAIL")
		switch {
		case get("CODE LIST NAME") != "":
			// A table heading: name in the first column, abbreviation in
			// the second. Only headings fill the first column; some repeat
			// the abbreviation in the third, so that one cannot decide.
			l.Tables[abbr] = &Table{Abbr: abbr, Name: get("CODE LIST NAME")}
		case detail != "":
			t, ok := l.Tables[detail]
			if !ok {
				return nil, fmt.Errorf("%s: row %d: code for undeclared table %q", file, start+i+1, detail)
			}
			v, err := strconv.Atoi(abbr)
			if err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: row %d: %s code %q is not a positive integer", file, start+i+1, detail, abbr)
			}
			t.Codes = append(t.Codes, Code{
				Value: v, Name: get("CODE NAME"), Definition: get("DEFINITION"),
				Changed: get("CREATE/ CHANGE DATE"), Hotel: get("HOTEL/ HOSTEL") != "",
			})
		}
	}
	return l, nil
}

// header finds the column-heading row and maps each needed heading to its
// column letter.
func header(rows []map[string]string) (map[string]string, int, error) {
	for i, r := range rows {
		col := map[string]string{}
		for letter, text := range r {
			col[clean(text)] = letter
		}
		if _, ok := col["CODE VALUE"]; !ok {
			continue
		}
		for _, h := range columns {
			if _, ok := col[h]; !ok {
				return nil, 0, fmt.Errorf("no %q column", h)
			}
		}
		return col, i + 1, nil
	}
	return nil, 0, fmt.Errorf("no heading row")
}

var space = regexp.MustCompile(`\s+`)

func clean(s string) string { return strings.TrimSpace(space.ReplaceAllString(s, " ")) }

// citation matches OTA's prose reference to a table. The wording varies --
// "Code List", "Code list", "Code list OpenTravel Code List" -- so what is
// matched is the three-letter abbreviation in parentheses after it.
var citation = regexp.MustCompile(`(?i)Refer to OpenTravel Code ?List[^()]*\(([A-Z]{3})\)`)

// Cited returns the abbreviations a piece of documentation cites, in order,
// without repeats.
func Cited(doc string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range citation.FindAllStringSubmatch(doc, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// Abbrs lists the table abbreviations, sorted.
func (l *List) Abbrs() []string {
	out := make([]string, 0, len(l.Tables))
	for a := range l.Tables {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}
