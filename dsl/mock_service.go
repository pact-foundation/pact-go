package dsl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// PactMockService is the HTTP interface to setup the Pact Mock Service
// See https://github.com/bethesque/pact-mock_service and
// https://gist.github.com/bethesque/9d81f21d6f77650811f4.
type PactMockService struct {
	// BaseURL is the base host for the Pact Mock Service.
	BaseURL string

	// Consumer name.
	Consumer string

	// Provider name.
	Provider string
}

// call sends a message to the Pact service
func (m *PactMockService) call(method string, url string, object interface{}) error {
	body, err := json.Marshal(object)
	if err != nil {
		return err
	}
	client := &http.Client{}
	var req *http.Request
	if method == "POST" {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return err
	}

	req.Header.Set("X-Pact-Mock-Service", "true")
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	log.Printf("Response Body: %s\n", responseBody)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.New(string(responseBody))
	}
	return err
}

// DeleteInteractions removes any previous Mock Service Interactions.
func (m *PactMockService) DeleteInteractions() error {
	url := fmt.Sprintf("%s/interactions", m.BaseURL)
	return m.call("DELETE", url, nil)
}

// AddInteraction adds a new Pact Mock Service interaction.
func (m *PactMockService) AddInteraction(interaction *Interaction) error {
	url := fmt.Sprintf("%s/interactions", m.BaseURL)
	return m.call("POST", url, interaction)
}

// Verify confirms that all interactions were called.
func (m *PactMockService) Verify() error {
	url := fmt.Sprintf("%s/interactions/verification", m.BaseURL)
	return m.call("GET", url, nil)
}

// WritePact writes the pact file to disk.
func (m *PactMockService) WritePact() error {
	if m.Consumer == "" || m.Provider == "" {
		return errors.New("Consumer and Provider name need to be provided")
	}
	pact := map[string]map[string]string{
		"consumer": map[string]string{
			"name": m.Consumer,
		},
		"provider": map[string]string{
			"name": m.Provider,
		},
	}

	url := fmt.Sprintf("%s/pact", m.BaseURL)
	return m.call("POST", url, pact)
}
