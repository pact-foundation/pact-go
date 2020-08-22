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

// Checks to see if someone has tried to submit a JSON string
// for an object, which is no longer supported
func isJSONFormattedObject(stringOrObject interface{}) bool {
	switch content := stringOrObject.(type) {
	case []byte:
	case string:
		var obj interface{}
		err := json.Unmarshal([]byte(content), &obj)

		if err != nil {
			return false
		}

		// Check if a map type
		if _, ok := obj.(map[string]interface{}); ok {
			return true
		}
	}

	return false
}
