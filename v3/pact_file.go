package v3

import (
	"log"

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

var pactGoMetadata = map[string]interface{}{
	"pactGo": map[string]string{
		"version": version.Version,
	},
}

type Pacticipant struct {
	Name string `json:"name"`
}

// newPactFileV2 generates a v2 formated pact file from the given interactions
func newPactFileV2(consumer string, provider string, interactions []*InteractionV2, options PactSerialisationOptionsV2) pactFileV2 {
	p := pactFileV2{
		Interactions:         make([]pactInteractionV2, 0),
		interactions:         interactions,
		Metadata:             pactGoMetadata,
		Consumer:             Pacticipant{consumer},
		Provider:             Pacticipant{provider},
		SpecificationVersion: V2,
		Options:              options,
	}

	p.generateV2PactFile()

	pactGoMetadata["pactSpecification"] = map[string]interface{}{
		"version": V2,
	}

	return p
}

// newPactFileV3 generates a v3 formated pact file from the given interactions
func newPactFileV3(consumer string, provider string, interactions []*InteractionV3, messages []*Message) pactFileV3 {
	log.Println("[DEBUG] creating v3 pact file")
	p := pactFileV3{
		Interactions:         make([]pactInteractionV3, 0),
		interactions:         interactions,
		Messages:             make([]pactMessageV3, 0),
		messages:             messages,
		Metadata:             pactGoMetadata,
		Consumer:             Pacticipant{consumer},
		Provider:             Pacticipant{provider},
		SpecificationVersion: V3,
	}

	p.generateV3PactFile()

	pactGoMetadata["pactSpecification"] = map[string]interface{}{
		"version": V3,
	}

	return p
}
