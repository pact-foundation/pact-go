package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestMatcher_TermString(t *testing.T) {
	expected := formatJSON(MatcherString(`
		{
			"json_class": "Pact::Term",
			"data": {
			  "generate": "myawesomeword",
			  "matcher": {
			    "json_class": "Regexp",
			    "o": 0,
			    "s": "\\w+"
			  }
			}
		}`))

	match := formatJSON(Term("myawesomeword", `\\w+`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeBasicString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": "myspecialvalue"
		}`)

	match := formatJSON(Like(`"myspecialvalue"`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeAsObject(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": {"baz":"bat"}
		}`)

	match := formatJSON(Like(`{"baz":"bat"}`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeNumber(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": 37
		}`)

	match := formatJSON(Like(37))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeNumberAsString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": 37
		}`)

	match := formatJSON(Like("37"))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeNumber(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": 37,
		  "min": 1
		}`)

	match := formatJSON(EachLike(37, 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeNumberAsString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": 37,
		  "min": 1
		}`)

	match := formatJSON(EachLike("37", 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": "someword",
		  "min": 7
		}`)

	match := formatJSON(EachLike(`"someword"`, 7))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeObject(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": {"somekey":"someval"},
		  "min": 3
		}`)

	match := formatJSON(EachLike(`{"somekey":"someval"}`, 3))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeArray(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": [1,2,3],
		  "min": 1
		}`)

	match := formatJSON(EachLike(`[1,2,3]`, 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_NestLikeInEachLike(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": {
		    "id": {
		      "json_class": "Pact::SomethingLike",
		      "contents": 10
		    }
		  },
		  "min": 1
		}`)

	match := formatJSON(EachLike(fmt.Sprintf(`{ "id": %s }`, Like(10)), 1))

	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_NestTermInEachLike(t *testing.T) {
	expected := formatJSON(`
		{
	    "json_class": "Pact::ArrayLike",
	    "contents": {
	      "colour": {
	        "json_class": "Pact::Term",
	        "data": {
	          "generate": "red",
	          "matcher": {
	            "json_class": "Regexp",
	            "o": 0,
	            "s": "red|green"
	          }
	        }
	      }
	    },
	    "min": 1
	  }`)

	match := formatJSON(
		EachLike(
			fmt.Sprintf(`{ "colour": %s }`,
				Term("red", `red|green`)),
			1))

	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_NestedEachLike(t *testing.T) {
	expected := formatJSON(`
		{
	    "json_class": "Pact::ArrayLike",
	    "contents": {
	      "json_class": "Pact::ArrayLike",
	      "contents": "blue",
	      "min": 1
	    },
	    "min": 1
	  }`)

	match := formatJSON(
		EachLike(
			EachLike(`"blue"`, 1),
			1))

	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_NestAllTheThings(t *testing.T) {
	expected := formatJSON(`{
					"json_class": "Pact::ArrayLike",
					"contents": {
						"json_class": "Pact::ArrayLike",
						"contents": {
							"size": {
								"json_class": "Pact::SomethingLike",
								"contents": 10
							},
							"colour": {
								"json_class": "Pact::Term",
								"data": {
									"generate": "red",
									"matcher": {
										"json_class": "Regexp",
										"o": 0,
										"s": "red|green|blue"
									}
								}
							},
							"tag": {
								"json_class": "Pact::ArrayLike",
								"contents": [
									{
										"json_class": "Pact::SomethingLike",
										"contents": "jumper"
									},
									{
										"json_class": "Pact::SomethingLike",
										"contents": "shirt"
									}
								],
								"min": 2
							}
						},
						"min": 1
					},
					"min": 1
				}`)

	jumper := Like(`"jumper"`)
	shirt := Like(`"shirt"`)
	tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
	size := Like(10)
	colour := Term("red", "red|green|blue")

	match := formatJSON(
		EachLike(
			EachLike(
				fmt.Sprintf(
					`{
						"size": %s,
						"colour": %s,
						"tag": %s
					}`, size, colour, tag),
				1),
			1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

// Format a JSON document to make comparison easier.
func formatJSON(object interface{}) interface{} {
	var out bytes.Buffer
	switch content := object.(type) {
	case string:
		json.Indent(&out, []byte(content), "", "\t")
	case MatcherString:
		json.Indent(&out, []byte(content), "", "\t")
	}

	return string(out.Bytes())
}

func ExampleLike_string() {
	match := Like(`"myspecialvalue"`)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::SomethingLike",
	//	"contents": "myspecialvalue"
	//}
}

func ExampleLike_object() {
	match := Like(`{"baz":"bat"}`)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::SomethingLike",
	//	"contents": {
	//		"baz": "bat"
	//	}
	//}
}

func ExampleLike_number() {
	match := Like(37)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::SomethingLike",
	//	"contents": 37
	//}
}

func ExampleTerm() {
	match := Term("myawesomeword", `\\w+`)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::Term",
	//	"data": {
	//		"generate": "myawesomeword",
	//		"matcher": {
	//			"json_class": "Regexp",
	//			"o": 0,
	//			"s": "\\w+"
	//		}
	//	}
	//}
}

func ExampleEachLike() {
	match := EachLike(`[1,2,3]`, 1)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::ArrayLike",
	//	"contents": [
	//		1,
	//		2,
	//		3
	//	],
	//	"min": 1
	//}
}

func ExampleEachLike_nested() {
	jumper := Like(`"jumper"`)
	shirt := Like(`"shirt"`)
	tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
	size := Like(10)
	colour := Term("red", "red|green|blue")

	match := EachLike(
		EachLike(
			fmt.Sprintf(
				`{
							"size": %s,
							"colour": %s,
							"tag": %s
						}`, size, colour, tag),
			1),
		1)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::ArrayLike",
	//	"contents": {
	//		"json_class": "Pact::ArrayLike",
	//		"contents": {
	//			"size": {
	//				"json_class": "Pact::SomethingLike",
	//				"contents": 10
	//			},
	//			"colour": {
	//				"json_class": "Pact::Term",
	//				"data": {
	//					"generate": "red",
	//					"matcher": {
	//						"json_class": "Regexp",
	//						"o": 0,
	//						"s": "red|green|blue"
	//					}
	//				}
	//			},
	//			"tag": {
	//				"json_class": "Pact::ArrayLike",
	//				"contents": [
	//					{
	//						"json_class": "Pact::SomethingLike",
	//						"contents": "jumper"
	//					},
	//					{
	//						"json_class": "Pact::SomethingLike",
	//						"contents": "shirt"
	//					}
	//				],
	//				"min": 2
	//			}
	//		},
	//		"min": 1
	//	},
	//	"min": 1
	//}
}
