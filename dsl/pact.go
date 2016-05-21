package dsl

import "fmt"

type Pact struct {
}

func (p *Pact) Given(state string) *Pact {
	fmt.Println("Pact()")
	return p
}

func (p *Pact) UponReceiving(test string) *Pact {
	fmt.Println("UponReceiving()")
	return p
}

func (p *Pact) WithRequest(request Request) *Pact {
	fmt.Println("WithRequest()")
	return p
}

func (p *Pact) WillRespondWith(response Response) *Pact {
	fmt.Println("RespondWith()")
	return p
}

func (p *Pact) Verify() *Pact {
	fmt.Println("Verify()")
	return p
}
