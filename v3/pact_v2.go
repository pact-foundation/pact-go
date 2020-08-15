package v3

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

type pactRequestV2 struct {
	Method         string                 `json:"method"`
	Path           Matcher                `json:"path"`
	Query          MapMatcher             `json:"query,omitempty"`
	Headers        MapMatcher             `json:"headers,omitempty"`
	Body           interface{}            `json:"body"`
	MatchingRules  map[string]interface{} `json:"matchingRules"`
	MatchingRules2 matchingRule           `json:"matchingRules2,omitempty"`
	Generators     generator              `json:"generators"`
}

type pactResponseV2 struct {
	Status         int                    `json:"status"`
	Headers        MapMatcher             `json:"headers,omitempty"`
	Body           interface{}            `json:"body,omitempty"`
	MatchingRules  map[string]interface{} `json:"matchingRules"`
	MatchingRules2 matchingRule           `json:"matchingRules2,omitempty"`
	Generators     generator              `json:"generators"`
}

type pactInteractionV2 struct {
	Description string         `json:"description"`
	State       string         `json:"providerState,omitempty"`
	Request     pactRequestV2  `json:"request"`
	Response    pactResponseV2 `json:"response"`
}

// pactFileV2 is what will be serialised to the Pactfile in the request body examples and matching rules
// given a structure containing matchers.
// TODO: any matching rules will need to be merged with other aspects (e.g. headers, path, query).
// ...still very much spike/POC code
type pactFileV2 struct {
	// Consumer is the name of the Consumer/Client.
	Consumer string `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider string `json:"provider"`

	// SpecificationVersion is the version of the Pact Spec this implementation supports
	SpecificationVersion SpecificationVersion `json:"-"`

	interactions []*InteractionV2

	// Interactions are all of the request/response expectations, with matching rules and generators
	Interactions []pactInteractionV2 `json:"interactions"`

	Metadata map[string]interface{} `json:"metadata"`
}

func pactInteractionFromV2Interaction(interaction InteractionV2) pactInteractionV2 {
	return pactInteractionV2{
		Description: interaction.Description,
		State:       interaction.State,
		Request: pactRequestV2{
			Method:        interaction.Request.Method,
			Body:          interaction.Request.Body,
			Headers:       interaction.Request.Headers,
			Query:         interaction.Request.Query,
			Path:          interaction.Request.Path,
			MatchingRules: make(ruleValue),
		},
		Response: pactResponseV2{
			Status:        interaction.Response.Status,
			Body:          interaction.Response.Body,
			Headers:       interaction.Response.Headers,
			MatchingRules: make(ruleValue),
		},
	}
}

func (p *pactFileV2) generateV2PactFile() *pactFileV2 {
	for _, interaction := range p.interactions {
		fmt.Printf("Serialising interaction: %+v \n", *interaction)
		serialisedInteraction := pactInteractionFromV2Interaction(*interaction)

		// TODO: haven't done matchers for headers, path and status code
		_, serialisedInteraction.Request.Body, serialisedInteraction.Request.MatchingRules, _ = buildPactBody("", interaction.Request.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))
		_, serialisedInteraction.Response.Body, serialisedInteraction.Response.MatchingRules, _ = buildPactBody("", interaction.Response.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))

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
func pactBodyBuilder(root map[string]interface{}) pactFileV2 {
	// Generators:    make(generatorType),
	// MatchingRules: make(matchingRuleType),
	// Metadata:      pactGoMetadata,
	// return file
	return pactFileV2{}

}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

func recurseMapType(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules ruleValue, generators ruleValue) (string, map[string]interface{}, ruleValue, ruleValue) {
	mapped := reflect.ValueOf(value)
	entry := make(map[string]interface{})
	path = path + buildPath(key, "")

	iter := mapped.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		log.Println("[TRACE] generate pact: map[string]interface{}: recursing map type into key =>", k)

		// Starting position
		if key == "" {
			_, body, matchingRules, generators = buildPactBody(k.String(), v.Interface(), copyMap(body), path, matchingRules, generators)
		} else {
			_, body[key], matchingRules, generators = buildPactBody(k.String(), v.Interface(), entry, path, matchingRules, generators)
		}
	}

	return path, body, matchingRules, generators
}

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
//
// Returns path, body, matchingRules, generators
func buildPactBody(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules ruleValue, generators ruleValue) (string, map[string]interface{}, ruleValue, ruleValue) {
	log.Println("[TRACE] generate pact => key:", key, ", body:", body, ", value:", value, ", path:", path)

	switch t := value.(type) {

	case Matcher:
		switch t.Type() {

		case arrayMinLikeMatcher, arrayMaxLikeMatcher:
			log.Println("[TRACE] generate pact: ArrayMikeLikeMatcher/ArrayMaxLikeMatcher")
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
			builtPath := path + buildPath(key, allListItems)
			buildPactBody("0", t.GetValue(), arrayMap, builtPath, matchingRules, generators)
			log.Println("[TRACE] generate pact: ArrayMikeLikeMatcher/ArrayMaxLikeMatcher =>", builtPath)
			matchingRules[path+buildPath(key, "")] = m.MatchingRule()

			// TODO: Need to understand the .* notation before implementing it. Notably missing from Groovy DSL
			// log.Println("[TRACE] generate pact: matcher (type)              =>", path+buildPath(key, allListItems)+".*")
			// matchingRules[path+buildPath(key, allListItems)+".*"] = m.MatchingRule()

			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}

			// TODO: I think this assignment is working, but the next step seems to recurse again and this never writes
			// probably just a bad terminal case handling?
			body[key] = minArray
			fmt.Printf("Updating body: %+v, minArray: %+v", body, minArray)
			path = path + buildPath(key, "")

		case regexMatcher, likeMatcher:
			log.Println("[TRACE] generate pact: Regex/LikeMatcher")
			builtPath := path + buildPath(key, "")
			body[key] = t.GetValue()
			log.Println("[TRACE] generate pact: Regex/LikeMatcher =>", builtPath)
			matchingRules[builtPath] = t.MatchingRule()

		// This exists to server the v3.Match() interface
		case structTypeMatcher:
			log.Println("[TRACE] generate pact: StructTypeMatcher")
			_, body, matchingRules, generators = recurseMapType(key, t.GetValue().(StructMatcher), body, path, matchingRules, generators)

		default:
			log.Fatalf("unexpected matcher (%s) for current specification format (2.0.0)", t.Type())
		}

		// Slice/Array types
	case []interface{}:
		log.Println("[TRACE] generate pact: []interface{}")
		arrayValues := make([]interface{}, len(t))
		arrayMap := make(map[string]interface{})

		// This is a real hack. I don't like it
		// I also had to do it for the Array*LikeMatcher's, which I also don't like
		for i, el := range t {
			k := fmt.Sprintf("%d", i)
			builtPath := path + buildPath(key, fmt.Sprintf("%s%d%s", startList, i, endList))
			log.Println("[TRACE] generate pact: []interface{}: recursing into =>", builtPath)
			buildPactBody(k, el, arrayMap, builtPath, matchingRules, generators)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

		// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}, MapMatcher:
		log.Println("[TRACE] generate pact: MapMatcher")
		_, body, matchingRules, generators = recurseMapType(key, t, body, path, matchingRules, generators)

	// Primitives (terminal cases)
	default:
		log.Printf("[TRACE] generate pact: unknown type or primitive (%+v): %+v\n", reflect.TypeOf(t), value)
		body[key] = value
	}

	log.Printf("[TRACE] generate pact => returning body: %+v\n", body)

	return path, body, matchingRules, generators
}

func buildPactHeaders() {}
func buildPactPath()    {}
func buildPactQuery()   {}

// TODO: allow regex in request paths.
func buildPath(name string, children string) string {
	// We know if a key is an integer, it's not valid JSON and therefore is probably
	// the shitty array hack from above. Skip creating a new path if the key is bungled
	// TODO: save the children?
	if _, err := strconv.Atoi(name); err != nil && name != "" {
		return pathSep + name + children
	}

	return ""
}
