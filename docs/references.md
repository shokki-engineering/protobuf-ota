# References

A checked link index. `scripts/check-links.sh` will verify every URL here
resolves -- on change, and weekly, because a link rots without anyone touching
the repository. Until it exists, check a link before adding it.

The check exists because `CLAUDE.md` points at this file instead of working
from memory, and AIP rule semantics have been guessed wrong more than once. A
dead entry sends the reader back to guessing.

## The source model

- OpenTravel Alliance — https://opentravel.org
- OpenTravel specifications, as published — https://opentravel.org/specifications/
- OpenTravel-Specifications, the archive this repository pins — https://github.com/OpenTravel/OpenTravel-Specifications
- The 2017B message suite, at the pinned commit — https://github.com/OpenTravel/OpenTravel-Specifications/tree/28dcc880acf2d22a8a5e93d20174971080b6846c/OTA1.0/OTA2017B/OpenTravel_2017B_XML
- The 2024-09-04 code list, at the pinned commit — https://github.com/OpenTravel/OpenTravel-Specifications/tree/28dcc880acf2d22a8a5e93d20174971080b6846c/CodeLists/OTA_CodeLists_2024_09_04
- XML Schema Part 1: Structures (the XSD the suite is written in) — https://www.w3.org/TR/xmlschema-1/
- XML Schema Part 2: Datatypes (`xs:date`, `xs:duration`, facets) — https://www.w3.org/TR/xmlschema-2/

Read, but not generated from -- see [`decisions.md`](decisions.md):

- The OTA 2.0 2020A object suite, at the pinned commit — https://github.com/OpenTravel/OpenTravel-Specifications/tree/28dcc880acf2d22a8a5e93d20174971080b6846c/OTA2.0/OTA2.0_2020A_ObjectSuite

Implementations of the hotel messages, consulted for how OTA is used in
practice rather than how it is written:

- AlpineBits, an OTA profile for hotel distribution — https://www.alpinebits.org/
- Booking.com connectivity API, OTA-based — https://developers.booking.com/connectivity/docs

## The projects this follows

- coversa-protobuf, whose conventions this repository takes — https://github.com/oh-tarnished/coversa-protobuf
- protobuf-rfc, which coversa adapts — https://github.com/the-protobuf-project/protobuf-rfc

## AIP

The rules cited in the schema, and the linter page for each. For exact rule
semantics, the installed api-linter's source under
`rules/` of [api-linter 2.3.1](https://github.com/googleapis/api-linter/tree/v2.3.1/rules), the version CI installs, is the
authority: the reserved-word, preposition and abbreviation lists are there,
and the pages below do not always print them.

| AIP | Subject | Linter |
| --- | --- | --- |
| https://aip.dev/121 | Resource-oriented design | https://linter.aip.dev/121/ |
| https://aip.dev/122 | Resource names | https://linter.aip.dev/122/ |
| https://aip.dev/123 | Resource types | https://linter.aip.dev/123/ |
| https://aip.dev/126 | Enumerations | https://linter.aip.dev/126/ |
| https://aip.dev/131 | Get | https://linter.aip.dev/131/ |
| https://aip.dev/132 | List | https://linter.aip.dev/132/ |
| https://aip.dev/133 | Create | https://linter.aip.dev/133/ |
| https://aip.dev/134 | Update | https://linter.aip.dev/134/ |
| https://aip.dev/135 | Delete | https://linter.aip.dev/135/ |
| https://aip.dev/136 | Custom methods | https://linter.aip.dev/136/ |
| https://aip.dev/140 | Field names | https://linter.aip.dev/140/ |
| https://aip.dev/141 | Quantities | https://linter.aip.dev/141/ |
| https://aip.dev/142 | Time and duration | https://linter.aip.dev/142/ |
| https://aip.dev/143 | Standardized codes | https://linter.aip.dev/143/ |
| https://aip.dev/148 | Standard fields | https://linter.aip.dev/148/ |
| https://aip.dev/149 | Unset field values | *(no linter rules)* |
| https://aip.dev/154 | Resource freshness (etag) | https://linter.aip.dev/154/ |
| https://aip.dev/155 | Request identification | https://linter.aip.dev/155/ |
| https://aip.dev/156 | Singleton resources | https://linter.aip.dev/156/ |
| https://aip.dev/158 | Pagination | https://linter.aip.dev/158/ |
| https://aip.dev/160 | Filtering | *(no linter rules)* |
| https://aip.dev/164 | Soft delete and undelete | https://linter.aip.dev/164/ |
| https://aip.dev/191 | File and directory structure | https://linter.aip.dev/191/ |
| https://aip.dev/192 | Documentation | https://linter.aip.dev/192/ |
| https://aip.dev/203 | Field behavior | https://linter.aip.dev/203/ |
| https://aip.dev/215 | API-specific protos | https://linter.aip.dev/215/ |
| https://aip.dev/216 | States | https://linter.aip.dev/216/ |
| https://aip.dev/231 | Batch get | https://linter.aip.dev/231/ |
| https://aip.dev/234 | Batch update | https://linter.aip.dev/234/ |

- The linter itself — https://github.com/googleapis/api-linter

The VS Code Google API Linter extension reads `workspace.protobuf.yaml` rather
than the CLI's arguments, so it may report differently from `just aip`. The
workflow is the authority.

## Tooling

- buf — https://buf.build/docs
- buf lint rules — https://buf.build/docs/lint/rules/
- managed mode — https://buf.build/docs/generate/managed-mode/
- protovalidate — https://github.com/bufbuild/protovalidate
- protovalidate rule reference — https://buf.build/bufbuild/protovalidate/docs/main:buf.validate
- just, the task runner — https://github.com/casey/just

## Protobuf itself

- Language guide (proto3) — https://protobuf.dev/programming-guides/proto3/
- Field presence, and proto3 `optional` — https://protobuf.dev/programming-guides/field_presence/
- Well-known types — https://protobuf.dev/reference/protobuf/google.protobuf/
- `google.type`, whose `Money`, `Date` and `Decimal` shapes the generated value types follow — https://github.com/googleapis/googleapis/tree/master/google/type
- Go generated code reference — https://protobuf.dev/reference/go/go-generated/
