// sync turns the OpenTravel 2017B XML message suite and the OpenTravel code
// list into protobuf.
//
// No dependencies, and the require block should stay empty. XSD is XML and
// the code list workbook is a zip of XML; the standard library reads both.
// See docs/generator.md.
module github.com/shokki-engineering/protobuf-ota/sync

go 1.26
