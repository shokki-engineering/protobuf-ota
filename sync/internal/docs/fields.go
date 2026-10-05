// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package docs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/plan"
)

// fieldsSection tables the resource message's fields: the OTA ones first,
// then the AIP identity and lifecycle fields every resource carries.
func (g *Generator) fieldsSection(p *model.Package) string {
	var ota, aip []*plan.Field
	for _, f := range p.Message.Fields {
		if f.Path == "" {
			aip = append(aip, f)
		} else {
			ota = append(ota, f)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## Fields\n\n%d fields: %d AIP identity and lifecycle, %d from OTA. Field numbers\n"+
		"%d-%d are reserved, so a later identity field never renumbers an OTA one.\n\n",
		len(p.Message.Fields), len(aip), len(ota), p.Message.Reserved[0], p.Message.Reserved[1])
	b.WriteString("| Field | Type | Behavior | OTA | Description |\n| --- | --- | --- | --- | --- |\n")
	for _, f := range ota {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | `%s` | %s |\n",
			f.Name, typeOf(f), behavior(f), f.Path, cell(firstSentence(f.Doc)))
	}
	b.WriteString("\n<details>\n<summary>Identity and lifecycle -- the fields every resource carries</summary>\n\n")
	b.WriteString("| Field | Type | Behavior | Description |\n| --- | --- | --- | --- |\n")
	for _, f := range aip {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | %s |\n", f.Name, typeOf(f), behavior(f), cell(firstSentence(f.Doc)))
	}
	b.WriteString("\n</details>\n\n")
	return b.String()
}

func typeOf(f *plan.Field) string {
	t, _ := f.Target()
	if f.Repeated {
		return "repeated " + t
	}
	return t
}

func behavior(f *plan.Field) string {
	if len(f.Behavior) == 0 {
		return ""
	}
	return "`" + strings.Join(f.Behavior, "` `") + "`"
}

// codeTables lists the OpenTravel code tables the package's enums encode.
func codeTables(p *model.Package) string {
	var rows []string
	for _, e := range p.Plan.Enums {
		if e.CodeList != "" {
			rows = append(rows, fmt.Sprintf("| `%s` | %s | `%s` | %d |", e.CodeList, e.Table, e.Name, len(e.Values)))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	sort.Strings(rows)
	return "## Code tables\n\nEnums numbered by their OTA code, so the wire carries exactly what OTA's XML\n" +
		"does ([decisions](" + toRoot(p) + "docs/decisions.md)):\n\n| Table | Name | Enum | Codes |\n| --- | --- | --- | --- |\n" +
		strings.Join(rows, "\n") + "\n"
}

// example renders the resource as JSON: its AIP fields, and up to six of
// its scalar OTA fields with a placeholder of the right shape.
func example(p *model.Package) string {
	var lines []string
	add := func(name, value string) { lines = append(lines, fmt.Sprintf("  %q: %s", jsonName(name), value)) }
	add("name", fmt.Sprintf("%q", examplePattern(p.Pattern())))
	add("uid", `"b3f1c2de-4a5b-4c6d-8e9f-0a1b2c3d4e5f"`)
	add("create_time", `"2026-10-05T10:15:30Z"`)
	add("update_time", `"2026-10-05T10:15:30Z"`)
	shown := 0
	for _, f := range p.Message.Fields {
		if f.Path == "" || f.Repeated || f.OneOf != "" || shown == 6 {
			continue
		}
		if v, ok := placeholder(f); ok {
			add(f.Name, v)
			shown++
		}
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n}"
}

func jsonName(snake string) string { return naming.LowerCamel(snake) }

func examplePattern(pattern string) string {
	segs := strings.Split(pattern, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, "{") {
			segs[i] = strings.ToUpper(strings.ReplaceAll(strings.Trim(s, "{}"), "_", "")[:3]) + "1"
		}
	}
	return strings.Join(segs, "/")
}

// placeholder is an example value for a scalar or enum field.
func placeholder(f *plan.Field) (string, bool) {
	switch {
	case f.Enum != nil && len(f.Enum.Values) > 0:
		return fmt.Sprintf("%q", f.Enum.Values[0].Name), true
	case f.Msg != nil:
		return "", false
	}
	switch f.Type {
	case "string":
		return `"` + "<" + strings.ReplaceAll(f.Name, "_", " ") + ">" + `"`, true
	case "bool":
		return "true", true
	case "int32", "int64":
		return "1", true
	case "google.protobuf.Timestamp":
		return `"2026-10-05T10:15:30Z"`, true
	}
	return "", false
}
