// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// writeField writes one field and its options.
func writeField(w *writer, p *model.Package, m *plan.Message, f *plan.Field, indent int) {
	w.comment(indent, f.Doc)
	label := ""
	switch {
	case f.Repeated:
		label = "repeated "
	case f.Optional:
		label = "optional "
	}
	var opts []string
	for _, b := range f.Behavior {
		opts = append(opts, "(google.api.field_behavior) = "+b)
	}
	if f.Ref != "" {
		opts = append(opts, fmt.Sprintf("(google.api.resource_reference) = {type: %s}", quote(f.Ref)))
	}
	if f.ChildRef != "" {
		opts = append(opts, fmt.Sprintf("(google.api.resource_reference) = {child_type: %s}", quote(f.ChildRef)))
	}
	if f.UUID {
		opts = append(opts, "(google.api.field_info).format = UUID4")
	}
	for _, v := range f.Validate {
		opts = append(opts, "(buf.validate.field)."+v)
	}
	if f.Path != "" {
		kind := f.Kind
		if kind == "" {
			kind = "KIND_ELEMENT"
		}
		opts = append(opts, fmt.Sprintf("(%s.property) = {kind: %s path: %s}", vocab, kind, quote(f.Path)))
	}
	if f.Deprecated {
		opts = append(opts, "deprecated = true")
	}
	typ, file := f.Target()
	decl := fmt.Sprintf("%s%s %s = %d", label, crossRef(p, m, typ, file), f.Name, f.Number)
	if len(opts) == 0 {
		w.line(indent, "%s;", decl)
		return
	}
	w.line(indent, "%s [", decl)
	for i, o := range opts {
		sep := ","
		if i == len(opts)-1 {
			sep = ""
		}
		w.line(indent+1, "%s%s", o, sep)
	}
	w.line(indent, "];")
}

// crossRef writes a field type so it resolves to the declaration meant,
// across subpackages of v1 as well as within one.
//
// To a sibling subpackage -- types to codes -- the reference is relative,
// codes.RoomAmenityType, which protobuf resolves from any package under v1.
// Into v1/ itself it is fully qualified. Within a package, protobuf resolves
// a reference's first component by walking outward from the referring
// message and taking the first scope with a member of that name, without
// backtracking: inside a message with a nested Rate, a reference to the
// top-level Rate -- or GuestCount.GuestCount from inside GuestCount -- binds
// wrongly. Where any enclosing scope has such a member, the reference is
// written fully qualified.
func crossRef(p *model.Package, from *plan.Message, typ, file string) string {
	if scalar[typ] || strings.HasPrefix(typ, "google.") {
		return typ
	}
	to, at := plan.SubOf(file), from.Top().Sub
	pkg := p.Proto()
	if to != "" {
		pkg += "." + to
	}
	switch {
	case to == at:
		return typeRef(pkg, from, typ, typ)
	case to == "":
		return "." + pkg + "." + typ
	}
	return typeRef(pkg, from, typ, to+"."+typ)
}

// typeRef returns rel, or the fully qualified pkg.typ when rel's first
// component is shadowed by a member of an enclosing scope.
func typeRef(pkg string, from *plan.Message, typ, rel string) string {
	first := strings.SplitN(rel, ".", 2)[0]
	for scope := from; scope != nil; scope = scope.Parent {
		for _, n := range scope.Nested {
			if n.Name == first {
				return "." + pkg + "." + typ
			}
		}
		for _, en := range scope.Enums {
			if en.Name == first {
				return "." + pkg + "." + typ
			}
		}
	}
	return rel
}

var scalar = map[string]bool{
	"string": true, "bool": true, "int32": true, "int64": true,
	"double": true, "float": true, "bytes": true,
}
