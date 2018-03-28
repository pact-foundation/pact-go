package dsl

// Message is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// Message is the main implementation of the Pact Message interface.
type Message struct {
	// Message Body
	Content interface{} `json:"content,omitempty"`

	// Provider state to be written into the Pact file
	State string `json:"providerState,omitempty"`

	// Message metadata
	Metadata MapMatcher `json:"metadata,omitempty"`

	// Description to be written into the Pact file
	Description string `json:"description"`

	Args []string `json:"-"`
}

// Given specifies a provider state. Optional.
func (p *Message) Given(state string) *Message {
	p.State = state
	return p
}

// ExpectsToReceive specifies the content it is expecting to be
// given from the Provider. The function must be able to handle this
// message for the interaction to succeed.
func (p *Message) ExpectsToReceive(description string) *Message {
	p.Description = description
	return p
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
func (p *Message) WithMetadata(metadata MapMatcher) *Message {
	p.Metadata = metadata
	return p
}

// WithContent specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (p *Message) WithContent(content interface{}) *Message {
	p.Content = content

	return p
}
