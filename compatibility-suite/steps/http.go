// Package steps holds the compatibility suite's step definitions and the
// helpers they share.
package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// sendRequest makes req against baseURL, returning the response with its
// body already read.
func sendRequest(baseURL string, req model.Request) (*http.Response, []byte, error) {
	target := baseURL + req.Path
	if len(req.Query) > 0 {
		target += "?" + req.Query.Encode()
	}

	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body.Contents)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), req.Method, target, body)
	if err != nil {
		return nil, nil, fmt.Errorf("building %s %s: %w", req.Method, target, err)
	}
	for name, values := range req.Headers {
		for _, v := range values {
			httpReq.Header.Add(name, v)
		}
	}
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", req.Body.ContentType)
	}

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("%s %s: %w", req.Method, target, err)
	}
	defer func() { _ = res.Body.Close() }()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading the response to %s %s: %w", req.Method, target, err)
	}

	return res, data, nil
}

// jsonEqual reports whether two JSON documents are equal, ignoring layout:
// pact-go serves and writes JSON bodies compacted.
func jsonEqual(got, want []byte) error {
	var g, w any

	err := json.Unmarshal(got, &g)
	if err != nil {
		return fmt.Errorf("want JSON, got %q: %w", got, err)
	}
	err = json.Unmarshal(want, &w)
	if err != nil {
		return fmt.Errorf("the expected document is not JSON: %w", err)
	}
	if !reflect.DeepEqual(g, w) {
		return fmt.Errorf("want JSON %s, got %s", want, got)
	}

	return nil
}
