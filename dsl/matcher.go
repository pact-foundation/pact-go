package dsl

import (
	"fmt"
	"html/template"
)

var funcMap = template.FuncMap{
	"like":     Like,
	"eachLike": EachLike,
	"term":     Term,
}

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, minRequired int) string {
	return fmt.Sprintf(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": %v,
		  "min": %d
		}`, content, minRequired)
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) string {
	return fmt.Sprintf(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": %v
		}`, content)
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) string {
	return fmt.Sprintf(`
		{
			"json_class": "Pact::Term",
			"data": {
			  "generate": "%s",
			  "matcher": {
			    "json_class": "Regexp",
			    "o": 0,
			    "s": "%s"
			  }
			}
		}`, generate, matcher)
}
