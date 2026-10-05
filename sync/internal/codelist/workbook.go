// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package codelist

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
)

// workbook is the part of a SpreadsheetML package this reader needs: sheet
// names mapped to their parts, and the shared string table cells point into.
type workbook struct {
	zip     *zip.ReadCloser
	sheets  map[string]string // sheet name -> part path
	strings []string
}

func openWorkbook(file string) (*workbook, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	w := &workbook{zip: z, sheets: map[string]string{}}

	var wb struct {
		Sheets []struct {
			Name string `xml:"name,attr"`
			RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
		} `xml:"sheets>sheet"`
	}
	var rels struct {
		Rel []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	var sst struct {
		SI []richText `xml:"si"`
	}
	for part, v := range map[string]any{
		"xl/workbook.xml": &wb, "xl/_rels/workbook.xml.rels": &rels, "xl/sharedStrings.xml": &sst,
	} {
		if err := w.decode(part, v); err != nil {
			z.Close()
			return nil, err
		}
	}
	targets := map[string]string{}
	for _, r := range rels.Rel {
		targets[r.ID] = path.Join("xl", r.Target)
	}
	for _, s := range wb.Sheets {
		w.sheets[s.Name] = targets[s.RID]
	}
	for _, si := range sst.SI {
		w.strings = append(w.strings, si.text())
	}
	return w, nil
}

func (w *workbook) Close() error { return w.zip.Close() }

// decode reads one XML part of the package into v.
func (w *workbook) decode(part string, v any) error {
	f, err := w.zip.Open(part)
	if err != nil {
		return fmt.Errorf("%s: %w", part, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", part, err)
	}
	return nil
}

// richText is a shared or inline string: plain, or a run of formatted
// pieces. The formatting -- OTA marks new codes in red -- is not data.
type richText struct {
	T string `xml:"t"`
	R []struct {
		T string `xml:"t"`
	} `xml:"r"`
}

func (r richText) text() string {
	var b strings.Builder
	b.WriteString(r.T)
	for _, run := range r.R {
		b.WriteString(run.T)
	}
	return b.String()
}

// rows reads a sheet as rows of cells keyed by column letter.
func (w *workbook) rows(sheet string) ([]map[string]string, error) {
	part, ok := w.sheets[sheet]
	if !ok {
		return nil, fmt.Errorf("workbook has no sheet %q", sheet)
	}
	var ws struct {
		Rows []struct {
			Cells []struct {
				Ref    string   `xml:"r,attr"`
				Type   string   `xml:"t,attr"`
				Value  string   `xml:"v"`
				Inline richText `xml:"is"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := w.decode(part, &ws); err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(ws.Rows))
	for _, r := range ws.Rows {
		row := map[string]string{}
		for _, c := range r.Cells {
			col := strings.TrimRight(c.Ref, "0123456789")
			switch c.Type {
			case "s":
				var i int
				if _, err := fmt.Sscan(c.Value, &i); err != nil || i >= len(w.strings) {
					return nil, fmt.Errorf("%s %s: bad shared string %q", sheet, c.Ref, c.Value)
				}
				row[col] = w.strings[i]
			case "inlineStr":
				row[col] = c.Inline.text()
			default:
				row[col] = c.Value
			}
		}
		out = append(out, row)
	}
	return out, nil
}
