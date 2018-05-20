package types

// ReificationResponse contains the ouput of the reification request
type ReificationResponse struct {
	// Interface wrapped object
	Response interface{}

	// Raw response from reification
	ResponseRaw []byte
}
