package types

// ReificationResponse contains the output of the reification request
type ReificationResponse struct {
	// Interface wrapped object
	Response interface{}

	// Raw response from reification
	ResponseRaw []byte
}
