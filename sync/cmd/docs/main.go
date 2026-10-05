// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Command docs renders the generated schema as Markdown reference
// documentation: one README.md beside each package's v1/ directory, and an
// index at protobuf/opentravel/README.md.
//
// It reads the same publications and builds the same model the sync command
// does, rather than parsing the emitted .proto files back in: which methods a
// resource's shape gets, which code tables a package carries, are decisions
// the generator made, and recovering them from its output would derive them a
// second time.
//
// Usage:
//
//	docs                write READMEs under protobuf/
//	docs -root DIR      read the pin and checkout from another repository root
//	docs -out DIR       write somewhere else
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/codelist"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/docs"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/ledger"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/model"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/spec"
	"github.com/shokki-engineering/protobuf-ota/sync/internal/xsd"
)

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "", "directory to write into (default <root>/protobuf)")
	flag.Parse()
	if *out == "" {
		*out = filepath.Join(*root, "protobuf")
	}

	pin, err := spec.Load(filepath.Join(*root, "sync", "spec.yaml"))
	check(err)
	checkout := filepath.Join(*root, "modules", "opentravel")
	s, err := xsd.Load(pin.SuiteDir(checkout))
	check(err)
	codes, err := codelist.Load(pin.CodeListFile(checkout))
	check(err)
	// The ledger is read, never saved: the docs describe the values sync
	// emitted, and only sync adds to codes.lock.
	lock, err := ledger.Load(filepath.Join(*root, "codes.lock"))
	check(err)
	m, err := model.Build(s, codes, lock)
	check(err)
	n, err := docs.New(m, pin).Generate(*out)
	check(err)
	fmt.Printf("docs: OTA %s, code list %s -> %d files\n", pin.SuiteLabel, pin.CodeListLabel, n)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "docs:", err)
		os.Exit(1)
	}
}
