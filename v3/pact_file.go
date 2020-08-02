package v3

import (
	"fmt"

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

// SpecificationVersion is used to determine the current specification version
type SpecificationVersion string

const (
	// V2 signals the use of version 2 of the pact spec
	V2 SpecificationVersion = "2.0.0"

	// V3 signals the use of version 3 of the pact spec
	V3 = "3.0.0"
)

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
}

func NewPactFile(consumer string, provider string, interactions []*Interaction, specificationVersion SpecificationVersion) pactFileV2 {
	p := pactFileV2{
		Interactions:         make([]pactInteractionV2, 0),
		interactions:         interactions,
		Metadata:             pactGoMetadata,
		Consumer:             consumer,
		Provider:             provider,
		SpecificationVersion: specificationVersion,
	}

	if specificationVersion == V2 {
		p.generatev2PactFile()
	} else {
		panic(fmt.Sprintf("specification version not supported: %+v", specificationVersion))
	}

	pactGoMetadata["pactSpecification"] = map[string]interface{}{
		"version": specificationVersion,
	}

	return p
}
