package v3

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type eachLike struct {
	Contents interface{} `json:"contents"`
	Min      int         `json:"min,omitempty"`
}

func (m eachLike) GetValue() interface{} {
	return m.Contents
}

func (m eachLike) Type() MatcherClass {
	return arrayMinLikeMatcher
}

func (m eachLike) MatchingRule() rule {
	r := rule{
		"match": "type",
	}
	if m.Min == 0 {
		r["min"] = 1
	}

	return r
}

type like struct {
	Contents interface{} `json:"contents"`
}

func (m like) GetValue() interface{} {
	return m.Contents
}

func (m like) Type() MatcherClass {
	return likeMatcher
}

func (m like) MatchingRule() rule {
	return rule{
		"match": "type",
	}
}

type term struct {
	Data termData `json:"data"`
}

func (m term) GetValue() interface{} {
	return m.Data.Generate
}

func (m term) Type() MatcherClass {
	return regexMatcher
}

func (m term) MatchingRule() rule {
	return rule{
		"match": "regex",
		"regex": m.Data.Matcher.Regex,
	}
}

// TODO: revisit these attributes and marshalling after refactor
type termData struct {
	Generate interface{} `json:"generate"`
	Matcher  termMatcher `json:"matcher"`
}

type termMatcher struct {
	Regex interface{} `json:"s"`
}

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, min int) MatcherV2 {
	return eachLike{
		Contents: content,
		Min:      min,
	}
}

var ArrayMinLike = EachLike

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) MatcherV2 {
	return like{
		Contents: content,
	}
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) MatcherV2 {
	return term{
		Data: termData{
			Generate: generate,
			Matcher: termMatcher{
				Regex: matcher,
			},
		},
	}
}

// HexValue defines a matcher that accepts hexidecimal values.
func HexValue() MatcherV2 {
	return Regex("3F", hexadecimalRegex)
}

// Identifier defines a matcher that accepts number values.
func Identifier() MatcherV2 {
	return Like(42)
}

// IPAddress defines a matcher that accepts valid IPv4 addresses.
func IPAddress() MatcherV2 {
	return Regex("127.0.0.1", ipAddressRegex)
}

// IPv4Address matches valid IPv4 addresses.
var IPv4Address = IPAddress

// IPv6Address defines a matcher that accepts IP addresses.
func IPv6Address() MatcherV2 {
	return Regex("::ffff:192.0.2.128", ipAddressRegex)
}

// Timestamp matches a pattern corresponding to the ISO_DATETIME_FORMAT, which
// is "yyyy-MM-dd'T'HH:mm:ss". The current date and time is used as the eaxmple.
func Timestamp() MatcherV2 {
	return Regex(timeExample.Format(time.RFC3339), timestampRegex)
}

// Date matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "yyyy-MM-dd". The current date is used as the eaxmple.
func Date() MatcherV2 {
	return Regex(timeExample.Format("2006-01-02"), dateRegex)
}

// Time matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "'T'HH:mm:ss". The current tem is used as the eaxmple.
func Time() MatcherV2 {
	return Regex(timeExample.Format("T15:04:05"), timeRegex)
}

// UUID defines a matcher that accepts UUIDs. Produces a v4 UUID as the example.
func UUID() MatcherV2 {
	return Regex("fc763eba-0905-41c5-a27f-3934ab26786c", uuidRegex)
}

// Regex is a more appropriately named alias for the "Term" matcher
var Regex = Term

// MatcherV2 allows various implementations such String or StructMatcher
// to be provided in when matching with the DSL
// We use the strategy outlined at http://www.jerf.org/iri/post/2917
// to create a "sum" or "union" type.
type MatcherV2 interface {
	// GetValue returns the raw generated value for the matcher
	// without any of the matching detail context
	GetValue() interface{}

	Type() MatcherClass

	// Generate the matching rule for this Matcher
	MatchingRule() rule
}

// S is the string primitive wrapper (alias) for the Matcher type,
// it allows plain strings to be matched
type S string

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (s S) GetValue() interface{} {
	return string(s)
}

func (s S) Type() MatcherClass {
	return likeMatcher
}

func (s S) MatchingRule() rule {
	return rule{
		"match": "type",
	}
}

// String is the longer named form of the string primitive wrapper,
// it allows plain strings to be matched
type String = S

// StructMatcher matches a complex object structure, which may itself
// contain nested Matchers
type StructMatcher map[string]MatcherV2

// type StructMatcher map[string]Matcher

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (m StructMatcher) GetValue() interface{} {
	return m
}

