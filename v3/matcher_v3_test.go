package v3

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestMatchV3(t *testing.T) {
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
		want      MatcherV2
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
			want: ArrayMinMaxLike(Like("string"), 1, 0),
		},
		{
			name: "recursive case - array",
			args: args{
				src: [1]string{},
			},
			want: ArrayMinMaxLike(Like("string"), 1, 0),
		},
		{
			name: "recursive case - struct",
			args: args{
				src: wordDTO{},
			},
			want: StructMatcher{
				"word":   Like("string"),
				"length": Integer(1),
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
				"words": ArrayMinMaxLike(Like("string"), 2, 0),
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
				"integer": Integer(42),
				"float":   Decimal(float32(6.66)),
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
			want: Integer(1),
		},
		{
			name: "base case - int8",
			args: args{
				src: int8(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - int16",
			args: args{
				src: int16(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - int32",
			args: args{
				src: int32(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - int64",
			args: args{
				src: int64(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - uint",
			args: args{
				src: uint(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - uint8",
			args: args{
				src: uint8(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - uint16",
			args: args{
				src: uint16(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - uint32",
			args: args{
				src: uint32(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - uint64",
			args: args{
				src: uint64(1),
			},
			want: Integer(1),
		},
		{
			name: "base case - float32",
			args: args{
				src: float32(1),
			},
			want: Decimal(1.1),
		},
		{
			name: "base case - float64",
			args: args{
				src: float64(1),
			},
			want: Decimal(1.1),
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
			var got MatcherV3
			var didPanic bool
			defer func() {
				if rec := recover(); rec != nil {
					fmt.Println(rec)
					didPanic = true
				}
				if tt.wantPanic != didPanic {
					t.Errorf("Match() - '%s': didPanic = %v, want %v", tt.name, didPanic, tt.wantPanic)
				} else if !didPanic && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Match() - '%s': got = %v, want %v", tt.name, got, tt.want)
				}
			}()

			got = MatchV3(tt.args.src)
			log.Println("Got matcher: ", got)
		})
	}
}

func Test_pluckParamsV3(t *testing.T) {
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
			name: "expected use - slice tag with min only",
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
			name: "expected use - slice tag with min and max",
			args: args{
				srcType: reflect.TypeOf([]string{}),
				pactTag: "min=2,max=3",
			},
			want: params{
				slice: sliceParams{
					min: 2,
					max: 3,
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
			name: "expected use - string tag example with no regex",
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
			name: "expected use - string tag with generator and format",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "2020-01-01'T'08:00:45",
				},
				generator: stringGenerator{
					contents:  "2020-01-01'T'08:00:45",
					format:    "yyyy-MM-dd'T'HH:mm:ss",
					generator: dateTimeGenerator,
				},
			},
			wantPanic: false,
		},
		{
			name: "expected use - string tag with generator, format and regex",
			args: args{
				srcType: reflect.TypeOf(""),
				pactTag: "example=2020-01-01'T'08:00:45,regex=[0-9-]+,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime",
			},
			want: params{
				slice: sliceParams{
					min: getDefaults().slice.min,
				},
				str: stringParams{
					example: "2020-01-01'T'08:00:45",
					regEx:   "[0-9-]+",
				},
				generator: stringGenerator{
					contents:  "2020-01-01'T'08:00:45",
					format:    "yyyy-MM-dd'T'HH:mm:ss",
					generator: dateTimeGenerator,
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
					t.Errorf("pluckParams() didPanic = %+v, want %+v", didPanic, tt.wantPanic)
				} else if !didPanic && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("pluckParams() got: \n%+v\n, want: \n%+v\n", got, tt.want)
				}
			}()
			got = pluckParamsV3(tt.args.srcType, tt.args.pactTag)
		})
	}
}
