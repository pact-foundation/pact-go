package v3

type v3Matcher interface {
	isV3Matcher()
}

// MatcherV3 denotes a V3 specific Matcher
type MatcherV3 interface {
	MatcherV2

	// denote a v3 matcher
	isV3Matcher()
}

// Integer defines a matcher that accepts any integer value.
type Integer int

func (i Integer) isMatcher()   {}
func (i Integer) isV3Matcher() {}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (i Integer) GetValue() interface{} {
	return int(i)
}

func (i Integer) Type() MatcherClass {
	return integerMatcher
}

func (i Integer) MatchingRule() rule {
	return rule{
		"match": "integer",
	}
}

// Decimal is a matcher that accepts a decimal type
type Decimal float64

func (d Decimal) GetValue() interface{} {
	return float64(d)
}

func (d Decimal) isMatcher()   {}
func (d Decimal) isV3Matcher() {}

func (d Decimal) Type() MatcherClass {
	return decimalMatcher
}

func (d Decimal) MatchingRule() rule {
	return rule{
		"match": "decimal",
	}
}

// Null is a matcher that only accepts nulls
type Null float64

func (n Null) GetValue() interface{} {
	return float64(n)
}

func (n Null) isMatcher()   {}
func (n Null) isV3Matcher() {}

func (n Null) Type() MatcherClass {
	return nullMatcher
}

func (n Null) MatchingRule() rule {
	return rule{
		"match": "null",
	}
}

// equality resets matching cascades back to equality
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
type equality struct {
	contents interface{}
}

func (e equality) GetValue() interface{} {
	return e.contents
}

func (e equality) isMatcher()   {}
func (e equality) isV3Matcher() {}

func (e equality) Type() MatcherClass {
	return equalityMatcher
}

func (e equality) MatchingRule() rule {
	return rule{
		"match": "equality",
	}
}

// Equality resets matching cascades back to equality
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
func Equality(content interface{}) MatcherV3 {
	return equality{
		contents: content,
	}
}

// Includes checks if the given string is contained by the actual value
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
type Includes string

func (i Includes) GetValue() interface{} {
	return string(i)
}
func (i Includes) isMatcher()   {}
func (i Includes) isV3Matcher() {}
func (i Includes) Type() MatcherClass {
	return includesMatcher
}

func (i Includes) MatchingRule() rule {
	return rule{
		"match": "include",
		"value": string(i),
	}
}

type minMaxLike struct {
	Contents interface{} `json:"contents"`
	Min      int         `json:"min,omitempty"`
	Max      int         `json:"max,omitempty"` // NOTE: only used for V3
}

func (m minMaxLike) GetValue() interface{} {
	return m.Contents
}

func (m minMaxLike) isMatcher()   {}
func (m minMaxLike) isV3Matcher() {}

func (m minMaxLike) Type() MatcherClass {
	return arrayMinMaxLikeMatcher
}

func (m minMaxLike) MatchingRule() rule {
	r := rule{
		"match": "type",
	}
	if m.Min == 0 {
		r["min"] = 1
	}
	if m.Max != 0 {
		r["max"] = m.Max
	}

	return r
}

// ArrayMinMaxLike is like EachLike except has a bounds on the max and the min
// https://github.com/pact-foundation/pact-specification/tree/version-3#add-a-minmax-type-matcher
func ArrayMinMaxLike(content interface{}, min int, max int) MatcherV3 {
	return minMaxLike{
		Contents: content,
		Min:      min,
		Max:      max,
	}
}

// ArrayMaxLike is like EachLike except has a bounds on the max
// https://github.com/pact-foundation/pact-specification/tree/version-3#add-a-minmax-type-matcher
func ArrayMaxLike(content interface{}, max int) MatcherV3 {
	return minMaxLike{
		Contents: content,
		Min:      1,
		Max:      max,
	}
}
