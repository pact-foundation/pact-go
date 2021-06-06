package models

// SpecificationVersion is used to determine the current specification version
type SpecificationVersion string

const (
	// V2 signals the use of version 2 of the pact spec
	V2 SpecificationVersion = "2.0.0"

	// V3 signals the use of version 3 of the pact spec
	V3 = "3.0.0"
)
