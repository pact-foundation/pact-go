package dsl

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"
)

// Matcher regexes
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
	Type     string      `json:"json_class"`
	Contents interface{} `json:"contents"`
	Min      int         `json:"min"`
}

type like struct {
	Type     string      `json:"json_class"`
	Contents interface{} `json:"contents"`
}

type term struct {
	Type string `json:"json_class"`
	Data struct {
		Generate interface{} `json:"generate"`
		Matcher  struct {
			Type  string      `json:"json_class"`
			O     int         `json:"o"`
			Regex interface{} `json:"s"`
		} `json:"matcher"`
	} `json:"data"`
}

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, minRequired int) Matcher {
	return Matcher{
		"json_class": "Pact::ArrayLike",
		"contents":   content,
		"min":        minRequired,
	}
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) Matcher {
	return Matcher{
		"json_class": "Pact::SomethingLike",
		"contents":   content,
	}
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) Matcher {
	return Matcher{
		"json_class": "Pact::Term",
		"data": map[string]interface{}{
			"generate": generate,
			"matcher": map[string]interface{}{
				"json_class": "Regexp",
				"o":          0,
				"s":          matcher,
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

// StringMatcher allows a string or Matcher to be provided in
// when matching with the DSL
// We use the strategy outlined at http://www.jerf.org/iri/post/2917
// to create a "sum" or "union" type.
type StringMatcher interface {
	// isMatcher is how we tell the compiler that strings
	// and other types are the same / allowed
	isMatcher()

	// GetValue returns the raw generated value for the matcher
	// without any of the matching detail context
	GetValue() interface{}
}

// S is the string primitive wrapper (alias) for the StringMatcher type,
// it allows plain strings to be matched
// To keep backwards compatible with previous versions
// we aren't using an alias here
type S string

func (s S) isMatcher() {}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (s S) GetValue() interface{} {
	return s
}

// String is the longer named form of the string primitive wrapper,
// it allows plain strings to be matched
type String string

func (s String) isMatcher() {}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (s String) GetValue() interface{} {
	return s
}

// Matcher matches a complex object structure, which may itself
// contain nested Matchers
type Matcher map[string]interface{}

func (m Matcher) isMatcher() {}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func (m Matcher) GetValue() interface{} {
	class, ok := m["json_class"]

	if !ok {
		return nil
	}

	// extract out the value
	switch class {
	case "Pact::ArrayLike":
		contents := m["contents"]
		min := m["min"].(int)
		data := make([]interface{}, min)

		for i := 0; i < min; i++ {
			data[i] = contents
		}

	case "Pact::SomethingLike":
		return m["contents"]
	case "Pact::Term":
		data := m["data"].(map[string]interface{})
		return data["generate"]
	}

	return nil
}

// GetValue returns the raw generated value for the matcher
// without any of the matching detail context
func getMatcherValue(m interface{}) interface{} {
	matcher, ok := getMatcher(m)
	if !ok {
		return nil
	}

	class, ok := matcher["json_class"]

	if !ok {
		return nil
	}

	// extract out the value
	switch class {
	case "Pact::ArrayLike":
		contents := matcher["contents"]
		min := matcher["min"].(int)
		data := make([]interface{}, min)

		for i := 0; i < min; i++ {
			data[i] = contents
		}
		return data

	case "Pact::SomethingLike":
		return matcher["contents"]
	case "Pact::Term":
		data := matcher["data"].(map[string]interface{})
		return data["generate"]
	}

	return nil
}

// func isMatcher(obj map[string]interface{}) bool {
func isMatcher(obj interface{}) bool {
	m, ok := obj.(map[string]interface{})

	if ok {
		if _, match := m["json_class"]; match {
			return true
		}
	}

	if _, match := obj.(Matcher); match {
		return true
	}

	return false
}

func getMatcher(obj interface{}) (Matcher, bool) {
	// If an object, but not a map[string]interface{} then just return?
	m, ok := obj.(map[string]interface{})

	if ok {
		if _, match := m["json_class"]; match {
			return m, true
		}
	}

	m, ok = obj.(Matcher)
	if ok {
		return m, true
	}

	fmt.Println("NOT a matcher")
	return nil, false
}

func extractPayload(obj interface{}) interface{} {
	fmt.Println("extractpaload")

	// special case: top level matching object
	// we need to strip the properties
	matcher, ok := getMatcher(obj)

	if ok {
		fmt.Println("top level matcher", matcher, "returning value:", getMatcherValue(matcher))
		return extractPayload(getMatcherValue(matcher))
	}

	fmt.Println("not a top level matcher", matcher, "returning value:", obj)
	return extractPayloadRecursive(obj, make(map[string]interface{}))
}

// Recurse the object removing any underlying matching guff, returning
// the raw example content (ready for JSON marshalling)
// NOTE: type information is going to be lost here which is OK
//       because it must be mapped to JSON encodable types
//       It is expected that any object is marshalled to JSON and into a map[string]interface{}
//       for use here
//       It will probably break custom, user-supplied types? e.g. a User{} or ShoppingCart{}?
//       But then any enclosed Matchers will likely break them anyway
func extractPayloadRecursive(obj interface{}, stack map[string]interface{}) map[string]interface{} {
	fmt.Println("extracting payload recursively")

	objectMap, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	// recurse the (remaining) object, replacing Matchers with their
	// actual contents
	for k, rawValue := range objectMap {
		fmt.Println(k, "=>", rawValue, "(raw)")
		// v, ok := rawValue.(map[string]interface{})
		// fmt.Println(k, "=>", v)

		if ok && isMatcher(rawValue) {
			fmt.Println("v is Matcher")
			matcherValue := getMatcherValue(rawValue)
			stack[k] = matcherValue
			extractPayloadRecursive(matcherValue, stack)
		} else {
			fmt.Println("v is not Matcher but of type", reflect.TypeOf(rawValue))
			stack[k] = rawValue
			extractPayloadRecursive(rawValue, stack)
		}
	}

	return stack
}

// MapMatcher allows a map[string]string-like object
// to also contain complex matchers
type MapMatcher map[string]StringMatcher

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
