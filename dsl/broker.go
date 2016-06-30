package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pact-foundation/pact-go/types"
)

var pactURLPattern = "%s/pacts/provider/%s/latest"
var pactURLPatternWithTag = "%s/pacts/provider/%s/latest/%s"

// PactLink represents the Pact object in the HAL response.
type PactLink struct {
	Href  string `json:"href"`
	Title string `json:"title"`
	Name  string `json:"name"`
}

// HalLinks represents the _links key in a HAL document.
type HalLinks struct {
	Pacts []PactLink `json:"pacts"`
}

// HalDoc is a simple representation of the HAL response from a Pact Broker.
type HalDoc struct {
	Links HalLinks `json:"_links"`
}

// findConsumers navigates a Pact Broker's HAL system to find consumers
// based on the latest Pacts or using tags.
func findConsumers(provider string, request *types.VerifyRequest) error {
	log.Println("findConsumers")

	// Check for Broker-based requests and if so, fetch from Broker before
	// verifying.

	// 2 Scenarios:
	//   1. Ask for all 'latest' consumers.
	//   2. Pass a set of tags (e.g. 'latest' and 'prod') and find all consumers that match.

	// 1. Find all consumers
	// 2. Construct all URLs with relevent tags
	// 3. Populate 'PactURLs'
	// 4. Send off to Daemon

	client := &http.Client{}
	var url string
	if len(request.Tags) > 0 {
		url = fmt.Sprintf(pactURLPatternWithTag, request.BrokerURL, provider, "dev")
	} else {
		url = fmt.Sprintf(pactURLPattern, request.BrokerURL, provider)
	}
	var req *http.Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/hal+json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	log.Printf("[DEBUG] pact broker response Body: %s\n", responseBody)

	var doc HalDoc
	err = json.Unmarshal(responseBody, &doc)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.New(string(responseBody))
	}

	if err != nil {
		return err
	}

	for _, p := range doc.Links.Pacts {
		request.PactURLs = append(request.PactURLs, p.Href)
	}

	return nil
}
