// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

// Group assigns files so no two import each other.
//
// One top-level message per file is the rule, and OTA's types are
// recursive: a type can reach itself through others, and protobuf forbids
// files importing one another in a cycle. So the messages of each strongly
// connected component of the type graph share one file -- that of the
// member the package reached first -- and every other message keeps its own.
// This is the only exception to one message per file, and protobuf leaves
// no other way to declare the types at all.
func (b *Builder) Group() {
	index := map[*Message]int{}
	low := map[*Message]int{}
	onStack := map[*Message]bool{}
	var stack []*Message
	next := 0

	var visit func(m *Message)
	visit = func(m *Message) {
		index[m], low[m] = next, next
		next++
		stack = append(stack, m)
		onStack[m] = true
		for _, d := range deps(m) {
			if _, seen := index[d]; !seen {
				visit(d)
				low[m] = min(low[m], low[d])
			} else if onStack[d] {
				low[m] = min(low[m], index[d])
			}
		}
		if low[m] != index[m] {
			return
		}
		var scc []*Message
		for {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[top] = false
			scc = append(scc, top)
			if top == m {
				break
			}
		}
		if len(scc) > 1 {
			first := scc[0]
			for _, s := range scc {
				if order(b, s) < order(b, first) || s.Sub == "" && first.Sub != "" {
					first = s // a cycle with v1/ in it stays in v1/
				}
			}
			for _, s := range scc {
				s.File, s.Sub = first.File, first.Sub
			}
		}
	}
	for _, m := range b.Messages {
		if _, seen := index[m]; !seen {
			visit(m)
		}
	}
	b.liftCycles()
}

// liftCycles moves into v1/ any types message that reaches v1/. v1/ imports
// types/, so types/ importing v1/ back is a package import cycle, which buf
// refuses even where no two files form one. It happens only through an
// anonymous type planned in v1/ first and reused from types/, and moving
// the few messages that reach it is the whole repair.
func (b *Builder) liftCycles() {
	for moved := true; moved; {
		moved = false
		for _, m := range b.Messages {
			if m.Sub != SubTypes {
				continue
			}
			for _, d := range deps(m) {
				if d.Sub == "" {
					b.Root(m, "")
					moved = true
					break
				}
			}
		}
	}
}

// deps are the other top-level messages a top-level message's fields name,
// its nested messages' included.
func deps(m *Message) []*Message {
	var out []*Message
	seen := map[*Message]bool{m: true}
	var walk func(x *Message)
	walk = func(x *Message) {
		for _, f := range x.Fields {
			var t *Message
			switch {
			case f.Msg != nil:
				t = f.Msg.Top()
			case f.Enum != nil && f.Enum.Parent != nil:
				t = f.Enum.Parent.Top()
			}
			if t != nil && !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
		for _, n := range x.Nested {
			walk(n)
		}
	}
	walk(m)
	return out
}

func order(b *Builder, m *Message) int {
	for i, x := range b.Messages {
		if x == m {
			return i
		}
	}
	return len(b.Messages)
}
