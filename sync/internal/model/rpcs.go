// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// RPC is one method of a package's service, as both emit and docs need it.
type RPC struct {
	Name, Request, Response string
	// ReqMsg and RespMsg are set when the message is planned into its own
	// file, as a custom method's is; otherwise it is in messages.proto, or
	// is the resource.
	ReqMsg, RespMsg *plan.Message
	Verb            string // get, post, patch, delete
	Path, Body      string
	Signature       string // "" for none
	Doc             string
}

// RPCs is the package's methods: the standard set for its shape (CLAUDE.md
// rule 7), then its custom and batch methods, in the customs table's order.
func (p *Package) RPCs() []RPC {
	r, field, coll := p.Res.Name, naming.Snake(p.Res.Name), naming.Pascal(p.Plural())
	name := "/v1/{name=" + Glob(p.Pattern()) + "}"
	out := []RPC{
		{Name: "Get" + r, Request: "Get" + r + "Request", Response: r, Verb: "get", Path: name,
			Signature: "name", Doc: fmt.Sprintf("Retrieves one %s.", r)},
		{Name: "Update" + r, Request: "Update" + r + "Request", Response: r, Verb: "patch",
			Path: "/v1/{" + field + ".name=" + Glob(p.Pattern()) + "}", Body: field,
			Signature: field + ",update_mask", Doc: fmt.Sprintf("Updates a %s. Only the fields in update_mask are written.", r)},
	}
	if p.Res.Singleton {
		return append(out, p.customRPCs()...)
	}
	sig := field + "," + field + "_id"
	listSig := ""
	if p.Parent != nil {
		sig, listSig = "parent,"+sig, "parent"
	}
	out = append(out,
		RPC{Name: "List" + coll, Request: "List" + coll + "Request", Response: "List" + coll + "Response",
			Verb: "get", Path: p.collectionPath(), Signature: listSig,
			Doc: fmt.Sprintf("Lists %s, paginated per AIP-158 <https://aip.dev/158>.", p.Plural())},
		RPC{Name: "Create" + r, Request: "Create" + r + "Request", Response: r, Verb: "post",
			Path: p.collectionPath(), Body: field, Signature: sig, Doc: fmt.Sprintf("Creates a %s.", r)},
		RPC{Name: "Delete" + r, Request: "Delete" + r + "Request", Response: "google.protobuf.Empty",
			Verb: "delete", Path: name, Signature: "name",
			Doc: fmt.Sprintf("Soft-deletes a %s; it is purged at expire_time unless restored.", r)},
		RPC{Name: "Undelete" + r, Request: "Undelete" + r + "Request", Response: r, Verb: "post",
			Path: strings.TrimSuffix(name, "}") + "}:undelete", Body: "*", Signature: "name",
			Doc: fmt.Sprintf("Restores a soft-deleted %s before it is purged.", r)},
	)
	return append(out, p.customRPCs()...)
}

// collectionPath is the collection's URI: /v1/hotels, or under its parent.
func (p *Package) collectionPath() string {
	if p.Parent == nil {
		return "/v1/" + p.Plural()
	}
	return "/v1/{parent=" + Glob(p.ParentPattern()) + "}/" + p.Plural()
}

// customRPCs are the package's custom and batch methods.
func (p *Package) customRPCs() []RPC {
	var out []RPC
	for _, c := range p.Customs {
		rpc := RPC{Name: c.Name, Request: c.Name + "Request", Response: p.Res.Name, Doc: c.Doc,
			ReqMsg: p.customReq[c.Name], RespMsg: p.customResp[c.Name], Verb: "post", Body: "*"}
		if rpc.RespMsg != nil {
			rpc.Response = rpc.RespMsg.Name
		}
		switch c.Kind {
		case KindResource:
			rpc.Path, rpc.Signature = "/v1/{name="+Glob(p.Pattern())+"}:"+c.Verb, "name"
		case KindCollection:
			rpc.Path = p.collectionPath() + ":" + c.Verb
		case KindBatchGet:
			rpc.Verb, rpc.Body, rpc.Path = "get", "", p.collectionPath()+":batchGet"
		case KindBatchUpdate:
			rpc.Path = p.collectionPath() + ":batchUpdate"
		}
		if p.Parent != nil && c.Kind != KindResource {
			rpc.Signature = "parent"
		}
		if rpc.ReqMsg != nil {
			// A method signature names every required field (AIP-4232).
			var req []string
			for _, f := range rpc.ReqMsg.Fields {
				if len(f.Behavior) > 0 && f.Behavior[0] == "REQUIRED" && f.OneOf == "" {
					req = append(req, f.Name)
				}
			}
			rpc.Signature = strings.Join(req, ",")
		}
		out = append(out, rpc)
	}
	return out
}

// Glob turns a resource pattern into an HTTP path template:
// hotels/{hotel}/ratePlans/{rate_plan} gives hotels/*/ratePlans/*.
func Glob(pattern string) string {
	segs := strings.Split(pattern, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, "{") {
			segs[i] = "*"
		}
	}
	return strings.Join(segs, "/")
}
