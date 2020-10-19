package v3

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"reflect"

	"github.com/spf13/afero"
)

type object map[string]interface{}

// ruleSet are the set of matchers to apply at a path, and the logical operation in which to apply them
// TODO: this is actually more typed than this
//       once we understand the model better, let's make it more type-safe
type ruleSet map[string]matchers
type generators map[string]rule

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
type generatorV3 = struct {
	Body    generators `json:"body,omitempty"`
	Headers generators `json:"headers,omitempty"`
	Query   generators `json:"query,omitempty"`
	Path    matchers   `json:"path,omitempty"`
}

type pactRequestV3 struct {
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Query         map[string][]string `json:"query,omitempty"`
	Headers       interface{}         `json:"headers,omitempty"`
	Body          interface{}         `json:"body,omitempty"`
	MatchingRules matchingRuleV3      `json:"matchingRules,omitempty"`
	Generators    generatorV3         `json:"generators"`
}

type pactResponseV3 struct {
	Status        int            `json:"status"`
	Headers       interface{}    `json:"headers,omitempty"`
	Body          interface{}    `json:"body,omitempty"`
	MatchingRules matchingRuleV3 `json:"matchingRules,omitempty"`
}

type pactInteractionV3 struct {
	Description string            `json:"description"`
	States      []ProviderStateV3 `json:"providerStates,omitempty"`
	Request     pactRequestV3     `json:"request"`
	Response    pactResponseV3    `json:"response"`
}

type pactMessageV3 struct {
	// Message body
	Content interface{} `json:"contents,omitempty"`

	// Provider state to be written into the Pact file
	States []ProviderStateV3 `json:"providerStates,omitempty"`

	// Message metadata
	Metadata interface{} `json:"metadata,omitempty"`

	// Description to be written into the Pact file
	Description string `json:"description"`

	MatchingRules matchingRuleV3 `json:"matchingRules,omitempty"`
}

