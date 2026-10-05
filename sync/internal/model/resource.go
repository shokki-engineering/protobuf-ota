// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// absorbed are OTA's own record-keeping attributes, which on a resource
// become AIP's standard fields rather than fields of their own (CLAUDE.md
// rule 8).
var absorbed = map[string]string{
	"CreateDateTime":     "create_time",
	"LastModifyDateTime": "update_time",
	"LastUpdated":        "update_time",
}

// build plans the package: the resource message, then everything it reaches.
func (p *Package) build(s *xsd.Schema) error {
	root, err := s.Find(p.Res.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Res.Name, err)
	}
	c := root.Complex
	if c == nil {
		return fmt.Errorf("%s: root %s has simple content", p.Res.Name, p.Res.Root)
	}
	if err := p.checkMembers(c); err != nil {
		return err
	}

	doc := root.Doc
	if doc == "" {
		doc = c.Doc
	}
	shape := "A collection"
	if p.Res.Singleton {
		shape = "A singleton (AIP-156 <https://aip.dev/156>)"
	}
	m := &plan.Message{
		Name: p.Res.Name, Origin: p.Res.Root, Reserved: [2]int{8, 15},
		Doc: plan.Sanitise(fmt.Sprintf("%s\n\n%s resource, named %q.\n\nOTA: %s.", doc, shape, p.Pattern(), p.Res.Root)),
		Resource: &plan.Resource{
			Type: p.Type(), Pattern: p.Pattern(), Singular: p.Singular(), Plural: p.Plural(),
			Singleton: p.Res.Singleton,
		},
	}
	m.Fields = p.standardFields()
	p.Message = m
	p.Plan.Top(m)
	p.Plan.Root(m, "")

	renames := map[string]string{}
	for k, v := range p.Res.Renames {
		renames[k] = v
	}
	for _, a := range c.Attrs {
		if a.Name == "ID" && a.Name != p.Res.Identity {
			renames[a.Key] = "ota_uid" // docs/conventions.md rule 10
			p.Plan.Immutable[a.Key] = true
		}
	}
	p.Plan.ResourceRenames = renames
	p.Plan.Fill(m, c, p.skipper(p.Res.Identity, p.Res.Carve, nil))
	if p.Res.Merge != nil {
		return p.merge(s, m)
	}
	return nil
}

// skipper is the members a resource's source does not carry as fields: the
// identity, OTA's absorbed timestamps, the carve-outs, and dropped paths.
func (p *Package) skipper(identity string, carve, drop []string) func(string) bool {
	return func(key string) bool {
		for _, d := range drop {
			if d == key {
				return true
			}
		}
		name := key[strings.LastIndex(key, "/")+1:]
		if strings.HasPrefix(name, "@") {
			attr := name[1:]
			return attr == identity || absorbed[attr] != "" || carvedBy(carve, name, key)
		}
		return carvedBy(carve, name, key)
	}
}

// carved reports whether a root member is carved out.
func (p *Package) carved(name string, key ...string) bool { return carvedBy(p.Res.Carve, name, key...) }

// carvedBy reports whether a member is carved out: by its name, or by the
// attribute group or type that declares it -- "HotelReferenceGroup/*"
// carves every attribute the parent's reference group contributes.
func carvedBy(carve []string, name string, key ...string) bool {
	for _, c := range carve {
		if c == name {
			return true
		}
		if prefix, ok := strings.CutSuffix(c, "/*"); ok && len(key) > 0 && strings.HasPrefix(key[0], prefix+"/") {
			return true
		}
	}
	return false
}

// checkMembers fails on a table entry naming a member the root does not
// have: an identity or a carve-out that resolves nothing is a decision
// recorded about an attribute that no longer exists (rule 15).
func (p *Package) checkMembers(c *xsd.Complex) error {
	return membersExist(p.Res.Name, p.Res.Root, c, p.Res.Identity, p.Res.Carve)
}

// membersExist fails on an identity or carve-out a source does not have.
func membersExist(res, root string, c *xsd.Complex, identity string, carve []string) error {
	has := map[string]bool{}
	for _, a := range c.Attrs {
		has["@"+a.Name] = true
	}
	for _, e := range c.Elems {
		has[e.Name] = true
	}
	var want []string
	for _, cv := range carve {
		if !strings.HasSuffix(cv, "/*") {
			want = append(want, cv)
		}
	}
	if identity != "" {
		want = append(want, "@"+identity)
	}
	for _, w := range want {
		if !has[w] {
			return fmt.Errorf("%s: %s has no member %s", res, root, w)
		}
	}
	return nil
}

// standardFields are the AIP fields every resource carries, numbered 1-7
// with 8-15 held back, as in coversa. Soft delete (AIP-164) applies to a
// collection; nothing can delete a singleton.
func (p *Package) standardFields() []*plan.Field {
	name := &plan.Field{
		Name: "name", Number: 1, Type: "string", Behavior: []string{"IDENTIFIER"},
		Doc: fmt.Sprintf("Resource name, %q.", p.Pattern()),
		Validate: []string{"string = {pattern: " + strconv.Quote(Regexp(p.Pattern())) + "}",
			"ignore = IGNORE_IF_ZERO_VALUE"},
	}
	if p.Res.Identity != "" {
		name.Doc += fmt.Sprintf("\n\nThe last segment is OTA's @%s, verbatim: OTA codes are case-sensitive\n(docs/decisions.md).", p.Res.Identity)
	} else if !p.Res.Singleton {
		name.Doc += " The last segment is assigned by the server."
	}
	ts := func(name string, num int, doc string) *plan.Field {
		return &plan.Field{Name: name, Number: num, Type: "google.protobuf.Timestamp",
			TypeFile: "google/protobuf/timestamp.proto", Behavior: []string{"OUTPUT_ONLY"}, Doc: doc}
	}
	fields := []*plan.Field{
		name,
		{Name: "uid", Number: 2, Type: "string", Behavior: []string{"OUTPUT_ONLY"}, UUID: true,
			Doc:      "Server-assigned unique id.",
			Validate: []string{"string = {uuid: true}", "ignore = IGNORE_IF_ZERO_VALUE"}},
		{Name: "etag", Number: 3, Type: "string", Behavior: []string{"OUTPUT_ONLY"},
			Doc: "Weak validator for optimistic concurrency, AIP-154 <https://aip.dev/154>."},
	}
	if !p.Res.Singleton {
		fields = append(fields,
			ts("delete_time", 4, "When this resource was soft-deleted, unset while it is live. AIP-164\n<https://aip.dev/164>."),
			ts("expire_time", 5, "When a soft-deleted resource is purged for good. AIP-164\n<https://aip.dev/164>."))
	}
	return append(fields,
		ts("create_time", 6, "When this resource was created. OTA's CreateDateTime, where OTA carries it."),
		ts("update_time", 7, "When this resource was last modified. OTA's LastModifyDateTime and\nLastUpdated, where OTA carries them."))
}
