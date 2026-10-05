// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

// Package spec reads sync/spec.yaml, the pin naming which OpenTravel
// publications the schema is generated from.
//
// The file is flat `key: value` YAML and is read as exactly that. A key this
// reader does not know, or one it needs and does not find, is an error: a
// misspelt pin would otherwise stamp a revision into every banner that was
// never read.
package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pin is the revision of OpenTravel-Specifications the schema is read from.
type Pin struct {
	Repository    string
	Commit        string
	Version       string // the commit's date, YYYY-MM-DD
	Suite         string // path of the XML message suite inside the checkout
	SuiteLabel    string // OpenTravel's name for it, e.g. 2017B
	CodeList      string // path of the code list workbook
	CodeListLabel string // e.g. 2024-09-04
}

// Load reads the pin at path.
func Load(path string) (*Pin, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Pin{}
	fields := map[string]*string{
		"repository": &p.Repository, "commit": &p.Commit, "version": &p.Version,
		"suite": &p.Suite, "suite_label": &p.SuiteLabel,
		"codelist": &p.CodeList, "codelist_label": &p.CodeListLabel,
	}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		dst, known := fields[strings.TrimSpace(k)]
		if !ok || !known {
			return nil, fmt.Errorf("%s:%d: unknown line %q", path, n, line)
		}
		*dst = strings.TrimSpace(v)
	}
	for k, v := range fields {
		if *v == "" {
			return nil, fmt.Errorf("%s: %s is not set", path, k)
		}
	}
	return p, sc.Err()
}

// SuiteDir is the message suite's directory under the checkout root.
func (p *Pin) SuiteDir(root string) string { return filepath.Join(root, p.Suite) }

// CodeListFile is the code list workbook under the checkout root.
func (p *Pin) CodeListFile(root string) string { return filepath.Join(root, p.CodeList) }

// Short is the abbreviated commit stamped into banners.
func (p *Pin) Short() string { return p.Commit[:7] }
