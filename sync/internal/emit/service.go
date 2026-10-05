// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"fmt"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// service writes service.proto from the package's RPCs: the standard set
// for its shape (CLAUDE.md rule 7), then the custom and batch methods the
// methods table assigns it. model.Package.RPCs is the one list, so the
// reference docs cannot describe a method the schema does not have.
func (e *Emitter) service(p *model.Package) string {
	rpcs := p.RPCs()
	imports := map[string]bool{
		"google/api/annotations.proto": true, "google/api/client.proto": true,
		local(p, p.Message.File): true, local(p, "messages.proto"): true,
	}
	for _, r := range rpcs {
		if r.Response == "google.protobuf.Empty" {
			imports["google/protobuf/empty.proto"] = true
		}
		for _, f := range []string{fileOf(r.ReqMsg), fileOf(r.RespMsg)} {
			if f != "" {
				imports[local(p, f)] = true
			}
		}
	}
	var w writer
	e.header(&w, p.Proto(), "service.proto", imports)

	doc := fmt.Sprintf("%s manages the %s resource, %q.", p.Service(), p.Res.Name, p.Pattern())
	if p.Res.Singleton {
		doc += "\n\nA singleton, AIP-156 <https://aip.dev/156>: it exists exactly when its parent does, so it has Get and Update and nothing else of its own."
	} else {
		doc += "\n\nA collection, with the AIP-131..135 standard methods plus AIP-164 <https://aip.dev/164> undelete."
	}
	w.comment(0, doc)
	w.line(0, "service %s {", p.Service())
	for _, r := range rpcs {
		w.comment(1, r.Doc)
		w.line(1, "rpc %s(%s) returns (%s) {", r.Name, r.Request, r.Response)
		if r.Body == "" {
			w.line(2, "option (google.api.http) = {%s: %s};", r.Verb, quote(r.Path))
		} else {
			w.line(2, "option (google.api.http) = {")
			w.line(3, "%s: %s", r.Verb, quote(r.Path))
			w.line(3, "body: %s", quote(r.Body))
			w.line(2, "};")
		}
		if r.Signature != "" {
			w.line(2, "option (google.api.method_signature) = %s;", quote(r.Signature))
		}
		w.line(1, "}")
		w.line(0, "")
	}
	trimBlank(&w)
	w.line(0, "}")
	return w.String()
}

// fileOf is the file a planned message is written to, or "" for none.
func fileOf(m *plan.Message) string {
	if m == nil {
		return ""
	}
	return m.Top().File
}
