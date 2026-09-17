package utils

import (
	"bytes"
	"encoding/json"
	"log"
)

// FormatJSONString formats a JSON document to make comparison easier.
func FormatJSONString(object string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(object), "", "\t")
	if err != nil {
		log.Println("[ERROR] failed to format string:", err)
		return ""
	}
	return out.String()
}

// FormatJSONObject formats a JSON document for creating Pact files.
func FormatJSONObject(object any) string {
	out, err := json.Marshal(object)
	if err != nil {
		log.Println("[ERROR] failed to encode string to json:", err)
		return ""
	}
	return FormatJSONString(string(out))
}

// IsJSONFormattedObject checks to see if someone has tried to submit a
// JSON string for an object, which is no longer supported.
func IsJSONFormattedObject(stringOrObject any) bool {
	switch content := stringOrObject.(type) {
	case []byte:
	case string:
		var obj any
		err := json.Unmarshal([]byte(content), &obj)
		if err != nil {
			return false
		}

		// Check if a map type
		if _, ok := obj.(map[string]any); ok {
			return true
		}
	}

	return false
}
