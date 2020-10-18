package v3

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type rule map[string]interface{}
type requestQueryV2 map[string][]string

type pactRequestV2 struct {
	Method        string                 `json:"method"`
	Path          string                 `json:"path"`
	Query         string                 `json:"query,omitempty"`
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Body          interface{}            `json:"body"`
	MatchingRules map[string]interface{} `json:"matchingRules"`
}

type pactResponseV2 struct {
	Status        int                    `json:"status"`
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Body          interface{}            `json:"body,omitempty"`
	MatchingRules map[string]interface{} `json:"matchingRules"`
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
	Consumer Pacticipant `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider Pacticipant `json:"provider"`

	// SpecificationVersion is the version of the Pact Spec this implementation supports
	SpecificationVersion SpecificationVersion `json:"-"`

	// Raw incoming interactions from the consumer test
	interactions []*InteractionV2

	// Interactions are the annotated set the request/response expectations, with matching rules
	Interactions []pactInteractionV2 `json:"interactions"`

	Metadata map[string]interface{} `json:"metadata"`

	Options PactSerialisationOptionsV2 `json:"-"`
}

func pactInteractionFromV2Interaction(interaction InteractionV2) pactInteractionV2 {
	return pactInteractionV2{
		Description: interaction.Description,
		State:       interaction.State,
		Request: pactRequestV2{
			Method:        interaction.Request.Method,
			Body:          interaction.Request.Body,
			Headers:       interaction.Request.Headers,
			MatchingRules: make(rule),
		},
		Response: pactResponseV2{
			Status:        interaction.Response.Status,
			Body:          interaction.Response.Body,
			Headers:       interaction.Response.Headers,
			MatchingRules: make(rule),
		},
	}
}

func mergeRules(a rule, b rule) {
	for k, v := range b {
		a[k] = v
	}
}

func (p *pactFileV2) generateV2PactFile() *pactFileV2 {
	for _, interaction := range p.interactions {
		fmt.Printf("Serialising interaction: %+v \n", *interaction)
		serialisedInteraction := pactInteractionFromV2Interaction(*interaction)

		var requestBodyMatchingRules, requestHeaderMatchingRules, requestQueryMatchingRules, responseBodyMatchingRules, responseHeaderMatchingRules rule
		var requestQuery map[string]interface{}

		requestQuery, requestQueryMatchingRules = buildPartV2(interaction.Request.Query, "$.query")
		serialisedInteraction.Request.Headers, requestHeaderMatchingRules = buildPartV2(interaction.Request.Headers, "$.headers")
		serialisedInteraction.Request.Body, requestBodyMatchingRules = buildPartV2(interaction.Request.Body, "$.body")
		serialisedInteraction.Response.Body, responseBodyMatchingRules = buildPartV2(interaction.Response.Body, "$.body")
		serialisedInteraction.Response.Headers, responseHeaderMatchingRules = buildPartV2(interaction.Response.Headers, "$.headers")

		buildPactRequestQueryV2(requestQuery, interaction, &serialisedInteraction, p.Options)
		buildPactRequestPathV2(interaction, &serialisedInteraction)
		mergeRules(serialisedInteraction.Request.MatchingRules, requestQueryMatchingRules)
		mergeRules(serialisedInteraction.Request.MatchingRules, requestHeaderMatchingRules)
		mergeRules(serialisedInteraction.Request.MatchingRules, requestBodyMatchingRules)
		mergeRules(serialisedInteraction.Response.MatchingRules, responseBodyMatchingRules)
		mergeRules(serialisedInteraction.Response.MatchingRules, responseHeaderMatchingRules)

		fmt.Printf("appending interaction: %+v \n", serialisedInteraction)
		p.Interactions = append(p.Interactions, serialisedInteraction)
	}

	return p
}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

func recurseMapType(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules rule) (string, map[string]interface{}, rule) {
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
			_, body, matchingRules = buildPactPartV2(k.String(), v.Interface(), copyMap(body), path, matchingRules)
		} else {
			_, body[key], matchingRules = buildPactPartV2(k.String(), v.Interface(), entry, path, matchingRules)
		}
	}

	return path, body, matchingRules
}

func buildPartV2(value interface{}, path string) (map[string]interface{}, rule) {
	_, o, matchingRules := buildPactPartV2("", value, make(object), path, make(rule))

	return o, matchingRules
}

// Recurse the Matcher tree and buildPactPartV2 up an example body and set of matchers for
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
// Returns path, body, matchingRules
func buildPactPartV2(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules rule) (string, map[string]interface{}, rule) {
	log.Println("[TRACE] generate pact => key:", key, ", body:", body, ", value:", value, ", path:", path)

	switch t := value.(type) {

	case MatcherV2:
		switch t.Type() {

		case arrayMinLikeMatcher:
			log.Println("[TRACE] generate pact: ArrayMinLikeMatcher")
			times := 1

			m := t.(eachLike)
			arrayMap := make(map[string]interface{})
			minArray := make([]interface{}, m.Min)

			builtPath := path + buildPath(key, allListItems)
			buildPactPartV2("0", t.GetValue(), arrayMap, builtPath, matchingRules)
			log.Println("[TRACE] generate pact: ArrayMinLikeMatcher =>", builtPath)
			matchingRules[path+buildPath(key, "")] = m.MatchingRule()

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
			_, body, matchingRules = recurseMapType(key, t.GetValue().(StructMatcher), body, path, matchingRules)

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
			buildPactPartV2(k, el, arrayMap, builtPath, matchingRules)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

		// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}, MapMatcher, QueryMatcher:
		log.Println("[TRACE] generate pact: MapMatcher")
		_, body, matchingRules = recurseMapType(key, t, body, path, matchingRules)

	case MatcherV3:
		log.Fatalf("error: v3 matcher '%+s' provided to to a v2 specification. This will lead to inconsistent results", reflect.TypeOf(value))

	// Primitives (terminal cases)
	default:
		log.Printf("[TRACE] generate pact: unknown type or primitive (%+v): %+v\n", reflect.TypeOf(t), value)
		body[key] = value
	}

	log.Printf("[TRACE] generate pact => returning body: %+v\n", body)

	return path, body, matchingRules
}

// V2 query strings are stored as strings
// "age=30&children=Mary+Jane&children=James"
//
// * Are these two things equivalent?
//     * baz[]=bat
//     * baz=bat
// * Are these two things equivalent?
//     * baz[]=bat&baz[]=bar
//     * baz=bat&baz=bar
// See https://stackoverflow.com/questions/6243051/how-to-pass-an-array-within-a-query-string
// TODO: allow a specific generator to be provided?
func buildPactRequestQueryV2(input map[string]interface{}, sourceInteraction *InteractionV2, destInteraction *pactInteractionV2, options PactSerialisationOptionsV2) {
	var parts []string

	for k, v := range input {
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			slice := v.([]interface{})
			length := len(slice)

			for _, data := range slice {
				switch options.QueryStringStyle {
				case Array:
					if length == 1 {
						parts = append(parts, fmt.Sprintf("%s=%s", k, data))
					} else {
						parts = append(parts, fmt.Sprintf("%s[]=%s", k, data))
					}
				case AlwaysArray:
					parts = append(parts, fmt.Sprintf("%s[]=%s", k, data))
				default:
					parts = append(parts, fmt.Sprintf("%s=%s", k, data))
				}
			}
		default:
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}

	destInteraction.Request.Query = strings.Join(parts, "&")
}

func buildPactRequestPathV2(sourceInteraction *InteractionV2, destInteraction *pactInteractionV2) {
	fmt.Printf("[DEBUG] path matching rule %+v\n", sourceInteraction.Request.Path)

	switch val := sourceInteraction.Request.Path.(type) {
	case String:
		destInteraction.Request.Path = val.GetValue().(string)
	case MatcherV2:
		switch val.Type() {
		case likeMatcher, regexMatcher:
			destInteraction.Request.MatchingRules["$.path"] = sourceInteraction.Request.Path.MatchingRule()
			destInteraction.Request.Path = val.GetValue().(string)
		}
	default:
		delete(destInteraction.Request.MatchingRules, "path")
		log.Print("[WARN] ignoring unsupported matcher for request path:", val)
	}
}

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
