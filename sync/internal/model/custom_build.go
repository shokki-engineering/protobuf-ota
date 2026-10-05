// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

// envelope are the members of an OTA message that are about the message,
// not the hotel: dropped from every custom request and response
// (docs/decisions.md, "The envelope is dropped").
var envelope = map[string]bool{
	"POS": true, "Success": true, "Warnings": true, "Errors": true,
	"TPA_Extensions": true, "UniqueID": true, "MessageID": true,
}

// buildCustoms plans the request and response messages of the package's
// custom and batch methods into the package.
func (p *Package) buildCustoms(s *xsd.Schema) error {
	p.customReq, p.customResp = map[string]*plan.Message{}, map[string]*plan.Message{}
	for i := range customs {
		c := &customs[i]
		if c.Resource != p.Res.Name {
			continue
		}
		p.Customs = append(p.Customs, c)
		req := &plan.Message{Name: c.Name + "Request",
			Doc: fmt.Sprintf("Request message for %s.%s.", p.Service(), c.Name)}
		req.Fields = p.leadFields(c)
		if c.Request != "" {
			if err := p.fillFrom(s, req, c.Request, c.Skip); err != nil {
				return err
			}
		}
		p.Plan.Top(req)
		// A request and its response are meaningless apart: they share one
		// file in v1/, named for the method.
		file := naming.Snake(c.Name) + ".proto"
		p.Plan.Root(req, file)
		p.customReq[c.Name] = req

		resp := &plan.Message{Name: c.Name + "Response",
			Doc: fmt.Sprintf("Response message for %s.%s.", p.Service(), c.Name)}
		switch {
		case c.Response != "":
			if err := p.fillFrom(s, resp, c.Response, nil); err != nil {
				return err
			}
			if c.Paginated {
				if err := paginate(req, resp, c.Results); err != nil {
					return fmt.Errorf("%s: %w", c.Name, err)
				}
			}
		case c.ResponseOne:
			resp.Fields = []*plan.Field{{Name: naming.Snake(p.Res.Name), Number: 1, Msg: p.Message,
				Doc: fmt.Sprintf("The %s.", p.Res.Name)}}
		case c.ResponseList || c.Kind == KindBatchGet || c.Kind == KindBatchUpdate:
			resp.Fields = []*plan.Field{{Name: naming.Snake(p.Plural()), Number: 1, Msg: p.Message, Repeated: true,
				Doc: fmt.Sprintf("The %s.", p.Plural())}}
		default:
			continue // returns the resource
		}
		p.Plan.Top(resp)
		p.Plan.Root(resp, file)
		p.customResp[c.Name] = resp
	}
	return nil
}

// leadFields are a custom request's AIP fields: the resource's name, or its
// parent, and what the kind of method requires.
func (p *Package) leadFields(c *Custom) []*plan.Field {
	var out []*plan.Field
	add := func(f *plan.Field) { f.Number = len(out) + 1; out = append(out, f) }
	switch {
	case c.Kind == KindResource:
		add(&plan.Field{Name: "name", Type: "string", Behavior: []string{"REQUIRED"}, Ref: p.Type(),
			Doc: fmt.Sprintf("The %s, %q.", p.Res.Name, p.Pattern())})
	case p.Parent != nil:
		add(&plan.Field{Name: "parent", Type: "string", Behavior: []string{"REQUIRED"}, ChildRef: p.Type(),
			Doc: fmt.Sprintf("The %s the %s belong to, %q.", p.Parent.Res.Name, p.Plural(), p.ParentPattern())})
	}
	switch c.Kind {
	case KindBatchGet:
		add(&plan.Field{Name: "names", Type: "string", Repeated: true, Behavior: []string{"REQUIRED"}, Ref: p.Type(),
			Doc:      fmt.Sprintf("The %s to retrieve, at most 1000.", p.Plural()),
			Validate: []string{"repeated = {min_items: 1 max_items: 1000}"}})
	case KindBatchUpdate:
		r := p.Res.Name
		add(&plan.Field{Name: "requests", Type: "Update" + r + "Request", TypeFile: "messages.proto",
			Repeated: true, Behavior: []string{"REQUIRED"},
			Doc:      fmt.Sprintf("One update per %s, at most 1000. OTA's date ranges and weekday masks\nare expanded into one per day.", naming.Humanise(r)),
			Validate: []string{"repeated = {min_items: 1 max_items: 1000}"}})
	}
	if c.Body {
		add(&plan.Field{Name: naming.Snake(p.Res.Name), Msg: p.Message, Behavior: []string{"REQUIRED"},
			Doc: fmt.Sprintf("The %s to price.", p.Res.Name)})
	}
	return out
}

// fillFrom plans an OTA message's content, less its envelope, into m.
func (p *Package) fillFrom(s *xsd.Schema, m *plan.Message, message string, skip []string) error {
	e, err := s.Element(message)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Res.Name, err)
	}
	if e.Complex == nil {
		return fmt.Errorf("%s: %s has simple content", p.Res.Name, message)
	}
	p.Plan.Fill(m, e.Complex, func(key string) bool {
		if strings.HasPrefix(key, "OTA_PayloadStdAttributes/") {
			return true
		}
		name := key[strings.LastIndex(key, "/")+1:]
		for _, k := range skip {
			if k == name {
				return true
			}
		}
		return envelope[name]
	})
	return nil
}

// paginate adds AIP-158 paging to a custom method: page_size and page_token
// on the request, next_page_token on the response, and the paged repeated
// field moved to the response's front, where AIP-158 wants it.
func paginate(req, resp *plan.Message, results string) error {
	next := func(m *plan.Message) int {
		n := 0
		for _, f := range m.Fields {
			n = max(n, f.Number)
		}
		return n + 1
	}
	req.Fields = append(req.Fields,
		&plan.Field{Name: "page_size", Number: next(req), Type: "int32", Behavior: []string{"OPTIONAL"},
			Doc: "Maximum results to return. Defaults to 50, capped at 1000.", Validate: []string{"int32 = {gte: 0 lte: 1000}"}},
	)
	req.Fields = append(req.Fields,
		&plan.Field{Name: "page_token", Number: next(req), Type: "string", Behavior: []string{"OPTIONAL"},
			Doc: "Page token from a previous response's next_page_token."})
	at := -1
	for i, f := range resp.Fields {
		if f.Name == results && f.Repeated {
			at = i
		}
	}
	if at < 0 {
		return fmt.Errorf("no repeated field %q to page", results)
	}
	lead := resp.Fields[at]
	resp.Fields = append([]*plan.Field{lead}, append(resp.Fields[:at:at], resp.Fields[at+1:]...)...)
	resp.Fields = append(resp.Fields, &plan.Field{Name: "next_page_token", Type: "string",
		Doc: "Token for the next page, empty when this is the last page."})
	// AIP-158 wants the paged field to be field 1, not merely first. These
	// messages are new, so numbering them in order breaks nothing.
	for i, f := range resp.Fields {
		f.Number = i + 1
	}
	return nil
}
