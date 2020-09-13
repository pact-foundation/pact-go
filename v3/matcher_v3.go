package v3

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

// MatcherV3 denotes a V3 specific Matcher
type MatcherV3 interface {
	MatcherV2

	// denote a v3 matcher
	isV3Matcher()
}

type generator interface {
	Generator() rule
}

// Integer defines a matcher that accepts any integer value.
type Integer int

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

type stringGenerator struct {
	contents  string
	generator generatorType
	format    string
}

func (s stringGenerator) GetValue() interface{} {
	return s.contents
}

func (s stringGenerator) isV3Matcher() {}

func (s stringGenerator) Type() MatcherClass {
	return stringGeneratorMatcher
}

func (s stringGenerator) Generator() rule {
	r := rule{
		"type": s.generator,
	}
	if s.format != "" {
		r["format"] = s.format
	}

	return r
}
func (s stringGenerator) MatchingRule() rule {
	return nil
}

// DateGenerated matches a cross platform formatted date, and generates a current date during verification
// String example value must match the provided date format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func DateGenerated(example string, format string) MatcherV3 {
	return stringGenerator{
		contents:  example,
		generator: dateGenerator,
		format:    format,
	}
}

// TimeGenerated matches a cross platform formatted date, and generates a current time during verification
// String example value must match the provided time format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func TimeGenerated(example string, format string) MatcherV3 {
	return stringGenerator{
		contents:  example,
		generator: timeGenerator,
		format:    format,
	}
}

// DateTimeGenerated matches a cross platform formatted datetime, and generates a current datetime during verification
// String example value must match the provided datetime format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func DateTimeGenerated(example string, format string) MatcherV3 {
	return stringGenerator{
		contents:  example,
		generator: dateTimeGenerator,
		format:    format,
	}
}

// StructMatcherV3 matches a complex object structure, which may itself
// contain nested Matchers
type StructMatcherV3 = StructMatcher

func (m StructMatcherV3) isV3Matcher() {}

// MatchV3 recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
// By default, it requires slices to have a minimum of 1 element.
// For concrete types, it uses `dsl.Like` to assert that types match.
// Optionally, you may override these defaults by supplying custom
// pact tags on your structs.
//
// Supported Tag Formats
// Minimum Slice Size: `pact:"min=2"`
// Maximum Slice Size: `pact:"min=2"`
// String RegEx:       `pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
func MatchV3(src interface{}) MatcherV3 {
	return matchV3(reflect.TypeOf(src), getDefaults())
}

// Make v3 wrappers of v3 matchers
type likeV3Matcher = like

func (t likeV3Matcher) isV3Matcher() {}

type termV3Matcher = term

func (t termV3Matcher) isV3Matcher() {}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func likeV3(content interface{}) MatcherV3 {
	return likeV3Matcher{
		Contents: content,
	}
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func termV3(generate string, matcher string) MatcherV3 {
	return termV3Matcher{
		Data: termData{
			Generate: generate,
			Matcher: termMatcher{
				Regex: matcher,
			},
		},
	}
}

// match recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
func matchV3(srcType reflect.Type, params params) MatcherV3 {
	switch kind := srcType.Kind(); kind {
	case reflect.Ptr:
		return matchV3(srcType.Elem(), params)
	case reflect.Slice, reflect.Array:
		return ArrayMinMaxLike(matchV3(srcType.Elem(), getDefaults()), params.slice.min, params.slice.max)
	case reflect.Struct:
		result := StructMatcherV3{}

		for i := 0; i < srcType.NumField(); i++ {
			field := srcType.Field(i)
			result[field.Tag.Get("json")] = matchV3(field.Type, pluckParamsV3(field.Type, field.Tag.Get("pact")))
		}
		return result
	case reflect.String:
		if params.generator != nil {
			return params.generator
		}

		if params.str.regEx != "" {
			return termV3(params.str.example, params.str.regEx)
		}
		if params.str.example != "" {
			return likeV3(params.str.example)
		}

		return likeV3("string")
	case reflect.Bool:
		if params.boolean.defined {
			return likeV3(params.boolean.value)
		}
		return likeV3(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if params.number.integer != 0 {
			return Integer(params.number.integer)
		}
		return Integer(1)
	case reflect.Float32, reflect.Float64:
		if params.number.float != 0 {
			return Decimal(params.number.float)
		}
		return Decimal(1.1)
	default:
		panic(fmt.Sprintf("matchV3: unhandled type: %v", srcType))
	}
}

func pluckStringParams(pactTag string, p params) params {
	// Valid struct tag formats
	// example=2012-01-01:10:00
	// example=2012-01-01:10:00,regex=...
	// example=2012-01-01:10:00,generator=datetime,format=yyyy-MM-dd:HH:mm

	withGenerator, _ := regexp.Compile(`generator=(.*)`)
	withRegex, _ := regexp.Compile(`regex=(.*)$`)
	exampleOnly, _ := regexp.Compile(`^example=(.*)`)

	if withGenerator.Match([]byte(pactTag)) {
		components := strings.Split(pactTag, ",generator=")

		userDefinedGenerator := components[1]
		if userDefinedGenerator == "" {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: generator must not be empty"))
		}

		components = strings.Split(components[0], ",format=")
		format := components[1]

		if format == "" {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: generator format must not be empty"))
		}

		switch userDefinedGenerator {
		case "datetime":
			p.generator = stringGenerator{
				generator: dateTimeGenerator,
				format:    format,
			}
		case "date":
			p.generator = stringGenerator{
				generator: dateGenerator,
				format:    format,
			}
		case "time":
			p.generator = stringGenerator{
				generator: timeGenerator,
				format:    format,
			}
		default:
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: unknown generator %s", userDefinedGenerator))
		}

		return pluckStringParams(components[0], p)
	} else if withRegex.Match([]byte(pactTag)) {
		components := strings.Split(pactTag, ",regex=")

		if len(components[1]) == 0 {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: regex must not be empty"))
		}

		p.str.regEx = components[1]
		return pluckStringParams(components[0], p)
	} else if exampleOnly.Match([]byte(pactTag)) {
		components := strings.Split(pactTag, "example=")

		if len(components) != 2 || strings.TrimSpace(components[1]) == "" {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: example must not be empty"))
		}

		p.str.example = components[1]

		// This only allows a string generator at the moment
		if p.generator != nil {
			gen := p.generator.(stringGenerator)
			gen.contents = components[1]
			p.generator = gen
		}

		return pluckStringParams(components[0], p)
	}

	if p.str.example == "" {
		triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: example must not be empty"))
	}

	return p
}

// pluckParamsV3 converts a 'pact' tag into a pactParams struct
// Supported Tag Formats
// Minimum Slice Size: `pact:"min=2"`
// String RegEx:       `pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
func pluckParamsV3(srcType reflect.Type, pactTag string) params {
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
		if _, err := fmt.Sscanf(pactTag, "min=%d,max=%d", &params.slice.min, &params.slice.max); err != nil {
			// max is optional, but min is required
			if _, err := fmt.Sscanf(pactTag, "min=%d", &params.slice.min); err != nil {
				triggerInvalidPactTagPanic(pactTag, err)
			}
		}
	case reflect.String:
		params = pluckStringParams(pactTag, params)
		log.Println("[DEBUG] STRING PARAMS", params)

		// TODO: checks on what was parsed
	}

	return params
}
