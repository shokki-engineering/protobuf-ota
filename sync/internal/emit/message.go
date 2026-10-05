// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

const vocab = "protobuf.opentravel.annotations.v1"

// messageFile writes one top-level message and everything nested in it --
// or, for a cycle of types, each message of the cycle (plan.Group).
func (e *Emitter) messageFile(p *model.Package, base string, ms []*plan.Message) string {
	imports := map[string]bool{}
	for _, m := range ms {
		collectImports(p, m, imports)
	}
	delete(imports, local(p, base))
	var w writer
	e.header(&w, subPackage(p, base), base, imports)
	for i, m := range ms {
		if i > 0 {
			w.line(0, "")
		}
		writeMessage(&w, p, m, 0)
	}
	return w.String()
}

// collectImports gathers what a message and its nested declarations need.
func collectImports(p *model.Package, m *plan.Message, imports map[string]bool) {
	if m.Resource != nil {
		imports["google/api/resource.proto"] = true
	}
	if m.Origin != "" {
		imports["protobuf/opentravel/annotations/v1/annotations.proto"] = true
	}
	for _, f := range m.Fields {
		if _, file := f.Target(); importFor(p, file) != "" {
			imp := importFor(p, file)
			imports[imp] = true
		}
		if len(f.Behavior) > 0 {
			imports["google/api/field_behavior.proto"] = true
		}
		if len(f.Validate) > 0 {
			imports["buf/validate/validate.proto"] = true
		}
		if f.Ref != "" || f.ChildRef != "" {
			imports["google/api/resource.proto"] = true
		}
		if f.UUID {
			imports["google/api/field_info.proto"] = true
		}
		if f.Path != "" {
			imports["protobuf/opentravel/annotations/v1/annotations.proto"] = true
		}
	}
	for _, en := range m.Enums {
		if enumNeedsVocab(en) {
			imports["protobuf/opentravel/annotations/v1/annotations.proto"] = true
		}
	}
	for _, n := range m.Nested {
		collectImports(p, n, imports)
	}
}

// writeMessage writes m at the given depth.
func writeMessage(w *writer, p *model.Package, m *plan.Message, indent int) {
	w.comment(indent, m.Doc)
	w.line(indent, "message %s {", m.Name)
	in := indent + 1
	if r := m.Resource; r != nil {
		w.line(in, "option (google.api.resource) = {")
		w.line(in+1, "type: %s", quote(r.Type))
		w.line(in+1, "pattern: %s", quote(r.Pattern))
		w.line(in+1, "singular: %s", quote(r.Singular))
		w.line(in+1, "plural: %s", quote(r.Plural))
		w.line(in, "};")
		w.line(0, "")
	}
	if m.Origin != "" {
		w.line(in, "option (%s.element) = {path: %s};", vocab, quote(m.Origin))
		w.line(0, "")
	}
	written := map[string]bool{}
	for i, f := range m.Fields {
		switch {
		case f.OneOf == "":
			writeField(w, p, m, f, in)
		case !written[f.OneOf]:
			written[f.OneOf] = true
			w.line(in, "// One of OTA's alternatives; at most one is present.")
			w.line(in, "oneof %s {", f.OneOf)
			for _, g := range m.Fields {
				if g.OneOf == f.OneOf {
					writeField(w, p, m, g, in+1)
				}
			}
			w.line(in, "}")
		default:
			continue
		}
		if m.Reserved[0] != 0 && f.Number < m.Reserved[0] && (i+1 == len(m.Fields) || m.Fields[i+1].Number > m.Reserved[1]) {
			w.line(0, "")
			w.line(in, "reserved %d to %d;", m.Reserved[0], m.Reserved[1])
		}
		w.line(0, "")
	}
	for _, n := range m.Nested {
		writeMessage(w, p, n, in)
		w.line(0, "")
	}
	for _, en := range m.Enums {
		writeEnum(w, en, in)
		w.line(0, "")
	}
	trimBlank(w)
	w.line(indent, "}")
}

// trimBlank drops a trailing blank line before a closing brace.
func trimBlank(w *writer) {
	s := w.b.String()
	if strings.HasSuffix(s, "\n\n") {
		w.b.Reset()
		w.b.WriteString(s[:len(s)-1])
	}
}
