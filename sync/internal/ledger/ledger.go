// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package ledger keeps codes.lock: every enum value the generator has ever
// emitted, with the number and identifier it was emitted as.
//
// An identifier, once emitted, never changes. proto3 JSON writes enum values
// by name, so OTA renaming a code -- or a later code making two names
// collide -- would otherwise rename a value and break every JSON consumer.
// The ledger only grows: a value OTA removes stays, deprecated, and its
// number is never reused (docs/decisions.md).
//
// A code table's value is keyed by its code, which is its number. An XSD
// enumeration's is keyed by its spelling: its number is positional, and a
// value OTA inserts must not renumber the ones after it.
package ledger

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Entry is one locked value.
type Entry struct {
	Number   int
	Name     string
	Spelling string
}

// Ledger is codes.lock in memory.
type Ledger struct {
	path    string
	enums   map[string]map[string]Entry // enum key -> value key -> entry
	changed bool
}

// Load reads the ledger at path; a missing file is an empty ledger.
func Load(path string) (*Ledger, error) {
	l := &Ledger{path: path, enums: map[string]map[string]Entry{}}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		num, err := strconv.Atoi(p[min(2, len(p)-1)])
		if len(p) != 5 || err != nil {
			return nil, fmt.Errorf("%s:%d: want enum, key, number, name, spelling", path, n)
		}
		l.slot(p[0])[p[1]] = Entry{Number: num, Name: p[3], Spelling: p[4]}
	}
	return l, sc.Err()
}

func (l *Ledger) slot(enum string) map[string]Entry {
	if l.enums[enum] == nil {
		l.enums[enum] = map[string]Entry{}
	}
	return l.enums[enum]
}

// Resolve returns the number and identifier a value is to be emitted as:
// the locked ones if the value was emitted before, or the proposed ones,
// which are then locked. A proposed number or identifier another value of
// the enum already holds is replaced: the number by the next free one, the
// identifier by the identifier with the number appended.
func (l *Ledger) Resolve(enum, key string, e Entry) Entry {
	vals := l.slot(enum)
	if got, ok := vals[key]; ok {
		return got
	}
	maxN := 0
	for _, v := range vals {
		maxN = max(maxN, v.Number)
		if v.Number == e.Number {
			e.Number = -1
		}
	}
	if e.Number < 0 {
		e.Number = maxN + 1
	}
	for _, v := range vals {
		if v.Name == e.Name {
			e.Name = fmt.Sprintf("%s_%d", e.Name, e.Number)
		}
	}
	vals[key] = e
	l.changed = true
	return e
}

// Values are an enum's locked entries by key.
func (l *Ledger) Values(enum string) map[string]Entry { return l.enums[enum] }

// Save writes the ledger back, sorted, when anything was added.
func (l *Ledger) Save() error {
	if !l.changed {
		return nil
	}
	var lines []string
	for enum, vals := range l.enums {
		for key, e := range vals {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%d\t%s\t%s", enum, key, e.Number, e.Name, e.Spelling))
		}
	}
	sort.Strings(lines)
	head := "# codes.lock -- every enum value sync has emitted. Generated; only ever added to.\n" +
		"# enum\tkey\tnumber\tidentifier\tOTA spelling. See sync/internal/ledger.\n"
	return os.WriteFile(l.path, []byte(head+strings.Join(lines, "\n")+"\n"), 0o644)
}
