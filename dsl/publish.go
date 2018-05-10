package dsl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pact-foundation/pact-go/types"
)

// PactFile is a simple representation of a Pact file to be able to
// parse Consumer/Provider from the file.
type PactFile struct {
	// The API Consumer name
	Consumer PactName `json:"consumer"`

	// The API Provider name
	Provider PactName `json:"provider"`
}

// PactName represents the name fields in the PactFile.
type PactName struct {
	Name string `json:"name"`
}

// Publisher is the API to send Pact files to a Pact Broker.
type Publisher struct {
	request types.PublishRequest
	client  *http.Client
}

// validate the publish requests.
func (p *Publisher) validate() error {
	log.Println("[DEBUG] pact publisher: validate")

	// At least 1 Pact URL
	if len(p.request.PactURLs) == 0 {
		return errors.New("PactURLs is mandatory")
	}

	// Validate that the files exist on the system
	var err error
	for _, url := range p.request.PactURLs {
		// Only check local files
		if !strings.HasPrefix(url, "http") {
			if _, err = os.Stat(url); err != nil {
				return err
			}
		}
	}

	if p.request.PactBroker == "" {
		return errors.New("PactBroker is mandatory")
	}

	if p.request.ConsumerVersion == "" {
		return errors.New("ConsumerVersion is mandatory")
	}

	if (p.request.BrokerUsername != "" && p.request.BrokerPassword == "") || (p.request.BrokerUsername == "" && p.request.BrokerPassword != "") {
		return errors.New("Must provide both or none of BrokerUsername and BrokerPassword")
	}

	return nil
}

// call sends a message to the Pact Broker.
func (p *Publisher) call(method string, url string, content []byte) error {
	if p.client == nil {
		p.client = &http.Client{}
	}
	var req *http.Request
	var err error
	req, err = http.NewRequest(method, url, bytes.NewReader(content))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if p.request.BrokerUsername != "" && p.request.BrokerPassword != "" {
		req.SetBasicAuth(p.request.BrokerUsername, p.request.BrokerPassword)
	}

	res, err := p.client.Do(req)
	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	log.Printf("[DEBUG] pact publisher response Body: %s\n", responseBody)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.New(string(responseBody))
	}
	return err
}

// readPactFile reads Pact files from local or remote sources.
func (p *Publisher) readPactFile(url string) (*PactFile, []byte, error) {
	log.Println("[DEBUG] pact publisher: readPactFile", url)
	if strings.HasPrefix(url, "http") {
		return p.readRemotePactFile(url)
	}
	return p.readLocalPactFile(url)
}

// readLocalPactFile reads a local Pact file.
func (p *Publisher) readLocalPactFile(file string) (*PactFile, []byte, error) {
	log.Println("[DEBUG] pact publisher: readLocalPactFile")
	_, err := os.Stat(file)
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}

	f, err := p.unmarshal(data)
	return f, data, err
}

// unmarshal creates and validates a subset of a Pact File.
func (p *Publisher) unmarshal(content []byte) (*PactFile, error) {
	log.Println("[DEBUG] pact publisher: unmarshal")
	var pactFile PactFile
	err := json.Unmarshal(content, &pactFile)

	if &pactFile == nil || pactFile.Consumer.Name == "" || pactFile.Provider.Name == "" {
		return &pactFile, errors.New("Invalid Pact file - cannot find the Consumer and Provider name")
	}
	return &pactFile, err
}

// readRemotePactFile reads a remote Pact file from an http(s) server.
func (p *Publisher) readRemotePactFile(file string) (*PactFile, []byte, error) {
	log.Println("[DEBUG] pact publisher: read remote pact file", file)
	res, err := http.Get(file)
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, nil, err
	}

	f, err := p.unmarshal(data)
	return f, data, err
}

// Publish sends the Pacts to a broker, optionally tagging them
func (p *Publisher) Publish(request types.PublishRequest) error {
	log.Println("[DEBUG] pact publisher: publish pact")
	p.request = request

	for _, url := range request.PactURLs {
		file, data, err := p.readPactFile(url)
		if err != nil {
			return err
		}

		endpoint := fmt.Sprintf("%s/pacts/provider/%s/consumer/%s/version/%s", request.PactBroker, file.Provider.Name, file.Consumer.Name, request.ConsumerVersion)
		log.Println("[DEBUG] pact publisher: putting Pact on endpoint:", endpoint)
		err = p.call("PUT", endpoint, data)
		if err != nil {
			return err
		}

		p.tagRequest(file.Consumer.Name, request)
	}

	return nil
}

// tag one or more Pact files
func (p *Publisher) tagRequest(consumerName string, request types.PublishRequest) error {
	log.Println("[DEBUG] pact publisher: tagging pacts...")
	for _, tag := range request.Tags {
		endpoint := fmt.Sprintf("%s/pacticipants/%s/versions/%s/tags/%s", request.PactBroker, consumerName, request.ConsumerVersion, tag)
		log.Println("[DEBUG] pact publisher: tagging Pact:", endpoint)
		err := p.call("PUT", endpoint, []byte{})
		if err != nil {
			return err
		}
	}

	return nil
}

// SetClient allows dsl users to configure the http.Client used when publishing Pacts
func (p *Publisher) SetClient(client *http.Client) {
	p.client = client
}
