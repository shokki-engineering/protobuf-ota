// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"fmt"
	"strconv"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
)

// requests writes messages.proto: the request and response messages of the
// package's standard methods, AIP-131 to AIP-135 and AIP-164.
func (e *Emitter) requests(p *model.Package) string {
	imports := map[string]bool{
		"google/api/field_behavior.proto": true, "google/api/resource.proto": true,
		"buf/validate/validate.proto": true, "google/protobuf/field_mask.proto": true,
		local(p, p.Message.File): true,
	}
	var w writer
	e.header(&w, p.Proto(), "messages.proto", imports)
	r, svc, coll := p.Res.Name, p.Service(), naming.Pascal(p.Plural())
	field, plural := naming.Snake(r), naming.Snake(p.Plural())
	nameField := func(verb string) {
		w.comment(1, fmt.Sprintf("Name of the %s %s, %q.", r, verb, p.Pattern()))
		w.line(1, "string name = 1 [")
		w.line(2, "(google.api.field_behavior) = REQUIRED,")
		w.line(2, "(google.api.resource_reference) = {type: %s},", quote(p.Type()))
		w.line(2, "(buf.validate.field).string.pattern = %s", strconv.Quote(model.Regexp(p.Pattern())))
		w.line(1, "];")
	}
	parent := func(n int, doc string) {
		w.comment(1, doc)
		w.line(1, "string parent = %d [", n)
		w.line(2, "(google.api.field_behavior) = REQUIRED,")
		w.line(2, "(google.api.resource_reference) = {child_type: %s}", quote(p.Type()))
		w.line(1, "];")
		w.line(0, "")
	}
	message := func(name, doc string, body func()) {
		w.comment(0, doc)
		w.line(0, "message %s {", name)
		body()
		trimBlank(&w)
		w.line(0, "}")
		w.line(0, "")
	}

	message("Get"+r+"Request", fmt.Sprintf("Request message for %s.Get%s, AIP-131 <https://aip.dev/131>.", svc, r), func() {
		nameField("to retrieve")
	})
	resourceField := func(n int, verb string) {
		w.comment(1, fmt.Sprintf("The %s to %s.", r, verb))
		w.line(1, "%s %s = %d [", r, field, n)
		w.line(2, "(google.api.field_behavior) = REQUIRED,")
		w.line(2, "(buf.validate.field).required = true")
		w.line(1, "];")
		w.line(0, "")
	}
	message("Update"+r+"Request", fmt.Sprintf("Request message for %s.Update%s, AIP-134 <https://aip.dev/134>.", svc, r), func() {
		resourceField(1, "update. Its name field identifies the resource")
		w.comment(1, "Fields to update. An empty mask updates every mutable field.\n\n"+
			"A field OTA's responding system sets is OUTPUT_ONLY, and a mask naming one is\n"+
			"rejected rather than silently ignored (CLAUDE.md rule 8).")
		w.line(1, "google.protobuf.FieldMask update_mask = 2 [(google.api.field_behavior) = OPTIONAL];")
	})
	if p.Res.Singleton {
		trimBlank(&w)
		return w.String()
	}

	message("List"+coll+"Request", fmt.Sprintf("Request message for %s.List%s, AIP-132 <https://aip.dev/132>.", svc, coll), func() {
		n := 1
		if p.Parent != nil {
			parent(1, fmt.Sprintf("The %s whose %s to list, %q.", p.Parent.Res.Name, p.Plural(), p.ParentPattern()))
			n = 2
		}
		w.comment(1, fmt.Sprintf("Maximum %s to return. Defaults to 50, capped at 1000.", p.Plural()))
		w.line(1, "int32 page_size = %d [", n)
		w.line(2, "(google.api.field_behavior) = OPTIONAL,")
		w.line(2, "(buf.validate.field).int32 = {gte: 0 lte: 1000}")
		w.line(1, "];")
		w.line(0, "")
		w.comment(1, fmt.Sprintf("Page token from a previous List%sResponse.next_page_token.", coll))
		w.line(1, "string page_token = %d [(google.api.field_behavior) = OPTIONAL];", n+1)
		w.line(0, "")
		w.comment(1, "Filter expression, AIP-160 <https://aip.dev/160>.")
		w.line(1, "string filter = %d [(google.api.field_behavior) = OPTIONAL];", n+2)
		w.line(0, "")
		w.comment(1, fmt.Sprintf("Include soft-deleted %s. AIP-164 <https://aip.dev/164>.", p.Plural()))
		w.line(1, "bool show_deleted = %d [(google.api.field_behavior) = OPTIONAL];", n+3)
	})
	message("List"+coll+"Response", fmt.Sprintf("Response message for %s.List%s, AIP-132 <https://aip.dev/132>.", svc, coll), func() {
		w.comment(1, fmt.Sprintf("The %s on this page.", p.Plural()))
		w.line(1, "repeated %s %s = 1;", r, plural)
		w.line(0, "")
		w.comment(1, "Token for the next page, empty when this is the last page.")
		w.line(1, "string next_page_token = 2;")
	})
	message("Create"+r+"Request", fmt.Sprintf("Request message for %s.Create%s, AIP-133 <https://aip.dev/133>.", svc, r), func() {
		n := 1
		if p.Parent != nil {
			parent(1, fmt.Sprintf("The %s to create the %s under, %q.", p.Parent.Res.Name, r, p.ParentPattern()))
			n = 2
		}
		resourceField(n, "create")
		idDoc := "Client-assigned id, becoming the last segment of the resource name. The\nserver generates one when empty."
		if p.Res.Identity != "" {
			idDoc = fmt.Sprintf("The id, becoming the last segment of the resource name: OTA's @%s,\nverbatim.", p.Res.Identity)
		}
		w.comment(1, idDoc)
		w.line(1, "string %s_id = %d [(google.api.field_behavior) = OPTIONAL];", field, n+1)
	})
	message("Delete"+r+"Request", fmt.Sprintf("Request message for %s.Delete%s, AIP-135 <https://aip.dev/135>.", svc, r), func() {
		nameField("to delete")
	})
	message("Undelete"+r+"Request", fmt.Sprintf("Request message for %s.Undelete%s, AIP-164 <https://aip.dev/164>.", svc, r), func() {
		nameField("to restore")
	})
	trimBlank(&w)
	return w.String()
}
