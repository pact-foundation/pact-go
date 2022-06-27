package graphql

// GraphQLRseponse models the GraphQL Response format.
// See also http://spec.graphql.org/October2021/#sec-Response-Format
type Response struct {
	Data       interface{}            `json:"data,omitempty"`
	Errors     []interface{}          `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}
