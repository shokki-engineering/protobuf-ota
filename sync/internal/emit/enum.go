// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"fmt"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// enumFile writes one top-level enum.
func (e *Emitter) enumFile(p *model.Package, en *plan.Enum) string {
	imports := map[string]bool{}
	if enumNeedsVocab(en) {
		imports["protobuf/opentravel/annotations/v1/annotations.proto"] = true
	}
	var w writer
	e.header(&w, subPackage(p, en.File), en.File, imports)
	writeEnum(&w, en, 0)
	return w.String()
}

// enumNeedsVocab reports whether an enum carries an OTA annotation: every
// enum from OTA does, a value's spelling being the least of it.
func enumNeedsVocab(en *plan.Enum) bool { return len(en.Values) > 0 }

// writeEnum writes an enum. The zero value is AIP-126's UNSPECIFIED; every
// OTA value keeps its OTA spelling in an annotation, and a code table's
// value its OTA code as its number.
func writeEnum(w *writer, en *plan.Enum, indent int) {
	w.comment(indent, en.Doc)
	w.line(indent, "enum %s {", en.Name)
	in := indent + 1
	switch {
	case en.CodeList != "":
		w.line(in, "option (%s.enumeration) = {", vocab)
		w.line(in+1, "code_list: %s", quote(en.CodeList))
		w.line(in+1, "table: %s", quote(en.Table))
		w.line(in, "};")
		w.line(0, "")
	case en.Origin != "":
		w.line(in, "option (%s.enumeration) = {xsd_type: %s};", vocab, quote(en.Origin))
		w.line(0, "")
	}
	prefix := naming.Screaming(en.Name)
	w.comment(in, "Not specified.")
	w.line(in, "%s_UNSPECIFIED = 0;", prefix)
	for _, v := range en.Values {
		w.line(0, "")
		doc := v.Doc
		if doc == "" {
			doc = fmt.Sprintf("OTA's %q.", v.Spelling)
		}
		if v.Deprecated {
			doc = "Deprecated: " + doc
		}
		w.comment(in, doc)
		opts := fmt.Sprintf("(%s.value) = {source_spelling: %s}", vocab, quote(v.Spelling))
		if v.Deprecated {
			opts += ", deprecated = true"
		}
		w.line(in, "%s = %d [%s];", v.Name, v.Number, opts)
	}
	w.line(indent, "}")
}
