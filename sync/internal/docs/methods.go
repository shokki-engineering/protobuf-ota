// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package docs

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
)

// rpc is one method as the service declares it.
type rpc struct {
	Name, Verb, Path, Doc string
}

// methods is the package's methods as service.proto declares them, from
// the one list emit writes it from: model.Package.RPCs.
func methods(p *model.Package) []rpc {
	var out []rpc
	for _, r := range p.RPCs() {
		out = append(out, rpc{r.Name, strings.ToUpper(r.Verb), r.Path, r.Doc})
	}
	return out
}

// packet renders a resource name as a mermaid packet diagram: each segment
// occupies as many cells as it has characters, with the slashes between.
func packet(pattern string) string {
	var b strings.Builder
	b.WriteString("```mermaid\npacket\n")
	pos := 0
	for i, seg := range strings.Split(pattern, "/") {
		if i > 0 {
			b.WriteString(itoa(pos) + ": \"/\"\n")
			pos++
		}
		end := pos + len(seg) - 1
		if end > pos {
			b.WriteString(itoa(pos) + "-" + itoa(end))
		} else {
			b.WriteString(itoa(pos))
		}
		b.WriteString(": \"" + seg + "\"\n")
		pos = end + 1
	}
	b.WriteString("```\n")
	return b.String()
}
