package consumer

import (
	"errors"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestInteraction(t *testing.T) {
	t.Run("validateMatchers for V2 Specification", func(t *testing.T) {
		testCases := []struct {
			description string
			test        interface{}
			want        error
		}{
			{
				description: "string body should return nil",
				test:        "I'm a string",
				want:        nil,
			},
			{
				description: "boolean body should return nil",
				test:        true,
				want:        nil,
			},
			{
				description: "numeric body should return nil",
				test:        27,
				want:        nil,
			},
			{
				description: "v3 matches should error",
				test: map[string]interface{}{
					"dateTime":    matchers.Regex("2020-01-01", "[0-9\\-]+"),
					"name":        matchers.S("Billy"),
					"superstring": matchers.Includes("foo"),
					"nested": map[string]matchers.Matcher{
						"val": matchers.Includes("val"),
					},
				},
				want: errors.New("test error"),
			},
			{
				description: "v2 matches should no terror",
				test: map[string]interface{}{
					"dateTime": matchers.Regex("2020-01-01", "[0-9\\-]+"),
					"name":     matchers.S("Billy"),
					"nested": map[string]matchers.Matcher{
						"val": matchers.Regex("val", ".*"),
					},
				},
				want: nil,
			},
		}

		for _, test := range testCases {
			if test.want != nil {
				assert.Error(t, validateMatchers(models.V2, test.test), test.description)
			} else {
				assert.NoError(t, validateMatchers(models.V2, test.test), test.description)
			}
		}
	})
}

func TestHasMatcherGreaterThanSpec(t *testing.T) {
	testCases := []struct {
		description string
		obj         map[string]interface{}
		want        []string
	}{
		{
			description: "matcher within the spec version is not reported",
			obj: map[string]interface{}{
				"pact:specification": "2.0.0",
				"pact:matcher:type":  "regex",
			},
			want: []string{},
		},
		{
			description: "matcher beyond the spec version is reported by type",
			obj: map[string]interface{}{
				"pact:specification": "3.0.0",
				"pact:matcher:type":  "include",
			},
			want: []string{"include"},
		},
		{
			description: "matcher beyond the spec version with no type is still reported",
			obj: map[string]interface{}{
				"pact:specification": "3.0.0",
			},
			want: []string{"<unknown matcher type>"},
		},
		{
			description: "matcher beyond the spec version with a non-string type is still reported",
			obj: map[string]interface{}{
				"pact:specification": "3.0.0",
				"pact:matcher:type":  27,
			},
			want: []string{"<unknown matcher type>"},
		},
		{
			description: "a non-string specification version is ignored",
			obj: map[string]interface{}{
				"pact:specification": 3,
				"pact:matcher:type":  "include",
			},
			want: []string{},
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.want, hasMatcherGreaterThanSpec(models.V2, test.obj))
		})
	}
}
