package v3

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

type pactRequestV3 struct {
	Method        string       `json:"method"`
	Path          Matcher      `json:"path"`
	Query         MapMatcher   `json:"query,omitempty"`
	Headers       MapMatcher   `json:"headers,omitempty"`
	Body          interface{}  `json:"body"`
	MatchingRules matchingRule `json:"matchingRules,omitempty"`
	Generators    generator    `json:"generators"`
}

type pactResponseV3 struct {
	Status        int          `json:"status"`
	Headers       MapMatcher   `json:"headers,omitempty"`
	Body          interface{}  `json:"body,omitempty"`
	MatchingRules matchingRule `json:"matchingRules,omitempty"`
	Generators    generator    `json:"generators"`
}

type pactInteractionV3 struct {
	Description string         `json:"description"`
	State       string         `json:"providerState,omitempty"`
	Request     pactRequestV3  `json:"request"`
	Response    pactResponseV3 `json:"response"`
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

	interactions []*Interaction

	// Interactions are all of the request/response expectations, with matching rules and generators
	Interactions []pactInteractionV3 `json:"interactions"`

	Metadata map[string]interface{} `json:"metadata"`
}

func pactInteractionFromV3Interaction(interaction Interaction) pactInteractionV3 {
	return pactInteractionV3{
		Description: interaction.Description,
		State:       interaction.State,
		Request: pactRequestV3{
			Method:  interaction.Request.Method,
			Body:    interaction.Request.Body,
			Headers: interaction.Request.Headers,
			Query:   interaction.Request.Query,
			Path:    interaction.Request.Path,
			// Generators:    make(generatorType),
			MatchingRules: generator{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
		Response: pactResponseV3{
			Status:  interaction.Response.Status,
			Body:    interaction.Response.Body,
			Headers: interaction.Response.Headers,
			// Generators:    make(generatorType),
			MatchingRules: matchingRule{
				Body:    make(ruleValue),
				Headers: make(ruleValue),
				Path:    make(ruleValue),
				Query:   make(ruleValue),
			},
		},
	}
}

// func (p *pactFileV3) generatev3PactFile() *pactFileV3 {
// 	for _, interaction := range p.interactions {
// 		fmt.Printf("Serialising interaction: %+v \n", *interaction)
// 		serialisedInteraction := PactInteractionFromV3Interaction(*interaction)

// 		// TODO: haven't done matchers for headers, path and status code
// 		_, serialisedInteraction.Request.Body, serialisedInteraction.Request.MatchingRules, _ = buildPactBody("", interaction.Request.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))
// 		_, serialisedInteraction.Response.Body, serialisedInteraction.Response.MatchingRules, _ = buildPactBody("", interaction.Response.Body, make(map[string]interface{}), "$.body", make(ruleValue), make(ruleValue))

// 		// v3
// 		// serialisedInteraction.Request.MatchingRules = requestBodyMatchingRules
// 		// serialisedInteraction.Response.MatchingRules = responseBodyMatchingRules

// 		// v3 only
// 		// serialisedInteraction.Request.MatchingRules.Body = requestBodyMatchingRules
// 		// serialisedInteraction.Response.MatchingRules.Body = responseBodyMatchingRules

// 		// TODO
// 		buildPactHeaders()
// 		buildPactQuery()
// 		buildPactPath()

// 		fmt.Printf("appending interaction: %+v \n", serialisedInteraction)
// 		p.Interactions = append(p.Interactions, serialisedInteraction)
// 	}

// 	return p
// }
