package v3

import (
	version "github.com/pact-foundation/pact-go/command"
)

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

// newPactFileV2 generates a v2 formated pact file from the given interactions
func newPactFileV2(consumer string, provider string, interactions []*InteractionV2) pactFileV2 {
	p := pactFileV2{
		Interactions:         make([]pactInteractionV2, 0),
		interactions:         interactions,
		Metadata:             pactGoMetadata,
		Consumer:             consumer,
		Provider:             provider,
		SpecificationVersion: V2,
	}

	p.generateV2PactFile()

	pactGoMetadata["pactSpecification"] = map[string]interface{}{
		"version": V2,
	}

	return p
}

// newPactFileV3 generates a v3 formated pact file from the given interactions
func newPactFileV3(consumer string, provider string, interactions []*InteractionV3) pactFileV3 {
	p := pactFileV3{
		Interactions:         make([]pactInteractionV3, 0),
		interactions:         interactions,
		Metadata:             pactGoMetadata,
		Consumer:             consumer,
		Provider:             provider,
		SpecificationVersion: V3,
	}

	p.generateV3PactFile()

	pactGoMetadata["pactSpecification"] = map[string]interface{}{
		"version": V3,
	}

	return p
}
