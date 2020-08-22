package v3

import (
	"bytes"
	"encoding/json"
)

// Format a JSON document to make comparison easier.
func formatJSONString(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}

// Format a JSON document for creating Pact files.
func formatJSONObject(object interface{}) string {
	out, _ := json.Marshal(object)
	return formatJSONString(string(out))
}
