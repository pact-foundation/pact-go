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
      "json_class": "Pact::Term",
      "data": {
        "generate": "myawesomeword",
			  "matcher": {
          "json_class": "Regexp",
			    "o": 0,
			    "s": "\\w+"
			  }
			}
		}`)

	match := formatJSON(Term("myawesomeword", `\w+`))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_TermGetValue(t *testing.T) {
	expected := "myawesomeword"
	match := Term("myawesomeword", `\w+`).GetValue()

	if expected != match {
		t.Fatalf("Expected Term Value to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeBasicString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": "myspecialvalue"
		}`)

	match := formatJSON(Like("myspecialvalue"))
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
		  "json_class": "Pact::SomethingLike",
		  "contents": 42
		}`)

	match := formatJSON(Like(42))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeNumberAsString(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::SomethingLike",
		  "contents": "42"
		}`)

	match := formatJSON(Like("42"))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_LikeGetValue(t *testing.T) {
	expected := "myspecialvalue"
	match := Like("myspecialvalue").GetValue()

	if expected != match {
		t.Fatalf("Expected Like Value to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeNumber(t *testing.T) {
	expected := formatJSON(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": 42,
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
		  "json_class": "Pact::ArrayLike",
		  "contents": "42",
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
		  "json_class": "Pact::ArrayLike",
		  "contents": "someword",
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
		  "json_class": "Pact::ArrayLike",
		  "contents": {"somekey":"someval"},
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
		  "json_class": "Pact::ArrayLike",
		  "contents": {"somekey":"someval"},
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
		  "json_class": "Pact::ArrayLike",
		  "contents": [1,2,3],
		  "min": 1
		}`)

	match := formatJSON(EachLike([]int{1, 2, 3}, 1))
	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_EachLikeGetValue(t *testing.T) {
	expected := "42"
	match := EachLike("42", 1).GetValue()

	if expected != match {
		t.Fatalf("Expected EachLike Value to match. '%s' != '%s'", expected, match)
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

	match := formatJSON(EachLike(StructMatcher{
		"id": Like(10),
	}, 1))

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
			StructMatcher{
				"colour": Term("red", "red|green")},
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
			EachLike("blue", 1),
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
							"size": {
							  "json_class": "Pact::SomethingLike",
							  "contents": 10
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

	match := formatJSON(
		EachLike(
			EachLike(
				StructMatcher{
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

// Instrument the StructMatcher type to be able to assert the
// values and regexs contained within!
func getMatcherValue(m interface{}) interface{} {
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
		if err = v.testCase(getMatcherValue(v.matcher)); err != nil {
			t.Fatalf("error validating matcher '%s': %v", k, err)
		}
	}
}

func ExampleLike_string() {
	match := Like("myspecialvalue")
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::SomethingLike",
	//	"contents": "myspecialvalue"
	//}
}

func ExampleLike_object() {
	match := Like(map[string]string{"baz": "bat"})
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
	match := Like(42)
	fmt.Println(formatJSON(match))
	// Output:
	//{
	//	"json_class": "Pact::SomethingLike",
	//	"contents": 42
	//}
}

func ExampleTerm() {
	match := Term("myawesomeword", `\w+`)
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
	match := EachLike([]int{1, 2, 3}, 1)
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

func TestMatch(t *testing.T) {
	type wordDTO struct {
		Word   string `json:"word"`
		Length int    `json:"length"`
	}
	type dateDTO struct {
		Date string `json:"date" pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
	}
	type wordsDTO struct {
		Words []string `json:"words" pact:"min=2"`
	}
	type boolDTO struct {
		Boolean bool `json:"boolean" pact:"example=true"`
	}
	type numberDTO struct {
		Integer int     `json:"integer" pact:"example=42"`
		Float   float32 `json:"float" pact:"example=6.66"`
	}
	str := "str"
	type args struct {
		src interface{}
	}
	tests := []struct {
		name      string
		args      args
		want      Matcher
		wantPanic bool
	}{
		{
			name: "recursive case - ptr",
			args: args{
				src: &str,
			},
			want: Like("string"),
		},
		{
			name: "recursive case - slice",
			args: args{
				src: []string{},
			},
			want: EachLike(Like("string"), 1),
		},
		{
			name: "recursive case - array",
			args: args{
				src: [1]string{},
			},
			want: EachLike(Like("string"), 1),
		},
		{
			name: "recursive case - struct",
			args: args{
				src: wordDTO{},
			},
			want: StructMatcher{
				"word":   Like("string"),
				"length": Like(1),
			},
		},
		{
			name: "recursive case - struct with custom string tag",
			args: args{
				src: dateDTO{},
			},
			want: StructMatcher{
				"date": Term("2000-01-01", `^\d{4}-\d{2}-\d{2}$`),
			},
		},
		{
			name: "recursive case - struct with custom slice tag",
			args: args{
				src: wordsDTO{},
			},
			want: StructMatcher{
				"words": EachLike(Like("string"), 2),
			},
		},
		{
			name: "recursive case - struct with bool",
			args: args{
				src: boolDTO{},
			},
			want: StructMatcher{
				"boolean": Like(true),
			},
		},
		{
			name: "recursive case - struct with int and float",
			args: args{
				src: numberDTO{},
			},
			want: StructMatcher{
				"integer": Like(42),
				"float":   Like(float32(6.66)),
			},
		},
		{
			name: "base case - string",
			args: args{
				src: "string",
			},
			want: Like("string"),
		},
		{
			name: "base case - bool",
			args: args{
				src: true,
			},
			want: Like(true),
		},
		{
			name: "base case - int",
			args: args{
				src: 1,
			},
			want: Like(1),
		},
		{
			name: "base case - int8",
			args: args{
				src: int8(1),
			},
			want: Like(1),
		},
		{
			name: "base case - int16",
			args: args{
				src: int16(1),
			},
			want: Like(1),
		},
		{
			name: "base case - int32",
			args: args{
				src: int32(1),
			},
			want: Like(1),
		},
		{
			name: "base case - int64",
			args: args{
				src: int64(1),
			},
			want: Like(1),
		},
		{
			name: "base case - uint",
			args: args{
				src: uint(1),
			},
			want: Like(1),
		},
		{
			name: "base case - uint8",
			args: args{
				src: uint8(1),
			},
			want: Like(1),
		},
		{
			name: "base case - uint16",
			args: args{
				src: uint16(1),
			},
			want: Like(1),
		},
		{
			name: "base case - uint32",
			args: args{
				src: uint32(1),
			},
			want: Like(1),
		},
		{
			name: "base case - uint64",
			args: args{
				src: uint64(1),
			},
			want: Like(1),
		},
		{
			name: "base case - float32",
			args: args{
				src: float32(1),
			},
			want: Like(1.1),
		},
		{
			name: "base case - float64",
			args: args{
				src: float64(1),
			},
			want: Like(1.1),
		},
		{
			name: "error - unhandled type",
			args: args{
				src: make(map[string]string),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Matcher
			var didPanic bool
			defer func() {
				if rec := recover(); rec != nil {
					fmt.Println(rec)
					didPanic = true
				}
				if tt.wantPanic != didPanic {
					t.Errorf("Match() - '%s': didPanic = %v, want %v", tt.name, didPanic, tt.wantPanic)
				} else if !didPanic && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Match() - '%s':  = %v, want %v", tt.name, got, tt.want)
				}
			}()

			got = Match(tt.args.src)
			log.Println("Got matcher: ", got)
		})
	}
}

