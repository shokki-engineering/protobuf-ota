// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// inScope reports whether an OTA message is one the methods table must map:
// every hotel request, and the generic requests a hotel uses.
func inScope(name string) bool {
	if !strings.HasSuffix(name, "RQ") {
		return false
	}
	switch {
	case strings.HasPrefix(name, "OTA_Hotel"), strings.HasPrefix(name, "OTA_Profile"):
		return true
	}
	return name == "OTA_CancelRQ" || name == "OTA_ReadRQ" || name == "OTA_DeleteRQ"
}

// checkMethods enforces the methods table (CLAUDE.md rule 7): every message
// in scope has exactly one row, every row names a message the suite has,
// and every RPC a row names is declared by its package.
func checkMethods(m *Model, s *xsd.Schema) []string {
	var errs []string
	declared := map[string]bool{}
	for _, p := range m.Packages {
		for _, r := range p.RPCs() {
			declared[p.Res.Name+"."+r.Name] = true
		}
	}
	rows := map[string]int{}
	for _, row := range methods {
		rows[row.Message]++
		if (len(row.RPCs) == 0) == (row.NotTaken == "") {
			errs = append(errs, fmt.Sprintf("methods: %s needs RPCs or a NotTaken reason, not both", row.Message))
		}
		for _, r := range row.RPCs {
			if !declared[r] {
				errs = append(errs, fmt.Sprintf("methods: %s maps to %s, which no package declares", row.Message, r))
			}
		}
	}
	suite := map[string]bool{}
	for _, name := range s.MessageNames() {
		suite[name] = true
		if inScope(name) && rows[name] != 1 {
			errs = append(errs, fmt.Sprintf("methods: %s has %d rows; every message in scope has exactly one", name, rows[name]))
		}
	}
	for name := range rows {
		if !suite[name] {
			errs = append(errs, fmt.Sprintf("methods: %s is not a message of the suite", name))
		}
	}
	return errs
}
