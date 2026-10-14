package v2

import (
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestParseVerifierResult(t *testing.T) {
	data := []byte(`{"result":false,"errors":[
	  {"interaction":"c_1 - Interaction 0","mismatch":{"type":"mismatches","mismatches":[
	    {"type":"StatusMismatch","expected":200,"actual":400,"mismatch":"expected 200 but was 400"},
	    {"type":"BodyMismatch","path":"$"}]}},
	  {"interaction":"c_2 - Interaction 1","mismatch":{"type":"error","message":"One or more of the setup state change handlers has failed"}}]}`)

	got, err := parseVerifierResult(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []model.VerificationError{
		{Interaction: "c_1 - Interaction 0", MismatchTypes: []string{"StatusMismatch", "BodyMismatch"}},
		{Interaction: "c_2 - Interaction 1", Message: "One or more of the setup state change handlers has failed"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %#v\nwant %#v", got, want)
	}
}

func TestParseVerifierResultWithNoErrors(t *testing.T) {
	got, err := parseVerifierResult([]byte(`{"result":true,"errors":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want no errors, got %+v", got)
	}
}

func TestParseVerifierResultRejectsMalformedJSON(t *testing.T) {
	_, err := parseVerifierResult([]byte(`{`))
	if err == nil {
		t.Fatal("want an error for malformed JSON")
	}
}
