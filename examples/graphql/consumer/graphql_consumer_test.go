package consumer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	graphqlserver "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/example/starwars"
	"github.com/graph-gophers/graphql-go/relay"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/pact-foundation/pact-go/dsl"
	g "github.com/pact-foundation/pact-go/dsl/graphql"
	"github.com/stretchr/testify/assert"
)

func TestGraphQLConsumer(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "GraphQLConsumer",
		Provider: "GraphQLProvider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to get foo").
		WithRequest(*g.Interaction(g.Query{
			Method: "POST",
			Path:   dsl.String("/query"),
			Query: `query ($characterID:ID!){
				hero {
					id,
					name
				},
				character(id: $characterID)
				{
					name,
					friends{
						name,
						__typename
					},
					appearsIn
				}
			}`,
			// Operation: "SomeOperation", // if needed
			Variables: g.Variables{
				"characterID": "1003",
			},
		})).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body: g.Response{
				Data: heroQuery{
					Hero: hero{
						ID:   graphql.ID("1003"),
						Name: "Darth Vader",
					},
					Character: character{
						Name: "Darth Vader",
						AppearsIn: []graphql.String{
							"EMPIRE",
						},
						Friends: []friend{
							{
								Name:     "Wilhuff Tarkin",
								Typename: "friends",
							},
						},
					},
				},
			}})

		// assert on the response
	var test = func() error {
		res, err := executeQuery(fmt.Sprintf("http://localhost:%d", pact.Server.Port))

		fmt.Println(res)
		assert.NoError(t, err)
		assert.NotNil(t, res.Hero.ID)

		return nil
	}

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

func executeQuery(baseURL string) (heroQuery, error) {
	var q heroQuery

	// Set up a GraphQL server.
	schema, err := graphqlserver.ParseSchema(starwars.Schema, &starwars.Resolver{})
	if err != nil {
		return q, err
	}
	mux := http.NewServeMux()
	mux.Handle("/query", &relay.Handler{Schema: schema})

	client := graphql.NewClient(fmt.Sprintf("%s/query", baseURL), nil)

	variables := map[string]interface{}{
		"characterID": graphql.ID("1003"),
	}
	err = client.Query(context.Background(), &q, variables)
	if err != nil {
		return q, err
	}

	return q, nil
}

type hero struct {
	ID   graphql.ID     `json:"ID"`
	Name graphql.String `json:"Name"`
}
type friend struct {
	Name     graphql.String `json:"Name"`
	Typename graphql.String `json:"__typename" graphql:"__typename"`
}
type character struct {
	Name      graphql.String   `json:"Name"`
	Friends   []friend         `json:"Friends"`
	AppearsIn []graphql.String `json:"AppearsIn"`
}

type heroQuery struct {
	Hero      hero      `json:"Hero"`
	Character character `json:"character" graphql:"character(id: $characterID)"`
}
