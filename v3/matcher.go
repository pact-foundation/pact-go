package v3

import (
	"time"
)

// Term Matcher regexes
const (
	hexadecimalRegex = `[0-9a-fA-F]+`
	ipAddressRegex   = `(\d{1,3}\.)+\d{1,3}`
	ipv6AddressRegex = `(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,6}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,5}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,4}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,3}\Z)|(\A([0-9a-f]{1,4}:){1,5}(:[0-9a-f]{1,4}){1,2}\Z)|(\A([0-9a-f]{1,4}:){1,6}(:[0-9a-f]{1,4}){1,1}\Z)|(\A(([0-9a-f]{1,4}:){1,7}|:):\Z)|(\A:(:[0-9a-f]{1,4}){1,7}\Z)|(\A((([0-9a-f]{1,4}:){6})(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A(([0-9a-f]{1,4}:){5}[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A([0-9a-f]{1,4}:){5}:[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,3}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,2}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,1}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A(([0-9a-f]{1,4}:){1,5}|:):(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A:(:[0-9a-f]{1,4}){1,5}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)`
	uuidRegex        = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	timestampRegex   = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))([T\s]((([01]\d|2[0-3])((:?)[0-5]\d)?|24\:?00)([\.,]\d+(?!:))?)?(\17[0-5]\d([\.,]\d+)?)?([zZ]|([\+-])([01]\d|2[0-3]):?([0-5]\d)?)?)?)?$`
	dateRegex        = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))?)`
	timeRegex        = `^(T\d\d:\d\d(:\d\d)?(\.\d+)?(([+-]\d\d:\d\d)|Z)?)?$`
)

var timeExample = time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)

// MatcherClass is used to differentiate the various matchers when serialising
type MatcherClass string

// Matcher Types used to discriminate when serialising the rules
const (
	// likeMatcher is the ID for the Like Matcher
	likeMatcher MatcherClass = "likeMatcher"

	// regexMatcher is the ID for the Term Matcher
	regexMatcher = "regexMatcher"

	// arrayMinLikeMatcher is the ID for the ArrayMinLike Matcher
	arrayMinLikeMatcher = "arrayMinLikeMatcher"

	// arrayMaxLikeMatcher is the ID for the arrayMaxLikeMatcher Matcher
	// arrayMaxLikeMatcher = "arrayMaxLikeMatcher"

	// arrayMinMaxLikeMatcher sets lower and upper bounds on the array size
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-a-minmax-type-matcher
	arrayMinMaxLikeMatcher = "arrayMinMaxLikeMatcher"

	// Matches map[string]interface{} types is basically a container for other matchers
	structTypeMatcher = "structTypeMatcher"

	// Matches integers
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-more-specific-type-matchers
	integerMatcher = "intMatcher"

	// Matches decimals
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-more-specific-type-matchers
	decimalMatcher = "decimalMatcher"

	// Matches nulls
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-more-specific-type-matchers
	nullMatcher = "nullMatcher"

	// Equality matcher
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
	equalityMatcher = "equalityMatcher"

	// includes matcher
	// https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-include-matcher
	includesMatcher = "includesMatcher"

	// string generator
	// https://github.com/pact-foundation/pact-specification/tree/version-3#introduce-example-generators
	stringGeneratorMatcher = "stringGeneratorMatcher"
)

// MatcherClass is used to differentiate the various matchers when serialising
type generatorType string

// Matcher Types used to discriminate when serialising the rules
const (
	dateTimeGenerator generatorType = "DateTime"
	dateGenerator                   = "Date"
	timeGenerator                   = "Time"
)

// params are plucked from 'pact' struct tags as matchV2() traverses
// struct fields. They are passed back into matchV2() along with their
// associated type to serve as parameters for the dsl functions.
type params struct {
	slice     sliceParams
	str       stringParams
	number    numberParams
	boolean   boolParams
	generator MatcherV3
}

type numberParams struct {
	integer int
	float   float32
}
type boolParams struct {
	value   bool
	defined bool
}

type sliceParams struct {
	min int
	max int
}

type stringParams struct {
	example string
	regEx   string
}

// getDefaults returns the default params
func getDefaults() params {
	return params{
		slice: sliceParams{
			min: 1,
		},
	}
}