func (s StructMatcher) Type() MatcherClass {
	return structTypeMatcher
}

func (s StructMatcher) MatchingRule() rule {
	return rule{
		"match": "type",
	}
}

// MapMatcher allows a map[string]string-like object
// to also contain complex matchers
// TODO: bring back this type?
// type MapMatcher map[string]Matcher
type MapMatcher map[string]interface{}
type HeadersMatcher = MapMatcher

// QueryMatcher matches a query string interface
type QueryMatcher map[string][]interface{}

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

// Match recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
// By default, it requires slices to have a minimum of 1 element.
// For concrete types, it uses `dsl.Like` to assert that types match.
// Optionally, you may override these defaults by supplying custom
// pact tags on your structs.
//
// Supported Tag Formats
// Minimum Slice Size: `pact:"min=2"`
// String RegEx:       `pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
// TODO: support generators
func Match(src interface{}) MatcherV2 {
	return match(reflect.TypeOf(src), getDefaults())
}

// match recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
func match(srcType reflect.Type, params params) MatcherV2 {
	switch kind := srcType.Kind(); kind {
	case reflect.Ptr:
		return match(srcType.Elem(), params)
	case reflect.Slice, reflect.Array:
		return EachLike(match(srcType.Elem(), getDefaults()), params.slice.min)
	case reflect.Struct:
		result := StructMatcher{}

		for i := 0; i < srcType.NumField(); i++ {
			field := srcType.Field(i)
			result[field.Tag.Get("json")] = match(field.Type, pluckParams(field.Type, field.Tag.Get("pact")))
		}
		return result
	case reflect.String:
		if params.str.regEx != "" {
			return Term(params.str.example, params.str.regEx)
		}
		if params.str.example != "" {
			return Like(params.str.example)
		}

		return Like("string")
	case reflect.Bool:
		if params.boolean.defined {
			return Like(params.boolean.value)
		}
		return Like(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if params.number.integer != 0 {
			return Like(params.number.integer)
		}
		return Like(1)
	case reflect.Float32, reflect.Float64:
		if params.number.float != 0 {
			return Like(params.number.float)
		}
		return Like(1.1)
	default:
		panic(fmt.Sprintf("match: unhandled type: %v", srcType))
	}
}

// params are plucked from 'pact' struct tags as match() traverses
// struct fields. They are passed back into match() along with their
// associated type to serve as parameters for the dsl functions.
type params struct {
	slice   sliceParams
	str     stringParams
	number  numberParams
	boolean boolParams
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

// pluckParams converts a 'pact' tag into a pactParams struct
// Supported Tag Formats
// Minimum Slice Size: `pact:"min=2"`
// String RegEx:       `pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
func pluckParams(srcType reflect.Type, pactTag string) params {
	params := getDefaults()
	if pactTag == "" {
		return params
	}

	switch kind := srcType.Kind(); kind {
	case reflect.Bool:
		if _, err := fmt.Sscanf(pactTag, "example=%t", &params.boolean.value); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}
		params.boolean.defined = true
	case reflect.Float32, reflect.Float64:
		if _, err := fmt.Sscanf(pactTag, "example=%g", &params.number.float); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if _, err := fmt.Sscanf(pactTag, "example=%d", &params.number.integer); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}
	case reflect.Slice:
		if _, err := fmt.Sscanf(pactTag, "min=%d", &params.slice.min); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}
	case reflect.String:
		fullRegex, _ := regexp.Compile(`regex=(.*)$`)
		exampleRegex, _ := regexp.Compile(`^example=(.*)`)

		if fullRegex.Match([]byte(pactTag)) {
			components := strings.Split(pactTag, ",regex=")

			if len(components[1]) == 0 {
				triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: regex must not be empty"))
			}

			if _, err := fmt.Sscanf(components[0], "example=%s", &params.str.example); err != nil {
				triggerInvalidPactTagPanic(pactTag, err)
			}
			params.str.regEx = components[1]

		} else if exampleRegex.Match([]byte(pactTag)) {
			components := strings.Split(pactTag, "example=")

			if len(components) != 2 || strings.TrimSpace(components[1]) == "" {
				triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: example must not be empty"))
			}

			params.str.example = components[1]
		}
	}

	return params
}

func triggerInvalidPactTagPanic(tag string, err error) {
	panic(fmt.Sprintf("match: encountered invalid pact tag %q . . . parsing failed with error: %v", tag, err))
}

// Generators
// https://github.com/pact-foundation/pact-specification/tree/version-3#introduce-example-generators
