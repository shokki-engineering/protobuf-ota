// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package emit

import (
	"os"
	"path/filepath"
)

// vocab writes the one custom annotation vocabulary (CLAUDE.md rule 5),
// protobuf.opentravel.annotations.v1. It is static: what it can say does
// not depend on the specification revision.
func (e *Emitter) vocab(out string) (int, error) {
	dir := filepath.Join(out, "opentravel", "annotations", "v1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	const pkg = "protobuf.opentravel.annotations.v1"

	var a writer
	e.header(&a, pkg, "annotations.proto", map[string]bool{
		"google/protobuf/descriptor.proto":                 true,
		"protobuf/opentravel/annotations/v1/options.proto": true,
	})
	a.b.WriteString(annotationsBody)

	var o writer
	e.header(&o, pkg, "options.proto", nil)
	o.b.WriteString(optionsBody)

	for base, body := range map[string]string{"annotations.proto": a.String(), "options.proto": o.String()} {
		if err := os.WriteFile(filepath.Join(dir, base), []byte(body), 0o644); err != nil {
			return 0, err
		}
	}
	return 2, nil
}

const annotationsBody = `// annotations.proto is the single import a package needs from this
// vocabulary. It records what protobuf has no place for: which OpenTravel
// code table an enum encodes, and where each value lives in OTA's XML --
// element or attribute, and its path -- without which a codec cannot write
// the OTA message back. Both are restated in every comment for a reader.
//
// Field numbers 50000-99999 are for non-Google options. This vocabulary
// uses 54100-54103, clear of coversa-protobuf's 54000-54003 and buffers'
// 53000-53007, so all three can be loaded into one descriptor pool. See
// docs/decisions.md.

extend google.protobuf.MessageOptions {
  // element records the OTA type a message carries: an XSD type name, or
  // the element path that declares an anonymous one.
  ElementOptions element = 54100;
}

extend google.protobuf.FieldOptions {
  // property records the OTA node a field carries.
  PropertyOptions property = 54101;
}

extend google.protobuf.EnumOptions {
  // enumeration records the XSD enumeration or the OpenTravel code table an
  // enum encodes.
  EnumerationOptions enumeration = 54102;
}

extend google.protobuf.EnumValueOptions {
  // value records an enum value's spelling in OTA.
  ValueOptions value = 54103;
}
`

const optionsBody = `// ElementOptions is the payload of the element message option.
message ElementOptions {
  // The OTA path of the type, e.g. "RatePlanType", or for an anonymous type
  // the path of the element declaring it, e.g. "HotelRatePlanType/Rates".
  string path = 1;
}

// PropertyOptions is the payload of the property field option.
message PropertyOptions {
  // Whether the field carries an attribute, an element or element text.
  Kind kind = 1;

  // The OTA path of the node: a named type or attribute group, then the
  // path inside it, attributes marked "@", e.g.
  // "RatePlanGroup/@RatePlanCode". A wrapper element the schema folds away
  // is kept in the path: "HotelRatePlanType/Rates/Rate".
  string path = 2;
}

// EnumerationOptions is the payload of the enumeration enum option.
message EnumerationOptions {
  // The XSD simple type the enum encodes, e.g. "PMS_ResStatusType", or the
  // path of the attribute declaring an anonymous one.
  string xsd_type = 1;

  // The OpenTravel code table the enum encodes, by abbreviation, e.g.
  // "RMA". Each value's number is then its OTA code.
  string code_list = 2;

  // The code table's name, e.g. "Room Amenity Type".
  string table = 3;
}

// ValueOptions is the payload of the value enum value option.
message ValueOptions {
  // The value as OTA spells it: the enumeration literal, such as
  // "Request denied", or a code table's code name, such as "Hairdryer".
  string source_spelling = 1;
}

// Kind is the kind of XML node a field carries.
enum Kind {
  // Not specified.
  KIND_UNSPECIFIED = 0;

  // An XML attribute.
  KIND_ATTRIBUTE = 1;

  // A child element.
  KIND_ELEMENT = 2;

  // The element's own text content.
  KIND_TEXT = 3;
}
`
