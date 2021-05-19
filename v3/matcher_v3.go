package v3

import (
	"encoding/json"
	"log"
)

// matcherV3 denotes a V3 specific Matcher
type matcherV3 interface {
	// denote a v3 matcher
	isV3Matcher()
}

// Decimal defines a matcher that accepts any decimal value.
func Decimal(example float64) Matcher {
	return like{
		Type:  "decimal",
		Value: example,
	}
}

// Integer defines a matcher that accepts any integer value.
func Integer(example int) Matcher {
	return like{
		Type:  "integer",
		Value: example,
	}
}

// Null is a matcher that only accepts nulls
type Null struct{}

func (n Null) GetValue() interface{} {
	return nil
}

func (n Null) isV3Matcher() {}
func (n Null) isMatcher()   {}

func (n Null) MarshalJSON() ([]byte, error) {
	type marshaler Null

	return json.Marshal(struct {
		Type string `json:"pact:matcher:type"`
		marshaler
	}{"null", marshaler(n)})
}

// equality resets matching cascades back to equality
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
type equality struct {
	Type     string      `json:"pact:matcher:type"`
	Contents interface{} `json:"value"`
}

func (e equality) GetValue() interface{} {
	return e.Contents
}

func (e equality) isV3Matcher() {}
func (e equality) isMatcher()   {}

// Equality resets matching cascades back to equality
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
func Equality(content interface{}) Matcher {
	return equality{
		Contents: content,
		Type:     "equality",
	}
}

// Includes checks if the given string is contained by the actual value
// see https://github.com/pact-foundation/pact-specification/tree/version-3#add-an-equality-matcher
type includes struct {
	Type     string `json:"pact:matcher:type"`
	Contents string `json:"value"`
}

func (i includes) GetValue() interface{} {
	return i.Contents
}

func (i includes) isV3Matcher() {}
func (i includes) isMatcher()   {}

func Includes(content string) Matcher {
	return includes{
		Type:     "include",
		Contents: content,
	}
}

type fromProviderState struct {
	Type       string `json:"pact:matcher:type"`
	Generator  string `json:"pact:generator:type"`
	Expression string `json:"expression"`
	Value      string `json:"value"`
}

func (s fromProviderState) GetValue() interface{} {
	return s.Value
}

func (s fromProviderState) isV3Matcher() {}
func (s fromProviderState) isMatcher()   {}

// Marks a item as to be injected from the provider state
//
// "expression" is used to lookup the dynamic value from the provider state context
// during verification
// "example" is the example value to used in the consumer test
func FromProviderState(expression, example string) Matcher {
	return fromProviderState{
		Type:       "type",
		Generator:  "ProviderState",
		Expression: expression,
		Value:      example,
	}
}

type eachKeyLike struct {
	Type     string      `json:"pact:matcher:type"`
	Contents interface{} `json:"value"`
}

func (e eachKeyLike) GetValue() interface{} {
	return e.Contents
}

func (e eachKeyLike) isV3Matcher() {}
func (e eachKeyLike) isMatcher()   {}

// Object where the key itself is ignored, but the value template must match.
//
// key - Example key to use (which will be ignored)
// template - Example value template to base the comparison on
func EachKeyLike(key string, template interface{}) Matcher {
	return eachKeyLike{
		Type:     "values",
		Contents: template,
	}
}

type arrayContaining struct {
	Type     string        `json:"pact:matcher:type"`
	Variants []interface{} `json:"variants"`
}

func (a arrayContaining) GetValue() interface{} {
	return a.Variants
}

func (a arrayContaining) isV3Matcher() {}
func (a arrayContaining) isMatcher()   {}

func ArrayContaining(variants []interface{}) Matcher {
	return arrayContaining{
		Type:     "arrayContains",
		Variants: variants,
	}
}

type minMaxLike struct {
	Type     string      `json:"pact:matcher:type"`
	Contents interface{} `json:"value"`
	Min      int         `json:"min,omitempty"`
	Max      int         `json:"max,omitempty"` // NOTE: only used for V3
}

func (m minMaxLike) GetValue() interface{} {
	return m.Contents
}

func (m minMaxLike) isV3Matcher() {}
func (m minMaxLike) isMatcher()   {}

// ArrayMinMaxLike is like EachLike except has a bounds on the max and the min
// https://github.com/pact-foundation/pact-specification/tree/version-3#add-a-minmax-type-matcher
func ArrayMinMaxLike(content interface{}, min int, max int) Matcher {
	if min < 1 {
		log.Println("[WARN] min value to an array matcher can't be less than one")
		min = 1
	}
	examples := make([]interface{}, max)
	for i := 0; i < max; i++ {
		examples[i] = content
	}
	return minMaxLike{
		Type:     "type",
		Contents: examples,
		Min:      min,
		Max:      max,
	}
}

// ArrayMaxLike is like EachLike except has a bounds on the max
// https://github.com/pact-foundation/pact-specification/tree/version-3#add-a-minmax-type-matcher
func ArrayMaxLike(content interface{}, max int) Matcher {
	examples := make([]interface{}, max)
	for i := 0; i < max; i++ {
		examples[i] = content
	}

	return minMaxLike{
		Type:     "type",
		Contents: examples,
		Min:      1,
		Max:      max,
	}
}

type stringGenerator struct {
	Type      string `json:"pact:matcher:type"`
	Contents  string `json:"value"`
	Format    string `json:"format"`
	Generator string `json:"pact:generator:type"`
}

func (s stringGenerator) GetValue() interface{} {
	return s.Contents
}

func (s stringGenerator) isV3Matcher() {}
func (s stringGenerator) isMatcher()   {}

// DateGenerated matches a cross platform formatted date, and generates a current date during verification
// String example value must match the provided date format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func DateGenerated(example string, format string) Matcher {
	return stringGenerator{
		Type:      "date",
		Generator: "Date",
		Contents:  example,
		Format:    format,
	}
}

// TimeGenerated matches a cross platform formatted date, and generates a current time during verification
// String example value must match the provided time format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func TimeGenerated(example string, format string) Matcher {
	return stringGenerator{
		Type:      "time",
		Generator: "Time",
		Contents:  example,
		Format:    format,
	}
}

// DateTimeGenerated matches a cross platform formatted datetime, and generates a current datetime during verification
// String example value must match the provided datetime format string.
// See Java SimpleDateFormat https://docs.oracle.com/javase/8/docs/api/java/text/SimpleDateFormat.html for formatting options
func DateTimeGenerated(example string, format string) Matcher {
	return stringGenerator{
		Type:      "timestamp",
		Generator: "DateTime",
		Contents:  example,
		Format:    format,
	}
}
