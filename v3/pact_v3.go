package v3

import (
	"fmt"
	"log"
	"reflect"
)

// ruleValue is essentially a key value JSON pairs for serialisation
// TODO: this is actually more typed than this
//       once we understand the model better, let's make it more type-safe
type ruleValue map[string]interface{}

// Matching Rule
type ruleV3 struct {
	Body    ruleValue `json:"body,omitempty"`
	Headers ruleValue `json:"headers,omitempty"`
	Query   ruleValue `json:"query,omitempty"`
	Path    ruleValue `json:"path,omitempty"`
}

type matchingRuleV3 = ruleV3
type generatorV3 = ruleV3

type pactRequestV3 struct {
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Query         map[string][]string `json:"query,omitempty"`
	Headers       map[string]string   `json:"headers,omitempty"`
	Body          interface{}         `json:"body"`
	MatchingRules matchingRuleV3      `json:"matchingRules,omitempty"`
	Generators    generatorV3         `json:"generators"`
}

type pactResponseV3 struct {
	Status        int               `json:"status"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          interface{}       `json:"body,omitempty"`
	MatchingRules matchingRuleV3    `json:"matchingRules,omitempty"`
	Generators    generatorV3       `json:"generators"`
}

type pactInteractionV3 struct {
	Description string            `json:"description"`
	States      []ProviderStateV3 `json:"providerStates,omitempty"`
	Request     pactRequestV3     `json:"request"`
	Response    pactResponseV3    `json:"response"`
}

// pactFileV3 is what will be serialised to the Pactfile in the request body examples and matching rules
// given a structure containing matchers.
// TODO: any matching rules will need to be merged with other aspects (e.g. headers, path, query).
// ...still very much spike/POC code
type pactFileV3 struct {
	// Consumer is the name of the Consumer/Client.
	Consumer string `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider string `json:"provider"`

	// SpecificationVersion is the version of the Pact Spec this implementation supports
	SpecificationVersion SpecificationVersion `json:"-"`

	interactions []*InteractionV3

	// Interactions are all of the request/response expectations, with matching rules and generators
	Interactions []pactInteractionV3 `json:"interactions"`

	Metadata map[string]interface{} `json:"metadata"`
}

func pactInteractionFromV3Interaction(interaction InteractionV3) pactInteractionV3 {
	return pactInteractionV3{
		Description: interaction.Description,
		States:      interaction.States,
		Request: pactRequestV3{
			Method: interaction.Request.Method,
			Generators: generatorV3{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
			MatchingRules: generatorV3{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
		Response: pactResponseV3{
			Status: interaction.Response.Status,
			Generators: generatorV3{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
			MatchingRules: matchingRuleV3{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
	}
}

func (p *pactFileV3) generateV3PactFile() *pactFileV3 {
	for _, interaction := range p.interactions {
		fmt.Printf("Serialising interaction: %+v \n", *interaction)
		serialisedInteraction := pactInteractionFromV3Interaction(*interaction)

		// TODO: haven't done matchers for headers, query, path and status code
		// _, serialisedInteraction.Request.Headers, serialisedInteraction.Request.MatchingRules.Headers, _ = buildPactBody("", interaction.Request.Headers, make(map[string]interface{}), "$", make(ruleValue), make(ruleValue))
		_, serialisedInteraction.Request.Body, serialisedInteraction.Request.MatchingRules.Body, _ = buildPactPart("", interaction.Request.Body, make(map[string]interface{}), "$", make(ruleValue), make(ruleValue))
		// _, serialisedInteraction.Response.Headers, serialisedInteraction.Response.MatchingRules.Headers, _ = buildPactBody("", interaction.Response.Headers, make(map[string]interface{}), "$", make(ruleValue), make(ruleValue))
		_, serialisedInteraction.Response.Body, serialisedInteraction.Response.MatchingRules.Body, _ = buildPactPart("", interaction.Response.Body, make(map[string]interface{}), "$", make(ruleValue), make(ruleValue))

		// 		// TODO
		// 		buildPactHeaders()
		// 		buildPactQuery()
		buildPactPathV3(interaction, &serialisedInteraction)

		fmt.Printf("appending interaction: %+v \n", serialisedInteraction)
		p.Interactions = append(p.Interactions, serialisedInteraction)
	}

	return p
}

func recurseMapTypeV3(key string, value interface{}, body map[string]interface{}, path string,
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
			_, body, matchingRules, generators = buildPactPart(k.String(), v.Interface(), copyMap(body), path, matchingRules, generators)
		} else {
			_, body[key], matchingRules, generators = buildPactPart(k.String(), v.Interface(), entry, path, matchingRules, generators)
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
//  - generators    => Current set of generators rules (generators rules will also be serialised into the Pact)
//
// Returns path, body, matchingRules, generators
func buildPactBodyV3(key string, value interface{}, body map[string]interface{}, path string,
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
			buildPactPart("0", t.GetValue(), arrayMap, builtPath, matchingRules, generators)
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
			_, body, matchingRules, generators = recurseMapTypeV3(key, t.GetValue().(StructMatcher), body, path, matchingRules, generators)

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
			buildPactPart(k, el, arrayMap, builtPath, matchingRules, generators)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

		// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}, MapMatcher:
		log.Println("[TRACE] generate pact: MapMatcher")
		_, body, matchingRules, generators = recurseMapTypeV3(key, t, body, path, matchingRules, generators)

	// Primitives (terminal cases)
	default:
		log.Printf("[TRACE] generate pact: unknown type or primitive (%+v): %+v\n", reflect.TypeOf(t), value)
		body[key] = value
	}

	log.Printf("[TRACE] generate pact => returning body: %+v\n", body)

	return path, body, matchingRules, generators
}

func buildPactPathV3(sourceInteraction *InteractionV3, destInteraction *pactInteractionV3) *pactInteractionV3 {

	destInteraction.Request.MatchingRules.Path = sourceInteraction.Request.Path.MatchingRule()

	switch val := sourceInteraction.Request.Path.GetValue().(type) {
	case String:
		destInteraction.Request.Path = val.GetValue().(string)
	case like:
		destInteraction.Request.Path = val.GetValue().(string)
	case term:
		destInteraction.Request.Path = val.GetValue().(string)
	default:
		destInteraction.Request.MatchingRules.Path = nil
		log.Print("[WARN] ignoring unsupported matcher for request path:", val)
	}

	return destInteraction
}
