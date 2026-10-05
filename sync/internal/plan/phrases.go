// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"regexp"
	"strings"

	"github.com/shokki-engineering/protobuf-ota/sync/internal/naming"
)

// phrases are AIP-140 rewrites OTA needs often enough to apply by rule:
// "unit of measure" is a unit, and "number of nights" is a night count.
var phrases = []struct {
	re   *regexp.Regexp
	with string
}{
	{regexp.MustCompile(`(^|_)unit_of_measure(_|$)`), "${1}unit${2}"},
	{regexp.MustCompile(`(^|_)(?:nmbr|nbr|number)_of_([a-z0-9]+?)s?(_|$)`), "${1}${2}_count${3}"},
	{regexp.MustCompile(`(^|_)days?_of_weeks?(_|$)`), "${1}weekday${2}"},
	{regexp.MustCompile(`(^|_)(check|walk|move|opt)_in(_|$)`), "${1}${2}in${3}"},
	{regexp.MustCompile(`(^|_)in_house(_|$)`), "${1}inhouse${2}"},
	{regexp.MustCompile(`(^|_)source_of_business(_|$)`), "${1}business_source${2}"},
	{regexp.MustCompile(`(^|_)number_in_party(_|$)`), "${1}party_size${2}"},
	{regexp.MustCompile(`(^|_)set_for_quantity(_|$)`), "${1}setup_quantity${2}"},
	{regexp.MustCompile(`^percent_of_([a-z0-9_]+)$`), "${1}_percent"},
	{regexp.MustCompile(`(^|_)([a-z0-9]+)_before_tax(_|$)`), "${1}pre_tax_${2}${3}"},
	{regexp.MustCompile(`(^|_)([a-z0-9]+)_after_tax(_|$)`), "${1}post_tax_${2}${3}"},
	{regexp.MustCompile(`(^|_)([a-z0-9]+)_including_markup(_|$)`), "${1}marked_up_${2}${3}"},
	{regexp.MustCompile(`(^|_)type_of_([a-z0-9]+)(_|$)`), "${1}${2}_type${3}"},
	{regexp.MustCompile(`(^|_)lengths?_of_stays?(_|$)`), "${1}stay_lengths${2}"},
	{regexp.MustCompile(`(^|_)balance_of_stay(_|$)`), "${1}stay_balance${2}"},
}

func rephrase(name string) string {
	for _, p := range phrases {
		name = p.re.ReplaceAllString(name, p.with)
	}
	return name
}

// anyPlural reports whether a name already has a plural word in it, as
// OTA's wrapper names do -- GuaranteesAccepted, LengthsOfStay -- so that
// pluralising its last word would double the plural.
func anyPlural(words []string) bool {
	for _, w := range words {
		if naming.Pluralise(w) == w && strings.HasSuffix(w, "s") {
			return true
		}
	}
	return false
}
