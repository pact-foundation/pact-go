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

var (
	pactURLPattern        = "%s/pacts/provider/%s/latest"
	pactURLPatternWithTag = "%s/pacts/provider/%s/latest/%s"

	// ErrNoConsumers is returned when no consumer are not found for a provider.
	ErrNoConsumers = errors.New("no consumers found")

	// ErrUnauthorized represents a Forbidden (403).
	ErrUnauthorized = errors.New("unauthorized")
)

// PactLink represents the Pact object in the HAL response.
type PactLink struct {
	Href  string `json:"href"`
	Title string `json:"title"`
	Name  string `json:"name"`
}

// HalLinks represents the _links key in a HAL document.
type HalLinks struct {
	Pacts    []PactLink `json:"pb:pacts"`
	OldPacts []PactLink `json:"pacts"`
}

// HalDoc is a simple representation of the HAL response from a Pact Broker.
type HalDoc struct {
	Links HalLinks `json:"_links"`
}

// findConsumers navigates a Pact Broker's HAL system to find consumers
// based on the latest Pacts or using tags.
//
// There are 2 Scenarios:
//
//   1. Ask for all 'latest' consumers
//   2. Pass a set of tags (e.g. 'latest' and 'prod') and find all consumers
//      that match
func findConsumers(provider string, request *types.VerifyRequest) error {
	log.Println("[DEBUG] broker - find consumers for provider:", provider)

	client := &http.Client{}
	var urls []string
	pactURLs := make(map[string]string)

	if len(request.Tags) > 0 {
		for _, tag := range request.Tags {
			urls = append(urls, fmt.Sprintf(pactURLPatternWithTag, request.BrokerURL, provider, tag))
		}
	} else {
		urls = append(urls, fmt.Sprintf(pactURLPattern, request.BrokerURL, provider))
	}

	for _, url := range urls {
		var req *http.Request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Accept", "application/hal+json")

		if request.BrokerUsername != "" && request.BrokerPassword != "" {
			req.SetBasicAuth(request.BrokerUsername, request.BrokerPassword)
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		switch res.StatusCode {
		case 401:
			return ErrUnauthorized
		case 404:
			return ErrNoConsumers
		}

		responseBody, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		log.Printf("[DEBUG] pact broker response Body: %s\n", responseBody)

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return errors.New(string(responseBody))
		}

		var doc HalDoc
		err = json.Unmarshal(responseBody, &doc)
		if err != nil {
			return err
		}

		// Collapse results on the URL the pact links to
		for _, p := range doc.Links.Pacts {
			pactURLs[p.Href] = p.Href
		}

		// Ensure backwards compatability with old pacts
		// See https://github.com/pact-foundation/pact_broker/issues/209#issuecomment-390437990
		for _, p := range doc.Links.OldPacts {
			pactURLs[p.Href] = p.Href
		}
	}

	// Scrub out duplicate pacts across tags (e.g. 'latest' may equal 'prod' pact)
	for _, p := range pactURLs {
		request.PactURLs = append(request.PactURLs, p)
	}

	fmt.Println("[DEBUG] discovered pacts to verify: ", request.PactURLs)

	return nil
}
