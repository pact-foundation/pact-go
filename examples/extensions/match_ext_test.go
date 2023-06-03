package extensions

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pkg/errors"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	Like     = matchers.Like
	EachLike = matchers.EachLike
	Term     = matchers.Term
)

type (
	StructMatcher = matchers.StructMatcher
	Matcher       = matchers.Matcher
)

func TestMatch(t *testing.T) {
	// mixedDTO in order to reuse protoc-gen-go where structs are compatible with protobuf and json
	type mixedDTO struct {
		// has tag and should be in output
		OnlyJsonTag string `json:"onlyJsonTag"`
		// no tag, skip
		NoTagString string
		// no tag, skip - this covers case of proto compatible structs that contain func fields
		NoTagFunc           func()
		BothUseJsonTag      int32 `protobuf:"varint,1,opt,name=both_use_json_tag,json=bothNameFromProtobufTag,proto3" json:"bothNameFromJsonTag,omitempty"`
		ProtoWithoutJsonTag *struct {
			OnlyJsonTag string `json:"onlyJsonTagNested"`
			// no tag, skip
			NoTag func()
		} `protobuf:"bytes,7,opt,name=proto_without_json_tag,json=onlyProtobufTag,proto3,oneof"`
	}
	type args struct {
		src interface{}
	}
	type matchTest struct {
		name      string
		args      args
		want      Matcher
		wantPanic bool
	}
	defaultTests := []matchTest{
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
			name: "base case - uint",
			args: args{
				src: uint(1),
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
			name: "error - unhandled type",
			args: args{
				src: make(map[string]string),
			},
			wantPanic: true,
		},
	}

	matchV2ProtoTests := append(defaultTests, matchTest{
		name: "structs mixed for compatibility with proto3 and json types",
		args: args{
			src: mixedDTO{},
		},
		want: StructMatcher{
			"onlyJsonTag":         Like("string"),
			"bothNameFromJsonTag": Like(1),
			"onlyProtobufTag":     StructMatcher{"onlyJsonTagNested": Like("string")},
		},
	})
	customMatcher := matchers.NewCustomMatchStructV2(matchers.CustomMatchStructV2Args{StrategyFunc: protoJsonFieldStrategyFunc})
	for _, tt := range matchV2ProtoTests {
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

			got = customMatcher.MatchV2(tt.args.src)
			log.Println("Got matcher: ", got)
		})
	}
}

// ProtoJsonFieldStrategyFunc provides an example to extend the matchers.CustomMatchStructV2Args.Match
// extension allows for parsing of custom structs tags, and this example demonstrates a
// potential solution relating to protojson
var protoJsonFieldStrategyFunc = func(field reflect.StructField) matchers.FieldMatchArgs {
	if fieldName, enum := fieldNameByTagStrategy(field); enum != "" {
		var pactTag string
		if _, ok := field.Tag.Lookup("pact"); ok {
			pactTag = field.Tag.Get("pact")
		} else {
			pactTag = generateDefaultTagForEnum(enum)
		}
		return matchers.FieldMatchArgs{Name: fieldName, MatchType: reflect.TypeOf("string"), PactTag: pactTag}
	} else if fieldName != "" {
		return matchers.FieldMatchArgs{Name: fieldName, MatchType: field.Type, PactTag: field.Tag.Get("pact")}
	} else {
		return matchers.DefaultFieldStrategyFunc(field)
	}
}

func fieldNameByTagStrategy(field reflect.StructField) (fieldName string, enum string) {
	var v string
	var ok bool
	if v, ok = field.Tag.Lookup("protobuf"); ok {
		arr := strings.Split(v, ",")
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], "json=") {
				fieldName = strings.Split(arr[i], "=")[1]
			}
			if strings.HasPrefix(arr[i], "enum=") {
				enum = strings.Split(arr[i], "=")[1]
			}
		}
	}

	if v, ok = field.Tag.Lookup("json"); ok {
		fieldName = strings.Split(v, ",")[0]
	}
	return fieldName, enum
}

func generateDefaultTagForEnum(enum string) string {
	var enumType protoreflect.EnumType
	var err error
	var example, regex string

	if enumType, err = protoregistry.GlobalTypes.FindEnumByName(protoreflect.FullName(enum)); err != nil {
		panic(errors.Wrapf(err, "could not find enum %s", enum))
	}

	values := enumType.Descriptor().Values()
	enumNames := make([]string, 0)
	for i := 0; i < values.Len(); i++ {
		enumNames = append(enumNames, fmt.Sprintf("%s", values.Get(i).Name()))
	}
	if len(enumNames) > 0 {
		example = enumNames[0]
	}
	regex = strings.Join(enumNames, "|")
	return fmt.Sprintf("example=%s,regex=^(%s)$", example, regex)
}
