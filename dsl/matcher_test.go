package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"
)

func TestMatcher_TermString(t *testing.T) {
	expected := formatJSON(`
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
		}`)

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

	match := formatJSON(Like("myspecialvalue"))
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
		  "contents": "42",
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
      "contents": "42",
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

	match := formatJSON(EachLike("someword", 7))
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

func TestMatcher_EachLikeObjectAsStringFail(t *testing.T) {
	expected := formatJSON(`
		{
		  "contents": {"somekey":"someval"},
		  "json_class": "Pact::ArrayLike",
		  "min": 3
		}`)

	match := formatJSON(EachLike(`{"somekey":"someval"}`, 3))
	if expected == match {
		t.Fatalf("Expected Term to NOT match. '%s' != '%s'", expected, match)
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

	match := formatJSON(EachLike(Matcher{
		"id": Like(10),
	}, 1))

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
			Matcher{
				"colour": Term("red", "red|green")},
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
			EachLike("blue", 1),
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

	match := formatJSON(
		EachLike(
			EachLike(
				Matcher{
					"colour": Term("red", "red|green|blue"),
					"size":   Like(10),
					"tag":    EachLike([]Matcher{Like("jumper"), Like("shirt")}, 2),
				},
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
	default:
		jsonString, err := json.Marshal(object)
		if err != nil {
			log.Println("[ERROR] unable to marshal json:", err)
		}
		json.Indent(&out, []byte(jsonString), "", "\t")
	}

	return string(out.Bytes())
}

// Instrument the Matcher type to be able to assert the
// values and regexs contained within!
func (m Matcher) getValue() interface{} {
	mString := objectToString(m)

	// try like
	likeValue := &like{}
	err := json.Unmarshal([]byte(mString), likeValue)
	if err == nil && likeValue.Contents != nil {
		return likeValue.Contents
	}

	// try term
	termValue := &term{}
	err = json.Unmarshal([]byte(mString), termValue)
	if err == nil && termValue != nil {
		return termValue.Data.Generate
	}

	return "no value found"
}

func TestMatcher_SugarMatchers(t *testing.T) {

	type matcherTestCase struct {
		matcher  Matcher
		testCase func(val interface{}) error
	}
	matchers := map[string]matcherTestCase{
		"HexValue": matcherTestCase{
			matcher: HexValue(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "3F" {
					err = fmt.Errorf("want '3F', got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Identifier": matcherTestCase{
			matcher: Identifier(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(float64) // JSON converts numbers to float64 in anonymous structs
				if !valid {
					err = fmt.Errorf("want int, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Integer": matcherTestCase{
			matcher: Integer(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(float64) // JSON converts numbers to float64 in anonymous structs
				if !valid {
					err = fmt.Errorf("want int, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"IPAddress": matcherTestCase{
			matcher: IPAddress(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "127.0.0.1" {
					err = fmt.Errorf("want '127.0.0.1', got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"IPv4Address": matcherTestCase{
			matcher: IPv4Address(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "127.0.0.1" {
					err = fmt.Errorf("want '127.0.0.1', got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"IPv6Address": matcherTestCase{
			matcher: IPv6Address(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "::ffff:192.0.2.128" {
					err = fmt.Errorf("want '::ffff:192.0.2.128', got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Decimal": matcherTestCase{
			matcher: Decimal(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(float64)
				if !valid {
					err = fmt.Errorf("want float64, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Timestamp": matcherTestCase{
			matcher: Timestamp(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Date": matcherTestCase{
			matcher: Date(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"Time": matcherTestCase{
			matcher: Time(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", reflect.TypeOf(v))
				}
				return
			},
		},
		"UUID": matcherTestCase{
			matcher: UUID(),
			testCase: func(v interface{}) (err error) {
				match, err := regexp.MatchString(uuid, v.(string))

				if !match {
					err = fmt.Errorf("want string, got '%v'. Err: %v", v, err)
				}
				return
			},
		},
	}
	var err error
	for k, v := range matchers {
		if err = v.testCase(v.matcher.getValue()); err != nil {
			t.Fatalf("error validating matcher '%s': %v", k, err)
		}
	}
}

func ExampleLike_string() {
	match := Like("myspecialvalue")
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
