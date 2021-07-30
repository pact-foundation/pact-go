package dsl

import "testing"

func TestRequest(t *testing.T) {
	req := Request{
		Method: "GET",
	}
	if req.Method != "GET" {
		t.Fatalf("Expected method to be 'GET' but got '%s'", req.Method)
	}
}

func TestRequest_Body(t *testing.T) {

}
