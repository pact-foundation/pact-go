package dsl

import (
	"encoding/json"
	"log"
)

// type Matcher interface{}

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, minRequired int) string {
	// TODO: should we just be marshalling these things as map[string]interface{} JSON objects anyway?
	//       this might remove the need for this ugly string/object combination
	// TODO: if content is a string, it should probably be immediately converted to an object
	// TODO: the above seems to have been fixed, but perhaps best to just _only_ allow objects
	//       instead of allowing string and other nonsense??
	return objectToString(map[string]interface{}{
		"json_class": "Pact::ArrayLike",
		"contents":   toObject(content),
		"min":        minRequired,
	})
	// return fmt.Sprintf(`
	// 	{
	// 	  "json_class": "Pact::ArrayLike",
	// 	  "contents": %v,
	// 	  "min": %d
	// 	}`, objectToString(content), minRequired)
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) string {
	return objectToString(map[string]interface{}{
		"json_class": "Pact::SomethingLike",
		"contents":   toObject(content),
	})
	// return fmt.Sprintf(`
	// 	{
	// 	  "json_class": "Pact::SomethingLike",
	// 	  "contents": %v
	// 	}`, objectToString(content))
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
func Term(generate string, matcher string) MatcherString {
	return MatcherString(objectToString(map[string]interface{}{
		"json_class": "Pact::Term",
		"data": map[string]interface{}{
			"generate": toObject(generate),
			"matcher": map[string]interface{}{
				"json_class": "Regexp",
				"o":          0,
				"s":          toObject(matcher),
			},
		},
	}))
}

// Takes an object and converts it to a JSON representation
func objectToString(obj interface{}) string {
	switch content := obj.(type) {
	case string:
		log.Println("STRING VALUE:", content)
		return content
	default:
		log.Printf("OBJECT VALUE: %v", obj)
		jsonString, err := json.Marshal(obj)
		log.Println("OBJECT -> JSON VALUE:", string(jsonString))
		if err != nil {
			log.Println("[DEBUG] interaction: error unmarshaling object into string:", err.Error())
			return ""
		}
		return string(jsonString)
	}
}