// pactFileV3 is what will be serialised to the Pactfile in the request body examples and matching rules
// given a structure containing matchers.
type pactFileV3 struct {
	// Consumer is the name of the Consumer/Client.
	Consumer Pacticipant `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider Pacticipant `json:"provider"`

	// SpecificationVersion is the version of the Pact Spec this implementation supports
	SpecificationVersion SpecificationVersion `json:"-"`

	interactions []*InteractionV3

	messages []*Message

	// Messages are the v3 message interaction types
	Messages []pactMessageV3 `json:"messages,omitempty"`

	// Interactions are all of the request/response expectations, with matching rules and generators
	Interactions []pactInteractionV3 `json:"interactions,omitempty"`

	Metadata map[string]interface{} `json:"metadata"`
}

func pactInteractionFromV3Interaction(interaction InteractionV3) pactInteractionV3 {
	return pactInteractionV3{
		Description: interaction.Description,
		States:      interaction.States,
		Request: pactRequestV3{
			Method: interaction.Request.Method,
			Generators: generatorV3{
				Body:    make(generators),
				Headers: make(generators),
				Query:   make(generators),
			},
			MatchingRules: matchingRuleV3{
				Body:    make(ruleSet),
				Headers: make(ruleSet),
				Query:   make(ruleSet),
			},
		},
		Response: pactResponseV3{
			Status: interaction.Response.Status,
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

		if interaction.Request.Query != nil {
			requestQuery, serialisedInteraction.Request.MatchingRules.Query, serialisedInteraction.Request.Generators.Query = buildPart(interaction.Request.Query)
		}
		if interaction.Request.Headers != nil {
			serialisedInteraction.Request.Headers, serialisedInteraction.Request.MatchingRules.Headers, serialisedInteraction.Request.Generators.Headers = buildPart(interaction.Request.Headers)
		}
		if interaction.Request.Body != nil {
			serialisedInteraction.Request.Body, serialisedInteraction.Request.MatchingRules.Body, serialisedInteraction.Request.Generators.Body = buildPart(interaction.Request.Body)
		}
		if interaction.Response.Headers != nil {
			serialisedInteraction.Response.Headers, serialisedInteraction.Response.MatchingRules.Headers, _ = buildPart(interaction.Response.Headers)
		}
		if interaction.Response.Body != nil {
			serialisedInteraction.Response.Body, serialisedInteraction.Response.MatchingRules.Body, _ = buildPart(interaction.Response.Body)
		}
		if interaction.Request.Query != nil {
			buildQueryV3(requestQuery, interaction, &serialisedInteraction)
		}
		buildPactPathV3(interaction, &serialisedInteraction)

		p.Interactions = append(p.Interactions, serialisedInteraction)
	}
	for _, message := range p.messages {
		serialisedMessage := pactMessageV3{
			Description: message.Description,
			States:      message.States,
		}

		serialisedMessage.Content, serialisedMessage.MatchingRules.Body, _ = buildPart(message.Content)
		serialisedMessage.Metadata, _, _ = buildPart(message.Metadata)

		p.Messages = append(p.Messages, serialisedMessage)
	}

	return p
}

func recurseMapTypeV3(key string, value interface{}, body object, path string,
	matchingRules ruleSet, generators generators) (string, object, ruleSet, generators) {
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
	return matchers{
		Combine:  AND,
		Matchers: []rule{r},
	}
}

func buildPart(value interface{}) (object, ruleSet, generators) {
	_, o, matchingRules, generators := buildPactPartV3("", value, make(object), "$", make(ruleSet), make(generators))
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
	matchingRules ruleSet, generators generators) (string, object, ruleSet, generators) {
	log.Println("[TRACE] generate pact => key:", key, ", body:", body, ", value:", value, ", path:", path)

	switch t := value.(type) {

	case MatcherV3:
		switch t.Type() {

		case decimalMatcher, integerMatcher, nullMatcher, equalityMatcher, includesMatcher, stringGeneratorMatcher:
			log.Println("[TRACE] generate pact: decimal/integer/null matcher")
			builtPath := path + buildPath(key, "")
			body[key] = t.GetValue()
			log.Println("[TRACE] generate pact: decimal/integer/null matcher => ", builtPath)
			matchingRules[builtPath] = wrapMatchingRule(t.MatchingRule())

			if g, ok := t.(generator); ok {
				log.Println("[TRACE] have generator:", g.Generator())
				generators[builtPath] = g.Generator()
			}

		case arrayMinMaxLikeMatcher:
			times := 1
			m := t.(minMaxLike)
			if m.Max > 0 {
				times = m.Max
			} else if m.Min > 0 {
				times = m.Min
			}

			log.Println("[TRACE] generate pact: ArrayMinMaxLikeMatcher")

			arrayMap := make(map[string]interface{})
			minArray := make([]interface{}, times)

			builtPath := path + buildPath(key, allListItems)
			buildPactPartV3("0", t.GetValue(), arrayMap, builtPath, matchingRules, generators)
			log.Println("[TRACE] generate pact: ArrayMinMaxLikeMatcher =>", builtPath, t.MatchingRule())
			matchingRules[path+buildPath(key, "")] = wrapMatchingRule(t.MatchingRule())

			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}

			body[key] = minArray
			fmt.Printf("Updating body: %+v, minArray: %+v", body, minArray)
			path = path + buildPath(key, "")

			// TODO: this is duplicated with below. Extract into common functions
			// It's only needed because the Match function can't use v2 matchers, due to its type
		case regexMatcher, likeMatcher:
			log.Println("[TRACE] generate pact: Regex/LikeMatcher")
			builtPath := path + buildPath(key, "")
			body[key] = t.GetValue()
			log.Println("[TRACE] generate pact: Regex/LikeMatcher =>", builtPath)
			matchingRules[builtPath] = wrapMatchingRule(t.MatchingRule())

			// TODO: this is duplicated with below. Extract into common functions
		case structTypeMatcher:
			log.Println("[TRACE] generate pact: StructTypeMatcher")
			_, body, matchingRules, generators = recurseMapTypeV3(key, t.GetValue().(StructMatcher), body, path, matchingRules, generators)

		default:
			log.Fatalf("unexpected matcher (%s) for current specification format (3.0.0)", t.Type())
		}
	case MatcherV2:
		switch t.Type() {

		case arrayMinLikeMatcher, arrayMinMaxLikeMatcher:
			log.Println("[TRACE] generate pact: ArrayMinLikeMatcher")
			m := t.(eachLike)
			times := m.Min

			arrayMap := make(map[string]interface{})
			minArray := make([]interface{}, times)

			builtPath := path + buildPath(key, allListItems)
			buildPactPartV3("0", t.GetValue(), arrayMap, builtPath, matchingRules, generators)
			log.Println("[TRACE] generate pact: ArrayMinLikeMatcher =>", builtPath)
			matchingRules[path+buildPath(key, "")] = wrapMatchingRule(m.MatchingRule())

			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}

			// TODO: I think this assignment is working, but the next step seems to recurse again and this never writes
			// probably just a bad terminal case handling?
			body[key] = minArray
			fmt.Printf("Updating body: %+v, minArray: %+v", body, minArray)
			path = path + buildPath(key, "")

			// TODO: this is duplicated above. Extract into common functions
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

type pactReaderV3 interface {
	read(string) (pactFileV3, error)
}

type pactWriterV3 interface {
	write(string, pactFileV3) error
}

type pactFileV3ReaderWriter struct {
	fs afero.Fs
}

func defaultPactFileV3ReaderWriter() pactFileV3ReaderWriter {
	return pactFileV3ReaderWriter{
		fs: afero.NewOsFs(),
	}
}

func pactFilePath(dir string, pactfile pactFileV3) string {
	return path.Join(dir, fmt.Sprintf("%s-%s.json", pactfile.Consumer.Name, pactfile.Provider.Name))
}

func (p *pactFileV3ReaderWriter) update(dir string, pactfile pactFileV3) error {
	// Check if file already exists
	existing, err := p.read(pactFilePath(dir, pactfile))
	if err != nil {
		log.Println("[INFO] existing pact file not found or error reading:", err)
	}

	combined, err := mergePactFiles(existing, pactfile)
	if err != nil {
		log.Println("[ERROR] unable to merge message pacts into existing pact file: ", err)
	}

	return p.write(dir, combined)
}

func (p *pactFileV3ReaderWriter) write(dir string, pactfile pactFileV3) error {
	bytes, err := json.Marshal(pactfile)
	if err != nil {
		return err
	}
	return afero.WriteFile(p.fs, pactFilePath(dir, pactfile), bytes, 0644)
}

func (p *pactFileV3ReaderWriter) read(file string) (pactFileV3, error) {
	bytes, err := afero.ReadFile(p.fs, file)
	if err != nil {
		return pactFileV3{}, err
	}

	var f pactFileV3
	err = json.Unmarshal(bytes, &f)

	return f, err
}

// Merge two pact files together
//
// Merging a pact file rules:
// Any interactions in the original pact file are appended to the new one
// ... this means any non-interaction info (e.g. metadata etc.) will also be replaced
// by the updated case
// Errors if attempt to add messages + interactions in the same file
func mergePactFiles(orig pactFileV3, updated pactFileV3) (pactFileV3, error) {
	if len(orig.Messages) > 0 && len(updated.Interactions) > 0 {
		return orig, fmt.Errorf("attempting to merge HTTP interactions to an existing contract containing messages, cannot have both")
	}
	if len(orig.Interactions) > 0 && len(updated.Messages) > 0 {
		return orig, fmt.Errorf("attempting to merge message interactions to an existing contract containing HTTP interactions, cannot have both")
	}

	if len(orig.Messages) > 0 {
		updated.Messages = append(orig.Messages, updated.Messages...)
	}
	if len(orig.Interactions) > 0 {
		updated.Interactions = append(orig.Interactions, updated.Interactions...)
	}

	// TODO: check consumer/provider match?
	// TODO: check for conflicting messages?
	// TODO: merge metadata?

	return updated, nil
}
