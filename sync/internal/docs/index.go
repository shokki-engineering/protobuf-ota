// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package docs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// index renders protobuf/opentravel/README.md: what the schema holds, how
// its resources relate, and a link to every package.
func (g *Generator) index() string {
	var b strings.Builder
	colls, singles, rpcs, protos := 0, 0, 0, 2 // the annotation vocabulary's two files
	for _, p := range g.M.Packages {
		if p.Res.Singleton {
			singles++
		} else {
			colls++
		}
		rpcs += len(methods(p))
		protos += fileCount(p)
	}
	b.WriteString("# OpenTravel protobuf reference\n\n" + g.banner() + "\n")
	b.WriteString("| | |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| Packages | %d (%d resources + the annotation vocabulary) |\n", len(g.M.Packages)+1, len(g.M.Packages))
	fmt.Fprintf(&b, "| Collections / singletons | %d / %d |\n", colls, singles)
	fmt.Fprintf(&b, "| RPCs | %d |\n| Protos | %d |\n\n", rpcs, protos)

	b.WriteString("## How the resources relate\n\n```mermaid\nerDiagram\n")
	for _, p := range g.M.Packages {
		if p.Parent == nil {
			continue
		}
		card, label := "||--o{", p.Plural()
		if p.Res.Singleton {
			card, label = "||--||", p.Singular()
		}
		fmt.Fprintf(&b, "  %s %s %s : %q\n", entity(p.Parent.Res.Name), card, entity(p.Res.Name), label)
	}
	for _, a := range associations(g.M) {
		b.WriteString(a + "\n")
	}
	b.WriteString("```\n\n")
	b.WriteString("> A `||--o{` or `||--||` is *containment*: the child's name extends the\n" +
		"> parent's. A `}o--||` is an *association*: the field holds the other resource's\n" +
		"> name, because AIP-215 forbids naming a message in another package.\n\n")

	b.WriteString("## Packages\n\n| Resource | Package | Shape | Pattern |\n| --- | --- | --- | --- |\n")
	for _, p := range g.M.Packages {
		shape := "collection"
		if p.Res.Singleton {
			shape = "singleton"
		}
		fmt.Fprintf(&b, "| [%s](%s) | `%s` | %s | `%s` |\n", p.Res.Name, relLink(p), p.Proto(), shape, p.Pattern())
	}
	return b.String()
}

// fileCount is how many .proto files emit writes for a package.
func fileCount(p *model.Package) int {
	files := map[string]bool{"messages.proto": true, "service.proto": true}
	for _, m := range p.Plan.Messages {
		files[m.File] = true
	}
	for _, e := range p.Plan.Enums {
		files[e.File] = true
	}
	return len(files)
}

// associations are the edges a resource_reference draws: a field anywhere
// in a package that holds another resource's name. One edge per pair of
// resources, labelled with the first such field.
func associations(m *model.Model) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range m.Packages {
		var walk func(msg *plan.Message)
		walk = func(msg *plan.Message) {
			for _, f := range msg.Fields {
				if f.Ref == "" {
					continue
				}
				target := f.Ref[strings.LastIndex(f.Ref, "/")+1:]
				pair := p.Res.Name + ">" + target
				if !seen[pair] {
					seen[pair] = true
					out = append(out, fmt.Sprintf("  %s }o--|| %s : %q", entity(p.Res.Name), entity(target), f.Name))
				}
			}
			for _, n := range msg.Nested {
				walk(n)
			}
		}
		for _, msg := range p.Plan.Messages {
			walk(msg)
		}
	}
	sort.Strings(out)
	return out
}

// entity is a resource as a mermaid entity name: ROOM_TYPE.
func entity(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}
