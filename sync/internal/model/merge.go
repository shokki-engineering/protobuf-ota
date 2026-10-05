// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// merge adds a second source's members to the resource, refusing a field
// name both sources produce until the merges table drops one.
func (p *Package) merge(s *xsd.Schema, m *plan.Message) error {
	mg := p.Res.Merge
	e, err := s.Find(mg.Root)
	if err != nil || e.Complex == nil {
		return fmt.Errorf("%s: merge root %s: %v", p.Res.Name, mg.Root, err)
	}
	if err := membersExist(p.Res.Name, mg.Root, e.Complex, mg.Identity, mg.Carve); err != nil {
		return err
	}
	first := len(m.Fields)
	p.Plan.Fill(m, e.Complex, p.skipper(mg.Identity, mg.Carve, mg.Drop))
	have := map[string]string{}
	for _, f := range m.Fields[:first] {
		have[f.Name] = f.Path
	}
	for _, f := range m.Fields[first:] {
		if prev, dup := have[f.Name]; dup {
			return fmt.Errorf("%s: %s and merged %s both become %q: add the second to Merge.Drop", p.Res.Name, prev, f.Path, f.Name)
		}
	}
	return nil
}
