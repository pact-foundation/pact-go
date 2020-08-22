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

// S is the string primitive wrapper (alias) for the Matcher type,
// it allows plain strings to be matched
type I32 int32

func (s I32) isMatcher()   {}
func (s I32) isV3Matcher() {}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (s I32) GetValue() interface{} {
	return int(s)
}

func (s I32) Type() MatcherClass {
	return likeMatcher
}

func (s I32) MatchingRule() rule {
	return rule{
		"match": "type",
	}
}
