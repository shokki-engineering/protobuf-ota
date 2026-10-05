// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

// valueTypes are the messages generated into each package that needs them,
// in place of google.type (CLAUDE.md rule 4). Each has google.type's shape,
// so a consumer converting between the two has nothing to map.
var valueTypes = map[string]func() *Message{
	"Money": func() *Message {
		return &Message{Name: "Money", Doc: "An amount of money in a currency, after google.type.Money.\n\n" +
			"OTA's Amount, CurrencyCode and DecimalPlaces become this: DecimalPlaces scales an\n" +
			"integral Amount (8500 with DecimalPlaces 2 is 85.00), so it is applied on the way\n" +
			"in and not kept. See docs/conventions.md rule 13.",
			Fields: []*Field{
				valueField("currency_code", 1, "string", "The ISO 4217 currency code, e.g. \"EUR\".",
					`string = {pattern: "^[A-Z]{3}$"}`, "ignore = IGNORE_IF_ZERO_VALUE"),
				valueField("units", 2, "int64", "The whole units of the amount."),
				valueField("nano_units", 3, "int32", "Nano units of the amount, with the same sign as units. google.type.Money\ncalls this nanos, which AIP-142 reads as a time.",
					"int32 = {gte: -999999999 lte: 999999999}"),
			}}
	},
	"Date": func() *Message {
		return &Message{Name: "Date", Doc: "A calendar date with no time zone, after google.type.Date.\n\n" +
			"A stay date is a date at the hotel; a Timestamp would invent an instant and a\n" +
			"zone OTA never sent. See docs/decisions.md.",
			Fields: []*Field{
				valueField("year", 1, "int32", "Year, 1 to 9999.", "int32 = {gte: 1 lte: 9999}"),
				valueField("month", 2, "int32", "Month of the year, 1 to 12.", "int32 = {gte: 1 lte: 12}"),
				valueField("day", 3, "int32", "Day of the month, 1 to 31.", "int32 = {gte: 1 lte: 31}"),
			}}
	},
	"TimeOfDay": func() *Message {
		return &Message{Name: "TimeOfDay", Doc: "A time of day with no date or time zone, after google.type.TimeOfDay.\n\n" +
			"Its fields are singular -- hour, not hours -- because AIP-142 reads seconds and\n" +
			"nanos as claims that a field is a time.",
			Fields: []*Field{
				valueField("hour", 1, "int32", "Hour of the day, 0 to 23.", "int32 = {gte: 0 lte: 23}"),
				valueField("minute", 2, "int32", "Minute of the hour, 0 to 59.", "int32 = {gte: 0 lte: 59}"),
				valueField("second", 3, "int32", "Second of the minute, 0 to 59.", "int32 = {gte: 0 lte: 59}"),
				valueField("nanosecond", 4, "int32", "Fraction of the second, in nanoseconds.", "int32 = {gte: 0 lte: 999999999}"),
			}}
	},
	"Decimal": func() *Message {
		return &Message{Name: "Decimal", Doc: "An exact decimal number, after google.type.Decimal.\n\n" +
			"OTA's xs:decimal values -- percentages, rates, measurements -- are carried as their\n" +
			"decimal string so no precision is lost to binary floating point.",
			Fields: []*Field{
				valueField("value", 1, "string", "The decimal value, e.g. \"12.5\" or \"-0.01\".",
					`string = {pattern: "^[+-]?(?:[0-9]+(?:\\.[0-9]*)?|\\.[0-9]+)$"}`, "ignore = IGNORE_IF_ZERO_VALUE"),
			}}
	},
}

func valueField(name string, n int, typ, doc string, validate ...string) *Field {
	return &Field{Name: name, Number: n, Type: typ, Doc: doc, Behavior: []string{"OPTIONAL"}, Validate: validate}
}

// value returns the package's generated value type, creating it on first
// use.
func (b *Builder) value(name string) *Message {
	if m, ok := b.values[name]; ok {
		return m
	}
	m := valueTypes[name]()
	b.values[name] = m
	b.Top(m)
	return m
}
