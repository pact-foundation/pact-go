package v3

import (
	"fmt"
	"log"
	"reflect"
)

type object map[string]interface{}

// ruleSet are the set of matchers to apply at a path, and the logical operation in which to apply them
// TODO: this is actually more typed than this
//       once we understand the model better, let's make it more type-safe
type ruleSet map[string]matchers

// type ruleValue map[string]interface{}
type matcherLogic string

const (
	// AND specifies a logical AND to the matching rule application
	AND matcherLogic = "AND"

	// OR specifies a logical OR to the matching rule application
	OR = "OR"
)

type matchers struct {
	Combine  matcherLogic `json:"combine,omitempty"`
	Matchers []rule       `json:"matchers,omitempty"`
}

// Matching Rule
type ruleV3 struct {
	Body    ruleSet  `json:"body,omitempty"`
	Headers ruleSet  `json:"headers,omitempty"`
	Query   ruleSet  `json:"query,omitempty"`
	Path    matchers `json:"path,omitempty"`
}

type matchingRuleV3 = ruleV3
type generatorV3 = ruleV3

type pactRequestV3 struct {
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Query         map[string][]string `json:"query,omitempty"`
	Headers       interface{}         `json:"headers,omitempty"`
	Body          interface{}         `json:"body"`
	MatchingRules matchingRuleV3      `json:"matchingRules,omitempty"`
	Generators    generatorV3         `json:"generators"`
}

type pactResponseV3 struct {
	Status        int            `json:"status"`
	Headers       interface{}    `json:"headers,omitempty"`
	Body          interface{}    `json:"body,omitempty"`
	MatchingRules matchingRuleV3 `json:"matchingRules,omitempty"`
	Generators    generatorV3    `json:"generators"`
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
				Body:    make(ruleSet),
				Headers: make(ruleSet),
				Query:   make(ruleSet),
			},
			MatchingRules: generatorV3{
				Body:    make(ruleSet),
				Headers: make(ruleSet),
				Query:   make(ruleSet),
			},
		},
		Response: pactResponseV3{
			Status: interaction.Response.Status,
			Generators: generatorV3{
				Body:    make(ruleSet),
				Headers: make(ruleSet),
				Query:   make(ruleSet),
			},
			MatchingRules: matchingRuleV3{
				Body:    make(ruleSet),
				Headers: make(ruleSet),
				Query:   make(ruleSet),
			},
		},
	}
}

func (p *pactFileV3) generateV3PactFile() *pactFileV3 {
	for _, interaction := range p.interactions {
		serialisedInteraction := pactInteractionFromV3Interaction(*interaction)

		var requestQuery object

		requestQuery, serialisedInteraction.Request.MatchingRules.Query, serialisedInteraction.Request.Generators.Query = buildPart(interaction.Request.Query)
		serialisedInteraction.Request.Headers, serialisedInteraction.Request.MatchingRules.Headers, serialisedInteraction.Request.Generators.Headers = buildPart(interaction.Request.Headers)
		serialisedInteraction.Request.Body, serialisedInteraction.Request.MatchingRules.Body, serialisedInteraction.Request.Generators.Body = buildPart(interaction.Request.Body)
		serialisedInteraction.Response.Headers, serialisedInteraction.Response.MatchingRules.Headers, serialisedInteraction.Response.Generators.Headers = buildPart(interaction.Response.Headers)
		serialisedInteraction.Response.Body, serialisedInteraction.Response.MatchingRules.Body, serialisedInteraction.Response.Generators.Body = buildPart(interaction.Response.Body)

		// TODO: Generators

		buildQueryV3(requestQuery, interaction, &serialisedInteraction)
		buildPactPathV3(interaction, &serialisedInteraction)

		p.Interactions = append(p.Interactions, serialisedInteraction)
		fmt.Printf("%+v", serialisedInteraction)
	}

	return p
}

func recurseMapTypeV3(key string, value interface{}, body object, path string,
	matchingRules ruleSet, generators ruleSet) (string, object, ruleSet, ruleSet) {
	mapped := reflect.ValueOf(value)
	entry := make(object)
	path = path + buildPath(key, "")

	iter := mapped.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		log.Println("[TRACE] generate pact: map[string]interface{}: recursing map type into key =>", k)

		if key == "" {
			// Starting position
			_, body, matchingRules, generators = buildPactPartV3(k.String(), v.Interface(), copyMap(body), path, matchingRules, generators)
		} else {
			_, body[key], matchingRules, generators = buildPactPartV3(k.String(), v.Interface(), entry, path, matchingRules, generators)
		}
	}

	return path, body, matchingRules, generators
}

func wrapMatchingRule(r rule) matchers {
	fmt.Println("[DEBUG] wrapmatchingrule")
	return matchers{
		Combine:  AND,
		Matchers: []rule{r},
	}
}

func buildPart(value interface{}) (object, ruleSet, ruleSet) {
	_, o, matchingRules, generators := buildPactPartV3("", value, make(object), "$", make(ruleSet), make(ruleSet))
	return o, matchingRules, generators
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
func buildPactPartV3(key string, value interface{}, body object, path string,
	matchingRules ruleSet, generators ruleSet) (string, object, ruleSet, ruleSet) {
	log.Println("[TRACE] generate pact => key:", key, ", body:", body, ", value:", value, ", path:", path)

	switch t := value.(type) {

	case MatcherV2:
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

			builtPath := path + buildPath(key, allListItems)
			buildPactPartV3("0", t.GetValue(), arrayMap, builtPath, matchingRules, generators)
			log.Println("[TRACE] generate pact: ArrayMikeLikeMatcher/ArrayMaxLikeMatcher =>", builtPath)
			matchingRules[path+buildPath(key, "")] = wrapMatchingRule(m.MatchingRule())

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
			matchingRules[builtPath] = wrapMatchingRule(t.MatchingRule())

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
			buildPactPartV3(k, el, arrayMap, builtPath, matchingRules, generators)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

		// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}, MapMatcher, QueryMatcher:
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
	destInteraction.Request.MatchingRules.Path = wrapMatchingRule(sourceInteraction.Request.Path.MatchingRule())

	switch val := sourceInteraction.Request.Path.GetValue().(type) {
	case String:
		destInteraction.Request.Path = val.GetValue().(string)
	case like:
		destInteraction.Request.Path = val.GetValue().(string)
	case term:
		destInteraction.Request.Path = val.GetValue().(string)
	case string:
		destInteraction.Request.Path = val
	default:
		destInteraction.Request.MatchingRules.Path = matchers{}
		log.Printf("[WARN] ignoring unsupported matcher for request path: %+v", val)
	}

	return destInteraction
}

func buildQueryV3(input object, sourceInteraction *InteractionV3, destInteraction *pactInteractionV3) *pactInteractionV3 {
	queryAsMap := make(map[string][]string)

	for k, v := range input {
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			slice := v.([]interface{})
			l := len(slice)

			values := make([]string, l)
			for i, data := range slice {
				values[i] = fmt.Sprintf("%s", data)
			}
			queryAsMap[k] = values
		default:
			queryAsMap[k] = []string{fmt.Sprintf("%s", v)}
		}
	}

	destInteraction.Request.Query = queryAsMap

	return destInteraction
}
