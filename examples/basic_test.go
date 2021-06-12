package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/pact-foundation/pact-go/v2/sugar"
	"github.com/stretchr/testify/assert"
)

func TestUserAPIClient(t *testing.T) {
	// Specify the two applications in the integration we are testing
	// NOTE: this can usually be extracted out of the individual test for re-use)
	mockProvider, err := NewV2Pact(MockHTTPProviderConfig{
		Consumer: "UserAPIConsumer",
		Provider: "UserAPI",
	})
	assert.NoError(t, err)

	// Arrange: Setup our expected interactions
	mockProvider.
		AddInteraction().
		Given("A user with ID 10 exists").
		UponReceiving("A request for User 10").
		WithRequest("GET", S("/user/10")).
		WillRespondWith(200).
		WithBodyMatch(&User{})

	// Act: test our API client behaves correctly
	err = mockProvider.ExecuteTest(func(config MockServerConfig) error {
		// Initialise the API client and point it at the Pact mock server
		client := newClient(config.Host, config.Port)

		// Execute the API client
		user, err := client.GetUser("10")

		// Assert: check the result
		assert.NoError(t, err)
		assert.Equal(t, 10, user.ID)

		return err
	})
	assert.NoError(t, err)
}

// User domain model
type User struct {
	ID       int    `json:"id" pact:"example=10"`
	Name     string `json:"name" pact:"example=Billy"`
	LastName string `json:"lastName" pact:"example=Sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
}

// User API Client to test
type userAPIClient struct {
	port int
	host string
}

func newClient(host string, port int) *userAPIClient {
	return &userAPIClient{
		host: host,
		port: port,
	}
}

func (u *userAPIClient) GetUser(id string) (*User, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d:%s%s", u.host, u.port, "/user/", id))

	if err != nil {
		return nil, err
	}

	user := new(User)
	err = json.NewDecoder(resp.Body).Decode(user)
	if err != nil {
		return nil, err
	}

	return user, err
}
