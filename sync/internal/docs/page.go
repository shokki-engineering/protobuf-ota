// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package docs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
)

func itoa(n int) string { return strconv.Itoa(n) }

// toRoot is the relative path from a package's README to the repository
// root: protobuf/, then each segment.
func toRoot(p *model.Package) string { return strings.Repeat("../", len(p.Segments())+1) }

// page renders one package's README.
func (g *Generator) page(p *model.Package) string {
	var b strings.Builder
	shape, shapeLink := "collection", "https://aip.dev/121"
	if p.Res.Singleton {
		shape, shapeLink = "singleton", "https://aip.dev/156"
	}
	ms := methods(p)
	fmt.Fprintf(&b, "# %s\n\n", p.Res.Name)
	for _, bd := range []string{
		badge("OTA", g.Pin.SuiteLabel, "0B6B5B", "https://github.com/OpenTravel/OpenTravel-Specifications"),
		badge("code list", g.Pin.CodeListLabel, "1D4ED8", "https://github.com/OpenTravel/OpenTravel-Specifications/tree/master/CodeLists"),
		badge("shape", shape, "2B3172", shapeLink),
		badge("RPCs", itoa(len(ms)), "555", "#methods"),
		badge("package", p.Proto(), "444", "v1/"),
	} {
		b.WriteString(bd + "\n")
	}
	b.WriteString("\n" + g.banner() + "\n")
	b.WriteString(strings.TrimSpace(resourceDoc(p.Message.Doc)) + "\n\n")
	fmt.Fprintf(&b, "```text\n%s\n```\n\n", p.Pattern())

	b.WriteString("## Example\n\nA " + p.Singular() + " as this API returns it:\n\n")
	b.WriteString("```json\n" + example(p) + "\n```\n\n")

	b.WriteString("## Resource names\n\n")
	fmt.Fprintf(&b, "Every %s is addressed by a name that alternates collection and identifier,\n"+
		"per [AIP-122](https://aip.dev/122):\n\n", p.Singular())
	b.WriteString(packet(p.Pattern()) + "\n")
	b.WriteString(identityNote(p) + "\n\n")

	b.WriteString(g.fieldsSection(p))
	b.WriteString(methodsSection(p, ms))
	b.WriteString(codeTables(p))
	return b.String()
}

// resourceDoc is the resource message's doc without the lines the README
// states elsewhere: its shape and its pattern.
func resourceDoc(doc string) string {
	var keep []string
	for _, para := range strings.Split(doc, "\n\n") {
		if strings.Contains(para, "resource, named") {
			continue
		}
		keep = append(keep, para)
	}
	return strings.Join(keep, "\n\n")
}

// identityNote says where the last segment of the name comes from.
func identityNote(p *model.Package) string {
	switch {
	case p.Res.Singleton:
		return "A singleton ([AIP-156](https://aip.dev/156)): its name ends in a literal, because a\n" +
			"parent has exactly one, so there is nothing to identify."
	case p.Res.Identity != "":
		return fmt.Sprintf("The last segment is OTA's `@%s`, verbatim -- OTA codes are case-sensitive,\n"+
			"so it is not lower-cased ([decisions](%sdocs/decisions.md)).", p.Res.Identity, toRoot(p))
	}
	return "The last segment is assigned by the server, or by the caller as `" +
		p.Leaf() + "_id` on create."
}

// methodsSection lists the RPCs with their HTTP bindings.
func methodsSection(p *model.Package, ms []rpc) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Methods\n\nService `%s`", p.Service())
	if p.Res.Singleton {
		b.WriteString(", a singleton's complete set ([AIP-156](https://aip.dev/156)):\n\n")
	} else {
		b.WriteString(", the AIP-131..135 standard methods plus [AIP-164](https://aip.dev/164) undelete:\n\n")
	}
	b.WriteString("| RPC | HTTP | Description |\n| --- | --- | --- |\n")
	for _, m := range ms {
		fmt.Fprintf(&b, "| `%s` | `%s %s` | %s |\n", m.Name, m.Verb, m.Path, m.Doc)
	}
	b.WriteString("\n")
	return b.String()
}
