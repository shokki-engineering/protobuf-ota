// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package model decides what becomes a package and a resource, and what
// each is called: the resources table, resource names and patterns, and the
// standard fields every resource carries. What each OTA member becomes
// inside a package is plan's job.
package model

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/codelist"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/ledger"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// Model is every package the generator writes.
type Model struct {
	Packages []*Package
}

// Package is one resource's proto package.
type Package struct {
	Res      *Resource
	Parent   *Package
	Plan     *plan.Builder
	Message  *plan.Message // the resource message
	Children []*Package

	Customs               []*Custom
	customReq, customResp map[string]*plan.Message
}

// Leaf is the package's last segment, the resource in snake case.
func (p *Package) Leaf() string { return naming.Snake(p.Res.Name) }

// Segments are the directories under protobuf/, before v1.
func (p *Package) Segments() []string {
	s := []string{"opentravel", p.Res.Spec}
	if p.Res.Domain != "" {
		s = append(s, p.Res.Domain)
	}
	return append(s, p.Leaf())
}

// Proto is the proto package name.
func (p *Package) Proto() string {
	return "protobuf." + strings.Join(p.Segments(), ".") + ".v1"
}

// Singular and Plural are the resource's AIP-123 names, in lowerCamel.
func (p *Package) Singular() string { return naming.LowerCamel(p.Res.Name) }
func (p *Package) Plural() string   { return naming.Pluralise(p.Singular()) }

// Service is the service's name: the plural, or for a resource whose plural
// is its singular -- AreaInfo, Multimedia -- the name with "Service", since a
// service and a message of one package cannot share a name.
func (p *Package) Service() string {
	if naming.Pascal(p.Plural()) == p.Res.Name {
		return p.Res.Name + "Service"
	}
	return naming.Pascal(p.Plural())
}

// Type is the resource type, e.g. "hotel.opentravel.org/RatePlan".
func (p *Package) Type() string { return p.Res.Spec + ".opentravel.org/" + p.Res.Name }

// Var is the pattern variable naming the resource's id, e.g. "rate_plan".
func (p *Package) Var() string { return p.Leaf() }

// Pattern is the resource name pattern. A collection alternates collection
// and id; a singleton ends in a literal (AIP-156).
func (p *Package) Pattern() string {
	tail := p.Plural() + "/{" + p.Var() + "}"
	if p.Res.Singleton {
		tail = p.Singular()
	}
	if p.Parent == nil {
		return tail
	}
	return p.Parent.Pattern() + "/" + tail
}

// ParentPattern is the parent's pattern, or "" for a root collection.
func (p *Package) ParentPattern() string {
	if p.Parent == nil {
		return ""
	}
	return p.Parent.Pattern()
}

// Regexp is a pattern as an anchored RE2 expression, for buf.validate.
func Regexp(pattern string) string {
	var b strings.Builder
	for _, seg := range strings.Split(pattern, "/") {
		if b.Len() > 0 {
			b.WriteString("/")
		}
		if strings.HasPrefix(seg, "{") {
			b.WriteString("[^/]+")
		} else {
			b.WriteString(seg)
		}
	}
	return "^" + b.String() + "$"
}

// Build plans every package in the resources table. lock is codes.lock,
// through which every enum value passes; nil plans them unlocked.
func Build(s *xsd.Schema, codes *codelist.List, lock *ledger.Ledger) (*Model, error) {
	m := &Model{}
	byName := map[string]*Package{}
	for _, r := range resources {
		p := &Package{Res: r, Plan: plan.New(codes)}
		p.Plan.Ledger = lock
		byName[r.Name] = p
		m.Packages = append(m.Packages, p)
	}
	for _, p := range m.Packages {
		if p.Res.Parent == "" {
			continue
		}
		parent, ok := byName[p.Res.Parent]
		if !ok {
			return nil, fmt.Errorf("%s: no parent resource %q", p.Res.Name, p.Res.Parent)
		}
		p.Parent = parent
		parent.Children = append(parent.Children, p)
	}
	var errs []string
	for _, p := range m.Packages {
		if err := p.build(s); err != nil {
			errs = append(errs, err.Error())
		}
		if err := p.buildCustoms(s); err != nil {
			errs = append(errs, err.Error())
		}
		p.Plan.Group()
		for _, e := range p.Plan.Errors() {
			errs = append(errs, p.Res.Name+": "+e)
		}
	}
	if err := checkUnique(m); err != nil {
		errs = append(errs, err.Error())
	}
	errs = append(errs, checkMethods(m, s)...)
	errs = append(errs, CheckCatalogue(s)...)
	if len(errs) > 0 {
		return m, fmt.Errorf("%d problems:\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return m, nil
}

// checkUnique asserts no two packages share a proto package, a pattern or a
// resource type -- the failure coversa found produces output every linter
// accepts.
func checkUnique(m *Model) error {
	seen := map[string]string{}
	for _, p := range m.Packages {
		for _, k := range []string{"package " + p.Proto(), "pattern " + p.Pattern(), "type " + p.Type()} {
			if prev, dup := seen[k]; dup {
				return fmt.Errorf("%s and %s share %s", prev, p.Res.Name, k)
			}
			seen[k] = p.Res.Name
		}
	}
	return nil
}
