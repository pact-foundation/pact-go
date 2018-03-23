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
      "data": {
        "generate": "myawesomeword",
			  "matcher": {
          "json_class": "Regexp",
			    "o": 0,
			    "s": "\\w+"
			  }
			},
      "json_class": "Pact::Term"
		}`))

	match := formatJSON(Term("myawesomeword", `\w+`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeBasicString(t *testing.T) {
	expected := formatJSON(`
		{
      "contents": "myspecialvalue",
		  "json_class": "Pact::SomethingLike"
		}`)

	match := formatJSON(Like(`"myspecialvalue"`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeAsObject(t *testing.T) {
	expected := formatJSON(`
		{
      "contents": {"baz":"bat"},
		  "json_class": "Pact::SomethingLike"
		}`)

	match := formatJSON(Like(map[string]string{
		"baz": "bat",
	}))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeAsObjectString(t *testing.T) {
	expected := formatJSON(`
		{
      "contents": {"baz":"bat"},
		  "json_class": "Pact::SomethingLike"
		}`)

	match := formatJSON(Like(`{"baz":"bat"}`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeNumber(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": 42,
		  "json_class": "Pact::SomethingLike"
		}`)

	match := formatJSON(Like(42))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeNumberAsString(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": 42,
		  "json_class": "Pact::SomethingLike"
		}`)

	match := formatJSON(Like("42"))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeNumber(t *testing.T) {
	expected := formatJSON(`
		{
      "contents": 42,
		  "json_class": "Pact::ArrayLike",
		  "min": 1
		}`)

	match := formatJSON(EachLike(42, 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}
func TestMatcher_EachLikeNumberAsString(t *testing.T) {
	expected := formatJSON(`
		{
      "contents": 42,
		  "json_class": "Pact::ArrayLike",
		  "min": 1
		}`)

	match := formatJSON(EachLike("42", 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeString(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": "someword",
		  "json_class": "Pact::ArrayLike",
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
      "contents": {"somekey":"someval"},
		  "json_class": "Pact::ArrayLike",
		  "min": 3
		}`)

	match := formatJSON(EachLike(map[string]string{
		"somekey": "someval",
	}, 3))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeObjectAsString(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": {"somekey":"someval"},
		  "json_class": "Pact::ArrayLike",
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
      "contents": [1,2,3],
		  "json_class": "Pact::ArrayLike",
		  "min": 1
		}`)

	match := formatJSON(EachLike([]int{1, 2, 3}, 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeArrayString(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": [1,2,3],
		  "json_class": "Pact::ArrayLike",
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
      "contents": {
        "id": {
          "contents": 10,
		      "json_class": "Pact::SomethingLike"
		    }
		  },
      "json_class": "Pact::ArrayLike",
		  "min": 1
		}`)

	match := formatJSON(EachLike(map[string]interface{}{
		"id": Like(10),
		// "id": map[string]interface{}{
		// 	"contents":   10,
		// 	"json_class": "Pact::SomethingLike",
		// },
	}, 1))

	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}
func TestMatcher_NestLikeInEachLikeString(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": {
		    "id": {
		      "contents": 10,
		      "json_class": "Pact::SomethingLike"
		    }
      },
      "json_class": "Pact::ArrayLike",
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
	    "contents": {
	      "colour": {
	        "data": {
	          "generate": "red",
	          "matcher": {
	            "json_class": "Regexp",
	            "o": 0,
	            "s": "red|green"
	          }
          },
          "json_class": "Pact::Term"
	      }
      },
      "json_class": "Pact::ArrayLike",
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
      "contents": {
	      "contents": "blue",
        "json_class": "Pact::ArrayLike",
	      "min": 1
	    },
      "json_class": "Pact::ArrayLike",
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
					"contents": {
						"contents": {
							"colour": {
                "data": {
                  "generate": "red",
									"matcher": {
                    "json_class": "Regexp",
										"o": 0,
										"s": "red|green|blue"
									}
								},
                "json_class": "Pact::Term"
              },
							"size": {
                "contents": 10,
								"json_class": "Pact::SomethingLike"
							},
							"tag": {
                "contents": [
                  {
                    "contents": "jumper",
                    "json_class": "Pact::SomethingLike"
									},
									{
                    "contents": "shirt",
                    "json_class": "Pact::SomethingLike"
									}
                ],
                "json_class": "Pact::ArrayLike",
								"min": 2
							}
            },
            "json_class": "Pact::ArrayLike",
						"min": 1
          },
          "json_class": "Pact::ArrayLike",
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
	//	"contents": "myspecialvalue",
	//	"json_class": "Pact::SomethingLike"
	//}
}

func ExampleLike_object() {
	match := Like(map[string]string{"baz": "bat"})
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"contents": {
	//		"baz": "bat"
	//	},
	//	"json_class": "Pact::SomethingLike"
	//}
}
func ExampleLike_objectString() {
	match := Like(`{"baz":"bat"}`)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"contents": {
	//		"baz": "bat"
	//	},
	//	"json_class": "Pact::SomethingLike"
	//}
}

func ExampleLike_number() {
	match := Like(42)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"contents": 42,
	//	"json_class": "Pact::SomethingLike"
	//}
}

func ExampleTerm() {
	match := Term("myawesomeword", `\w+`)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"data": {
	//		"generate": "myawesomeword",
	//		"matcher": {
	//			"json_class": "Regexp",
	//			"o": 0,
	//			"s": "\\w+"
	//		}
	//	},
	//	"json_class": "Pact::Term"
	//}
}

func ExampleEachLike() {
	match := EachLike([]int{1, 2, 3}, 1)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"contents": [
	//		1,
	//		2,
	//		3
	//	],
	//	"json_class": "Pact::ArrayLike",
	//	"min": 1
	//}
}

func E_xampleEachLike_nested() {
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
	// {
	//   "contents": {
	//     "contents": {
	//       "colour": {
	//         "data": {
	//           "generate": "%s",
	//           "matcher": {
	//             "json_class": "Regexp",
	//             "o": 0,
	//             "s": "red"
	//           }
	//         },
	//         "json_class": "Pact::Term"
	//       },
	//       "size": {
	//         "contents": 10,
	//         "json_class": "Pact::SomethingLike"
	//       },
	//       "tag": {
	//         "contents": [
	//           {
	//             "contents": "jumper",
	//             "json_class": "Pact::SomethingLike"
	//           },
	//           {
	//             "contents": "shirt",
	//             "json_class": "Pact::SomethingLike"
	//           }
	//         ],
	//         "json_class": "Pact::ArrayLike",
	//         "min": 2
	//       }
	//     },
	//     "json_class": "Pact::ArrayLike",
	//     "min": 1
	//   },
	//   "json_class": "Pact::ArrayLike",
	//   "min": 1
	// }
}
