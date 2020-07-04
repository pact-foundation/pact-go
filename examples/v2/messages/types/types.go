package types

type AccessLevel struct {
	Role string `json:"role,omitempty"`
}

type User struct {
	ID     int           `json:"id,omitempty"`
	Name   string        `json:"name,omitempty"`
	Access []AccessLevel `json:"access,omitempty"`
}

type Error struct {
	Message string `json:"message" pact:"example=user not found"`
}

type Order struct {
	ID   int    `json:"id" pact:"example=42"`
	Item string `json:"item" pact:"example=apple,regex=(apple|orange)"`
}
