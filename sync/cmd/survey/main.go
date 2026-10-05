// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Command survey reports what the generator reads, and writes nothing.
//
//	survey -tree OTA_HotelRatePlanNotifRQ/RatePlans/RatePlan
//
// prints the resolved type under an element path, the way the reader sees
// it: every attribute and element with its catalogue key.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/spec"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

func main() {
	tree := flag.String("tree", "", "print the resolved type under this element path")
	depth := flag.Int("depth", 3, "how deep -tree descends")
	why := flag.String("why", "", "with -tree: print the shortest path from it to this named type")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	pin, err := spec.Load(filepath.Join(*root, "sync", "spec.yaml"))
	check(err)
	s, err := xsd.Load(pin.SuiteDir(filepath.Join(*root, "modules", "opentravel")))
	check(err)
	if *tree != "" {
		e, err := s.Find(*tree)
		check(err)
		if *why != "" {
			fmt.Println(shortest(e, *why))
			return
		}
		printElem(e, 0, *depth, map[*xsd.Complex]bool{})
	}
}

func printElem(e *xsd.Elem, d, max int, seen map[*xsd.Complex]bool) {
	pad := strings.Repeat("  ", d)
	occ := fmt.Sprintf("[%d..%s]", e.Min, maxStr(e.Max))
	if e.Simple != nil {
		fmt.Printf("%s- %s %s : %s  {%s}\n", pad, e.Name, occ, simpleStr(e.Simple), e.Key)
		return
	}
	c := e.Complex
	fmt.Printf("%s- %s %s : %s  {%s}%s\n", pad, e.Name, occ, nameOr(c.Name, "anon"), c.Key, choice(e))
	if seen[c] || d >= max {
		return
	}
	seen[c] = true
	defer delete(seen, c)
	for _, a := range c.Attrs {
		req := ""
		if a.Required {
			req = " required"
		}
		fmt.Printf("%s    @%s : %s%s  {%s}\n", pad, a.Name, simpleStr(a.Simple), req, a.Key)
	}
	if c.Text != nil {
		fmt.Printf("%s    #text : %s\n", pad, simpleStr(c.Text))
	}
	for _, k := range c.Elems {
		printElem(k, d+1, max, seen)
	}
}

func choice(e *xsd.Elem) string {
	if e.Choice == 0 {
		return ""
	}
	return fmt.Sprintf(" choice#%d oneof=%v", e.Choice, e.OneOf)
}

func simpleStr(s *xsd.Simple) string {
	switch {
	case s.Union != nil:
		return nameOr(s.Name, "union") + "(union)"
	case s.List != nil:
		return nameOr(s.Name, "list") + "(list)"
	case s.IsEnum():
		return fmt.Sprintf("%s(enum %d)", nameOr(s.Name, s.Base), len(s.Enums))
	}
	return nameOr(s.Name, "xs:"+s.Base)
}

func nameOr(n, alt string) string {
	if n == "" {
		return alt
	}
	return n
}

func maxStr(m int) string {
	if m == xsd.Unbounded {
		return "n"
	}
	return fmt.Sprint(m)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "survey:", err)
		os.Exit(1)
	}
}

// shortest is a breadth-first search from e to the first element whose
// type is named target, returned as a path of element names.
func shortest(e *xsd.Elem, target string) string {
	type step struct {
		c    *xsd.Complex
		path string
	}
	queue := []step{{e.Complex, e.Name}}
	seen := map[*xsd.Complex]bool{e.Complex: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, k := range cur.c.Elems {
			if k.Complex == nil || seen[k.Complex] {
				continue
			}
			p := cur.path + "/" + k.Name
			if k.Complex.Name == target {
				return p
			}
			seen[k.Complex] = true
			queue = append(queue, step{k.Complex, p})
		}
	}
	return "unreachable"
}