func Test_pluckParams(t *testing.T) {
	type args struct {
		srcType reflect.Type
		pactTag string
	}
	tests := []struct {
		name      string
		args      args
		want      params
		wantPanic bool
	}{
		{
			name: "expected use - slice tag",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "min=2",
			},
			want: params{
				slice: sliceParams{
					min: 2,
				},
				str: stringParams{
					example: getDefaults().str.example,
					regEx:   getDefaults().str.regEx,
				},
			},
		},
		{
			name: "empty slice tag",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "",
			},
			want: getDefaults(),
		},
		{
			name: "invalid slice tag - no min",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "min=",
			},
			wantPanic: true,
		},
		{
			name: "invalid slice tag - min typo capital letter",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "Min=2",
			},
			wantPanic: true,
		},
		{
			name: "invalid slice tag - min typo non-number",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "min=a",
			},
			wantPanic: true,
		},
		{
			name: "expected use - string tag",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123,regex=[A-Za-z0-9]",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "aBcD123",
					regEx:   "[A-Za-z0-9]",
				},
			},
		},
		{
			name: "expected use - string tag with backslash",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=33,regex=\\d{2}",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "33",
					regEx:   `\d{2}`,
				},
			},
		},
		{
			name: "expected use - string tag with complex string",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=Jean-Marie de La Beaujardière😀😍",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "Jean-Marie de La Beaujardière😀😍",
				},
			},
		},
		{
			name: "expected use - example with no regex",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "aBcD123",
				},
			},
			wantPanic: false,
		},
		{
			name: "empty string tag",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "",
			},
			want: getDefaults(),
		},
		{
			name: "invalid string tag - no example value",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=,regex=[A-Za-z0-9]",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - no example",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "regex=[A-Za-z0-9]",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - empty example",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - example typo",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "exmple=aBcD123,regex=[A-Za-z0-9]",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - no regex value",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123,regex=",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - space inserted",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123 regex=[A-Za-z0-9]",
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got params
			var didPanic bool
			defer func() {
				if rec := recover(); rec != nil {
					didPanic = true
				}
				if tt.wantPanic != didPanic {
					t.Errorf("pluckParams() didPanic = %v, want %v", didPanic, tt.wantPanic)
				} else if !didPanic && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("pluckParams() = %v, want %v", got, tt.want)
				}
			}()
			got = pluckParams(tt.args.srcType, tt.args.pactTag)
		})
	}
}
