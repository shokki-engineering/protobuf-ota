// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Command sync regenerates protobuf/ from the pinned OpenTravel
// publications. Nothing under protobuf/ is written by hand (CLAUDE.md
// rule 2); this is what writes it.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/codelist"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/emit"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/ledger"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/spec"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	pin, err := spec.Load(filepath.Join(*root, "sync", "spec.yaml"))
	check(err)
	checkout := filepath.Join(*root, "modules", "opentravel")
	s, err := xsd.Load(pin.SuiteDir(checkout))
	check(err)
	codes, err := codelist.Load(pin.CodeListFile(checkout))
	check(err)
	lock, err := ledger.Load(filepath.Join(*root, "codes.lock"))
	check(err)
	m, err := model.Build(s, codes, lock)
	check(err)
	n, err := emit.New(m, pin).Generate(filepath.Join(*root, "protobuf"))
	check(err)
	check(lock.Save())
	fmt.Printf("sync: wrote %d files in %d packages\n", n, len(m.Packages)+1)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "sync:", err)
		os.Exit(1)
	}
}
