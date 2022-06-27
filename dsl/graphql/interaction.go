package graphql

import (
	"fmt"
	"regexp"

	"github.com/pact-foundation/pact-go/dsl"
)

// Variables represents values to be substituted into the query
type Variables map[string]interface{}

// Query is the main implementation of the Pact interface.
type Query struct {
	// HTTP Headers
	Headers dsl.MapMatcher

	// Path to GraphQL endpoint
	Path dsl.Matcher

	// HTTP Query String
	QueryString dsl.MapMatcher

	// GraphQL Query
	Query string

	// GraphQL Variables
	Variables Variables

	// GraphQL Operation
	Operation string

	// GraphQL method (usually POST, but can be get with a query string)
	// NOTE: for query string users, the standard HTTP interaction should suffice
	Method string

	// Supports graphql extensions such as https://www.apollographql.com/docs/apollo-server/performance/apq/
	Extensions Extensions
}
type Extensions map[string]interface{}

// Specify the operation (if any)
func (r *Query) WithOperation(operation string) *Query {
	r.Operation = operation

	return r
}

// WithContentType overrides the default content-type (application/json)
// for the GraphQL Query
func (r *Query) WithContentType(contentType dsl.Matcher) *Query {
	r.setHeader("content-type", contentType)

	return r
}

// Specify the method (defaults to POST)
func (r *Query) WithMethod(method string) *Query {
	r.Method = method

	return r
}

// Given specifies a provider state. Optional.
func (r *Query) WithQuery(query string) *Query {
	r.Query = query

	return r
}

// Given specifies a provider state. Optional.
func (r *Query) WithVariables(variables Variables) *Query {
	r.Variables = variables

	return r
}

// Set the query extensions
func (r *Query) WithExtensions(extensions Extensions) *Query {
	r.Extensions = extensions

	return r
}

var defaultHeaders = dsl.MapMatcher{"content-type": dsl.String("application/json")}

func (r *Query) setHeader(headerName string, value dsl.Matcher) *Query {
	if r.Headers == nil {
		r.Headers = defaultHeaders
	}

	r.Headers[headerName] = value

	return r
}

// Construct a Pact HTTP request for a GraphQL interaction
func Interaction(request Query) *dsl.Request {
	if request.Headers == nil {
		request.Headers = defaultHeaders
	}

	return &dsl.Request{
		Method: request.Method,
		Path:   request.Path,
		Query:  request.QueryString,
		Body: graphQLQueryBody{
			Operation: request.Operation,
			Query:     dsl.Regex(request.Query, escapeGraphQlQuery(request.Query)),
			Variables: request.Variables,
		},
		Headers: request.Headers,
	}

}

type graphQLQueryBody struct {
	Operation string      `json:"operationName,omitempty"`
	Query     dsl.Matcher `json:"query"`
	Variables Variables   `json:"variables,omitempty"`
}

func escapeSpace(s string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllString(s, `\s*`)
}

func escapeRegexChars(s string) string {
	r := regexp.MustCompile(`(?m)[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]`)

	f := func(s string) string {
		return fmt.Sprintf(`\%s`, s)
	}
	return r.ReplaceAllStringFunc(s, f)
}

func escapeGraphQlQuery(s string) string {
	return escapeSpace(escapeRegexChars(s))
}
