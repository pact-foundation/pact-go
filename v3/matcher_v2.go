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

// Term Matcher regexes
const (
	hexadecimal = `[0-9a-fA-F]+`
	ipAddress   = `(\d{1,3}\.)+\d{1,3}`
	ipv6Address = `(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,6}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,5}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,4}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,3}\Z)|(\A([0-9a-f]{1,4}:){1,5}(:[0-9a-f]{1,4}){1,2}\Z)|(\A([0-9a-f]{1,4}:){1,6}(:[0-9a-f]{1,4}){1,1}\Z)|(\A(([0-9a-f]{1,4}:){1,7}|:):\Z)|(\A:(:[0-9a-f]{1,4}){1,7}\Z)|(\A((([0-9a-f]{1,4}:){6})(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A(([0-9a-f]{1,4}:){5}[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A([0-9a-f]{1,4}:){5}:[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,3}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,2}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,1}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A(([0-9a-f]{1,4}:){1,5}|:):(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A:(:[0-9a-f]{1,4}){1,5}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)`
	uuid        = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	timestamp   = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))([T\s]((([01]\d|2[0-3])((:?)[0-5]\d)?|24\:?00)([\.,]\d+(?!:))?)?(\17[0-5]\d([\.,]\d+)?)?([zZ]|([\+-])([01]\d|2[0-3]):?([0-5]\d)?)?)?)?$`
	date        = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))?)`
	timeRegex   = `^(T\d\d:\d\d(:\d\d)?(\.\d+)?(([+-]\d\d:\d\d)|Z)?)?$`
)

var timeExample = time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)

type eachLike struct {
	Contents interface{} `json:"contents"`
	Min      int         `json:"min,omitempty"`
	Max      int         `json:"max,omitempty"`
}

func (m eachLike) GetValue() interface{} {
	return m.Contents
}

func (m eachLike) isV2Matcher() {
}

func (m eachLike) Type() MatcherClass {
	if m.Max != 0 {
		return arrayMaxLikeMatcher
	}
	return arrayMinLikeMatcher
}

func (m eachLike) MatchingRule() rule {
	matcher := rule{
		"match": "type",
	}

	if m.Max != 0 {
		matcher["max"] = m.Max
	} else {
		matcher["min"] = m.Min
	}

	return matcher
}

type like struct {
	Contents interface{} `json:"contents"`
}

func (m like) GetValue() interface{} {
	return m.Contents
}

func (m like) isV2Matcher() {
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

func (m term) isV2Matcher() {
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
func EachLike(content interface{}, min int) Matcher {
	return eachLike{
		Contents: content,
		Min:      min,
	}
}

var ArrayMinLike = EachLike

// ArrayMaxLike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no greater than the provided max.
func ArrayMaxLike(content interface{}, max int) Matcher {
	return eachLike{
		Contents: content,
		Max:      max,
	}
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) Matcher {
	return like{
		Contents: content,
	}
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) Matcher {
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
func HexValue() Matcher {
	return Regex("3F", hexadecimal)
}

// Identifier defines a matcher that accepts integer values.
func Identifier() Matcher {
	return Like(42)
}

// Integer defines a matcher that accepts ints. Identical to Identifier.
var Integer = Identifier

// IPAddress defines a matcher that accepts valid IPv4 addresses.
func IPAddress() Matcher {
	return Regex("127.0.0.1", ipAddress)
}

// IPv4Address matches valid IPv4 addresses.
var IPv4Address = IPAddress

// IPv6Address defines a matcher that accepts IP addresses.
func IPv6Address() Matcher {
	return Regex("::ffff:192.0.2.128", ipAddress)
}

// Decimal defines a matcher that accepts any decimal value.
func Decimal() Matcher {
	return Like(42.0)
}

// Timestamp matches a pattern corresponding to the ISO_DATETIME_FORMAT, which
// is "yyyy-MM-dd'T'HH:mm:ss". The current date and time is used as the eaxmple.
func Timestamp() Matcher {
	return Regex(timeExample.Format(time.RFC3339), timestamp)
}

// Date matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "yyyy-MM-dd". The current date is used as the eaxmple.
func Date() Matcher {
	return Regex(timeExample.Format("2006-01-02"), date)
}

// Time matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "'T'HH:mm:ss". The current tem is used as the eaxmple.
func Time() Matcher {
	return Regex(timeExample.Format("T15:04:05"), timeRegex)
}

// UUID defines a matcher that accepts UUIDs. Produces a v4 UUID as the example.
func UUID() Matcher {
	return Regex("fc763eba-0905-41c5-a27f-3934ab26786c", uuid)
}

// Regex is a more appropriately named alias for the "Term" matcher
var Regex = Term

// Matcher allows various implementations such String or StructMatcher
// to be provided in when matching with the DSL
// We use the strategy outlined at http://www.jerf.org/iri/post/2917
// to create a "sum" or "union" type.
type Matcher interface {
	// isV2Matcher is how we tell the compiler that strings
	// and other types are the same / allowed
	isV2Matcher()

	// GetValue returns the raw generated value for the matcher
	// without any of the matching detail context
	GetValue() interface{}

	Type() MatcherClass

	// Generate the matching rule for this Matcher
	MatchingRule() rule
}

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
	arrayMaxLikeMatcher = "arrayMaxLikeMatcher"

	// Matches map[string]interface{} types is basically a container for other matchers
	structTypeMatcher = "structTypeMatcher"
)

// S is the string primitive wrapper (alias) for the Matcher type,
// it allows plain strings to be matched
type S string

func (s S) isV2Matcher() {}

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
type StructMatcher map[string]Matcher

// type StructMatcher map[string]Matcher

func (m StructMatcher) isV2Matcher() {}

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
func Match(src interface{}) Matcher {
	return match(reflect.TypeOf(src), getDefaults())
}

// match recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
func match(srcType reflect.Type, params params) Matcher {
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
