package utils

// Simplistic map copy
func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))

	for k, v := range src {
		dst[k] = v
	}
	return dst
}
