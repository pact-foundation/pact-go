package v3

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	version "github.com/pact-foundation/pact-go/command"
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

// ruleValue is essentially a key value JSON pairs for serialisation
// TODO: this is actually more typed than this
//       once we understand the model better, let's make it more type-safe
type ruleValue map[string]interface{}

// Matching Rule
type rule struct {
	Body    ruleValue `json:"body,omitempty"`
	Headers ruleValue `json:"headers,omitempty"`
	Query   ruleValue `json:"query,omitempty"`
	Path    ruleValue `json:"path,omitempty"`
}

type matchingRule = rule
type generator = rule

var pactGoMetadata = map[string]interface{}{
	"pactGo": map[string]string{
		"version": version.Version,
	},
	"pactSpecification": map[string]interface{}{
		"version": "2.0.0",
	},
}

// type pactFileBody map[string]interface{}

type pactRequest struct {
	Method         string                 `json:"method"`
	Path           Matcher                `json:"path"`
	Query          MapMatcher             `json:"query,omitempty"`
	Headers        MapMatcher             `json:"headers,omitempty"`
	Body           interface{}            `json:"body"`
	MatchingRules  map[string]interface{} `json:"matchingRules"`
	MatchingRules2 matchingRule           `json:"matchingRules2,omitempty"`
	Generators     generator              `json:"generators"`
}

type pactResponse struct {
	Status         int                    `json:"status"`
	Headers        MapMatcher             `json:"headers,omitempty"`
	Body           interface{}            `json:"body,omitempty"`
	MatchingRules  map[string]interface{} `json:"matchingRules"`
	MatchingRules2 matchingRule           `json:"matchingRules2,omitempty"`
	Generators     generator              `json:"generators"`
}

type pactInteraction struct {
	Description string       `json:"description"`
	State       string       `json:"providerState,omitempty"`
	Request     pactRequest  `json:"request"`
	Response    pactResponse `json:"response"`
}

// pactFile is what will be serialised to the Pactfile in the request body examples and matching rules
// given a structure containing matchers.
// TODO: any matching rules will need to be merged with other aspects (e.g. headers, path, query).
// ...still very much spike/POC code
type pactFile struct {
	// Consumer is the name of the Consumer/Client.
	Consumer string `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider string `json:"provider"`

	// SpecificationVersion is the version of the Pact Spec this implementation supports
	SpecificationVersion int `json:"pactSpecificationVersion,string"`

	interactions []*Interaction

	// Interactions are all of the request/response expectations, with matching rules and generators
	Interactions []pactInteraction `json:"interactions"`

	Metadata map[string]interface{} `json:"metadata"`
}

func pactInteractionFromInteraction(interaction Interaction) pactInteraction {
	return pactInteraction{
		Description: interaction.Description,
		State:       interaction.State,
		Request: pactRequest{
			Method:  interaction.Request.Method,
			Body:    interaction.Request.Body,
			Headers: interaction.Request.Headers,
			Query:   interaction.Request.Query,
			Path:    interaction.Request.Path,
			// Generators:    make(generatorType),
			MatchingRules: make(ruleValue),
			MatchingRules2: generator{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
		Response: pactResponse{
			Status:  interaction.Response.Status,
			Body:    interaction.Response.Body,
			Headers: interaction.Response.Headers,
			// Generators:    make(generatorType),
			MatchingRules: make(ruleValue),
			MatchingRules2: matchingRule{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
	}
}

func NewPactFile(Consumer string, Provider string, interactions []*Interaction) pactFile {
	p := pactFile{
		Interactions:         make([]pactInteraction, 0),
		interactions:         interactions,
		Metadata:             pactGoMetadata,
		Consumer:             Consumer,
		Provider:             Provider,
		SpecificationVersion: 2,
	}
	p.generatePactFile()

	return p
}

func (p *pactFile) generatePactFile() *pactFile {
	for _, interaction := range p.interactions {
		fmt.Printf("Serialising interaction: %+v \n", *interaction)
		serialisedInteraction := pactInteractionFromInteraction(*interaction)

		// TODO: this is just the request body, need the same for response!
		_, _, requestBodyMatchingRules, _ := buildPactBody("", interaction.Request.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))
		_, _, responseBodyMatchingRules, _ := buildPactBody("", interaction.Response.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))

		// v2
		serialisedInteraction.Request.MatchingRules = requestBodyMatchingRules
		serialisedInteraction.Response.MatchingRules = responseBodyMatchingRules

		// v3 only
		// serialisedInteraction.Request.MatchingRules.Body = requestBodyMatchingRules
		// serialisedInteraction.Response.MatchingRules.Body = responseBodyMatchingRules

		// TODO
		buildPactHeaders()
		buildPactQuery()
		buildPactPath()

		fmt.Printf("appending interaction: %+v \n", serialisedInteraction)
		p.Interactions = append(p.Interactions, serialisedInteraction)
	}

	return p
}

// pactBodyBuilder takes a map containing recursive Matchers and generates the rules
// to be serialised into the Pact file.
func pactBodyBuilder(root map[string]interface{}) pactFile {
	// Generators:    make(generatorType),
	// MatchingRules: make(matchingRuleType),
	// Metadata:      pactGoMetadata,
	// return file
	return pactFile{}

}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

// Recurse the Matcher tree and buildPactBody up an example body and set of matchers for
// the Pact file. Ideally this stays as a pure function, but probably might need
// to store matchers externally.
//
// See PactBody.groovy line 96 for inspiration/logic.
//
// Arguments:
// 	- key           => Current key in the body to set
// 	- value         => Value for the current key, may be a primitive, object or another Matcher
// 	- body          => Current state of the body map to be built up (body will be the returned Pact body for serialisation)
// 	- path          => Path to the current key
//  - matchingRules => Current set of matching rules (matching rules will also be serialised into the Pact)
func buildPactBody(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules ruleValue, generators ruleValue) (string, map[string]interface{}, ruleValue, ruleValue) {
	log.Println("[DEBUG] dsl generator: recursing => key:", key, ", body:", body, ", value: ", value)

	switch t := value.(type) {

	case Matcher:
		switch t.Type() {

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
			buildPactBody("0", t.GetValue(), arrayMap, path+buildPath(key, allListItems), matchingRules, generators)
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
			buildPactBody(k, el, arrayMap, path+buildPath(key, fmt.Sprintf("%s%d%s", startList, i, endList)), matchingRules, generators)
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
				_, body, matchingRules, generators = buildPactBody(k, v, copyMap(body), path, matchingRules, generators)
			} else {
				_, body[key], matchingRules, generators = buildPactBody(k, v, entry, path, matchingRules, generators)
			}
		}

	// Specialised case of map (above)
	// TODO: DRY  this?
	case MapMatcher:
		entry := make(map[string]interface{})
		path = path + buildPath(key, "")

		for k, v := range t {
			log.Println("[DEBUG] dsl generator => map type. recursing into key =>", k)

			// Starting position
			if key == "" {
				_, body, matchingRules, generators = buildPactBody(k, v, copyMap(body), path, matchingRules, generators)
			} else {
				_, body[key], matchingRules, generators = buildPactBody(k, v, entry, path, matchingRules, generators)
			}
		}

	// Primitives (terminal cases)
	default:
		fmt.Println("type", reflect.TypeOf(t))
		log.Printf("[DEBUG] dsl generator => unknown type, probably just a primitive (string/int/etc.): %+v", value)
		body[key] = value
	}

	log.Println("[DEBUG] dsl generator => returning body: ", body)

	return path, body, matchingRules, generators
}

func buildPactHeaders() {}
func buildPactPath()    {}
func buildPactQuery()   {}

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
