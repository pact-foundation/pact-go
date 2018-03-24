package dsl

import (
	"encoding/json"
	"log"
)

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, minRequired int) Matcher {
	return Matcher{
		"json_class": "Pact::ArrayLike",
		"contents":   toObject(content),
		"min":        minRequired,
	}
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) Matcher {
	return Matcher{
		"json_class": "Pact::SomethingLike",
		"contents":   toObject(content),
	}
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) Matcher {
	return Matcher{
		"json_class": "Pact::Term",
		"data": map[string]interface{}{
			"generate": toObject(generate),
			"matcher": map[string]interface{}{
				"json_class": "Regexp",
				"o":          0,
				"s":          toObject(matcher),
			},
		},
	}
}

// Regex is a more appropriately named alias for the "Term" matcher
var Regex = Term

// StringMatcher allows a string or Matcher to be provided in
// when matching with the DSL
// We use the strategy outlined at http://www.jerf.org/iri/post/2917
// to create a "sum" or "union" type.
type StringMatcher interface {
	// isMatcher is how we tell the compiler that strings
	// and other types are the same / allowed
	isMatcher()
}

// S is the string primitive wrapper (alias) for the StringMatcher type,
// it allows plain strings to be matched
type S string

func (s S) isMatcher() {}

// String is the longer named form of the string primitive wrapper,
// it allows plain strings to be matched
type String string

func (s String) isMatcher() {}

// Matcher matches a complex object structure, which may itself
// contain nested Matchers
type Matcher map[string]interface{}

func (m Matcher) isMatcher() {}

// MarshalJSON is a custom encoder for Header type
func (m Matcher) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{}

	for header, value := range m {
		obj[header] = toObject(value)
	}

	return json.Marshal(obj)
}

// UnmarshalJSON is a custom decoder for Header type
func (m *Matcher) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}

// MapMatcher allows a map[string]string-like object
// to also contain complex matchers
type MapMatcher map[string]StringMatcher

// MarshalJSON is a custom encoder for Header type
func (h MapMatcher) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{}

	for header, value := range h {
		obj[header] = toObject(value)
	}

	return json.Marshal(obj)
}

// UnmarshalJSON is a custom decoder for Header type
func (h *MapMatcher) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil
}

// Takes an object and converts it to a JSON representation
func objectToString(obj interface{}) string {
	switch content := obj.(type) {
	case string:
		return content
	default:
		jsonString, err := json.Marshal(obj)
		if err != nil {
			log.Println("[DEBUG] objectToString: error unmarshaling object into string:", err.Error())
			return ""
		}
		return string(jsonString)
	}
}
