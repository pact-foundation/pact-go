package dsl

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// Example matching rule / generated doc
// {
//     "method": "POST",
//     "path": "/",
//     "query": "",
//     "headers": {"Content-Type": "application/json"},
//     "matchingRules": {
//       "$.body.animals": {"min": 1, "match": "type"},
//       "$.body.animals[*].*": {"match": "type"},
//       "$.body.animals[*].children": {"min": 1, "match": "type"},
//       "$.body.animals[*].children[*].*": {"match": "type"}
//     },
//     "body": {
//       "animals": [
//         {
//           "name" : "Fred",
//           "children": [
//             {
//               "age": 9
//             }
//           ]
//         }
//       ]
//     }
// 	}

// matcherType is essentially a key value JSON pairs for serialisation
type matcherType map[string]interface{}

// Matching Rule
type matchingRuleType map[string]matcherType

// PactBody is what will be serialised to the Pactfile in the request body examples and matching rules
// given a structure containing matchers.
// TODO: any matching rules will need to be merged with other aspects (e.g. headers, path).
// ...still very much spike/POC code
type PactBody struct {
	// Matching rules used by the verifier to confirm Provider confirms to Pact.
	MatchingRules matchingRuleType `json:"matchingRules"`

	// Generated test body for the consumer testing via the Mock Server.
	Body map[string]interface{} `json:"body"`
}

// PactBodyBuilder takes a map containing recursive Matchers and generates the rules
// to be serialised into the Pact file.
func PactBodyBuilder(root map[string]interface{}) PactBody {
	dsl := PactBody{}
	fmt.Printf("root: %+v", root)
	_, dsl.Body, dsl.MatchingRules = build("", root, make(map[string]interface{}),
		"$.body", make(matchingRuleType))

	return dsl
}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

// Recurse the Matcher tree and build up an example body and set of matchers for
// the Pact file. Ideally this stays as a pure function, but probably might need
// to store matchers externally.
//
// See PactBody.groovy line 96 for inspiration/logic.
//
// Arguments:
// 	- key           => Current key in the body to set
// 	- value         => Value for the current key, may be a primitive, object or another Matcher
// 	- body          => Current state of the body map (body will be the returned Pact body for serialisation)
// 	- path          => Path to the current key
//  - matchingRules => Current set of matching rules (matching rules will also be serialised into the Pact)
func build(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules matchingRuleType) (string, map[string]interface{}, matchingRuleType) {
	log.Println("[DEBUG] dsl generator: recursing => key:", key, ", body:", body, ", value: ", value)

	switch t := value.(type) {

	case Matcher:
		switch t.Type() {

		// ArrayLike Matchers
		case ArrayMinLikeMatcher, ArrayMaxLikeMatcher:
			fmt.Println("Array matcher")
			times := 1

			m := t.(eachLike)
			if m.Max > 0 {
				times = m.Max
			} else if m.Min > 0 {
				times = m.Min
			}

			arrayMap := make(map[string]interface{})
			minArray := make([]interface{}, times)

			// TODO: why does this exist? -> Umm, it's what recurses the array item values!
			build("0", t.GetValue(), arrayMap, path+buildPath(key, allListItems), matchingRules)
			log.Println("[DEBUG] dsl generator: adding matcher (arrayLike) =>", path+buildPath(key, ""))
			matchingRules[path+buildPath(key, "")] = m.MatchingRule()

			// TODO: Need to understand the .* notation before implementing it. Notably missing from Groovy DSL
			// log.Println("[DEBUG] dsl generator: Adding matcher (type)              =>", path+buildPath(key, allListItems)+".*")
			// matchingRules[path+buildPath(key, allListItems)+".*"] = m.MatchingRule()

			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}

			// TODO: I think this assignment is working, but the next step seems to recurse again and this never writes
			// probably just a bad terminal case handling
			body[key] = minArray
			fmt.Printf("Updating body: %+v, minArray: %+v", body, minArray)
			path = path + buildPath(key, "")

		// Simple Matchers (Terminal cases)
		case RegexMatcher, LikeMatcher:
			fmt.Println("regex matcher")
			body[key] = t.GetValue()
			log.Println("[DEBUG] dsl generator: adding matcher (Term/Like)         =>", path+buildPath(key, ""))
			matchingRules[path+buildPath(key, "")] = t.MatchingRule()
		default:
			log.Fatalf("unknown matcher: %d", t)
		}

	// Slice/Array types
	case []interface{}:
		arrayValues := make([]interface{}, len(t))
		arrayMap := make(map[string]interface{})

		// This is a real hack. I don't like it
		// I also had to do it for the Array*LikeMatcher's, which I also don't like
		for i, el := range t {
			k := fmt.Sprintf("%d", i)
			build(k, el, arrayMap, path+buildPath(key, fmt.Sprintf("%s%d%s", startList, i, endList)), matchingRules)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

	// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}:
		entry := make(map[string]interface{})
		path = path + buildPath(key, "")

		for k, v := range t {
			log.Println("[DEBUG] dsl generator => map type. recursing into key =>", k)

			// Starting position
			if key == "" {
				_, body, matchingRules = build(k, v, copyMap(body), path, matchingRules)
			} else {
				_, body[key], matchingRules = build(k, v, entry, path, matchingRules)
			}
		}

	// Primitives (terminal cases)
	default:
		fmt.Println("type", reflect.TypeOf(t))
		log.Printf("[DEBUG] dsl generator => unknown type, probably just a primitive (string/int/etc.): %+v", value)
		body[key] = value
	}

	log.Println("[DEBUG] dsl generator => returning body: ", body)

	return path, body, matchingRules
}

// TODO: allow regex in request paths.
func buildPath(name string, children string) string {
	// We know if a key is an integer, it's not valid JSON and therefore is Probably
	// the shitty array hack from above. Skip creating a new path if the key is bungled
	// TODO: save the children?
	if _, err := strconv.Atoi(name); err != nil && name != "" {
		return pathSep + name + children
	}

	return ""
}
