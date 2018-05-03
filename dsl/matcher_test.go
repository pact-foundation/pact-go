package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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
func formatJSON(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
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
	str := "str"
	type args struct {
		src interface{}
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{
			name: "recursive case - ptr",
			args: args{
				src: &str,
			},
			want: Like(`"string"`),
		},
		{
			name: "recursive case - slice",
			args: args{
				src: []string{},
			},
			want: EachLike(Like(`"string"`), 1),
		},
		{
			name: "recursive case - array",
			args: args{
				src: [1]string{},
			},
			want: EachLike(Like(`"string"`), 1),
		},
		{
			name: "recursive case - struct",
			args: args{
				src: wordDTO{},
			},
			want: fmt.Sprintf("{\"word\": %s,\"length\": %s}", Like(`"string"`), Like(1)),
		},
		{
			name: "recursive case - struct with custom string tag",
			args: args{
				src: dateDTO{},
			},
			want: fmt.Sprintf("{\"date\": %s}", Term("2000-01-01", `^\\d{4}-\\d{2}-\\d{2}$`)),
		},
		{
			name: "recursive case - struct with custom slice tag",
			args: args{
				src: wordsDTO{},
			},
			want: fmt.Sprintf("{\"words\": %s}", EachLike(Like(`"string"`), 2)),
		},
		{
			name: "base case - string",
			args: args{
				src: "string",
			},
			want: Like(`"string"`),
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
			want: Like(1),
		},
		{
			name: "base case - float64",
			args: args{
				src: float64(1),
			},
			want: Like(1),
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
			var got string
			var didPanic bool
			defer func() {
				if rec := recover(); rec != nil {
					didPanic = true
				}
				if tt.wantPanic != didPanic {
					t.Errorf("Match() didPanic = %v, want %v", didPanic, tt.wantPanic)
				} else if !didPanic && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Match() = %v, want %v", got, tt.want)
				}
			}()
			got = Match(tt.args.src)
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
					regEx:   `\\d{2}`,
				},
			},
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
			name: "invalid string tag - no regex",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123",
			},
			wantPanic: true,
		},
		{
			name: "invalid string tag - regex typo",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=aBcD123,regx=[A-Za-z0-9]",
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
