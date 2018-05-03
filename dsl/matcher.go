package dsl

import (
	"fmt"
	"reflect"
	"strings"
)

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
func Match(src interface{}) string {
	return match(reflect.TypeOf(src), getDefaults())
}

// match recursively traverses the provided type and outputs a
// matcher string for it that is compatible with the Pact dsl.
func match(srcType reflect.Type, params params) string {
	switch kind := srcType.Kind(); kind {
	case reflect.Ptr:
		return match(srcType.Elem(), params)
	case reflect.Slice, reflect.Array:
		return EachLike(match(srcType.Elem(), getDefaults()), params.slice.min)
	case reflect.Struct:
		result := `{`
		for i := 0; i < srcType.NumField(); i++ {
			field := srcType.Field(i)
			result += fmt.Sprintf(
				`"%s": %s,`,
				field.Tag.Get("json"),
				match(field.Type, pluckParams(field.Type, field.Tag.Get("pact"))),
			)
		}
		return strings.TrimSuffix(result, ",") + `}`
	case reflect.String:
		if params.str.regEx != "" {
			return Term(params.str.example, params.str.regEx)
		}
		return Like(`"string"`)
	case reflect.Bool:
		return Like(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return Like(1)
	default:
		panic(fmt.Sprintf("match: unhandled type: %v", srcType))
	}
}

// params are plucked from 'pact' struct tags as match() traverses
// struct fields. They are passed back into match() along with their
// associated type to serve as parameters for the dsl functions.
type params struct {
	slice sliceParams
	str   stringParams
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
	case reflect.Slice:
		if _, err := fmt.Sscanf(pactTag, "min=%d", &params.slice.min); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}
	case reflect.String:
		components := strings.Split(pactTag, ",regex=")

		if len(components) != 2 {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: unable to split on ',regex='"))
		} else if len(components[1]) == 0 {
			triggerInvalidPactTagPanic(pactTag, fmt.Errorf("invalid format: regex must not be empty"))
		} else if _, err := fmt.Sscanf(components[0], "example=%s", &params.str.example); err != nil {
			triggerInvalidPactTagPanic(pactTag, err)
		}

		params.str.regEx = strings.Replace(components[1], `\`, `\\`, -1)
	}

	return params
}

func triggerInvalidPactTagPanic(tag string, err error) {
	panic(fmt.Sprintf("match: encountered invalid pact tag %q . . . parsing failed with error: %v", tag, err))
}
