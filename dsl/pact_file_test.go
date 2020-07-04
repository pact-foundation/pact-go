package dsl

import (
	"reflect"
	"testing"
)

func TestPactFile_term(t *testing.T) {
	matcher := map[string]interface{}{
		"id": Like(127),
	}

	expectedBody := formatJSON(`{
		"id": 127
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.id": map[string]interface{}{
			"match": "type",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestPactFile_ArrayMinLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMinLike(27, 3),
	}

	expectedBody := formatJSON(`{
		"users": [
			27,
			27,
			27
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestPactFile_ArrayMinLikeWithNested(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMinLike(map[string]interface{}{
			"user": Regex("someusername", "\\s+")}, 3)}

	expectedBody := formatJSON(`{
		"users": [
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			}
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
		"$.body.users[*].user": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}
