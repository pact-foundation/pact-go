package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

type S = matchers.S

func TestProductAPIClient(t *testing.T) {
	// Specify the two applications in the integration we are testing
	// NOTE: this can usually be extracted out of the individual test for re-use)
	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoProductAPIConsumer",
		Provider: "PactGoProductAPI",
	})
	assert.NoError(t, err)

	// Arrange: Setup our expected interactions
	err = mockProvider.
		AddInteraction().
		Given("A product with ID 10 exists").
		UponReceiving("A request for Product 10").
		WithRequest("GET", "/products/10").
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.BodyMatch(&Product{})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			// Act: test our API client behaves correctly
			// Initialise the API client and point it at the Pact mock server
			client := newClient(config.Host, config.Port)

			// Execute the API client
			product, err := client.GetProduct("10")

			// Assert: check the result
			assert.NoError(t, err)
			assert.Equal(t, 10, product.ID)

			return err
		})

	assert.NoError(t, err)
}

// Product domain model
type Product struct {
	ID    int    `json:"id" pact:"example=10"`
	Name  string `json:"name" pact:"example=Billy"`
	Price string `json:"price" pact:"example=23.33"`
}

// Product API Client to test
type productAPIClient struct {
	port int
	host string
}

func newClient(host string, port int) *productAPIClient {
	return &productAPIClient{
		host: host,
		port: port,
	}
}

func (u *productAPIClient) GetProduct(id string) (*Product, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d:%s%s", u.host, u.port, "/products/", id))

	if err != nil {
		return nil, err
	}

	product := new(Product)
	err = json.NewDecoder(resp.Body).Decode(product)
	if err != nil {
		return nil, err
	}

	return product, err
}
