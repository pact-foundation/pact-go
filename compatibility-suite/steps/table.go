package steps

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// responseColumn is the "response" column of an interactions table, which
// several steps read.
const responseColumn = "response"

// headersCell is the "headers" cell of an interactions or changes table.
const headersCell = "headers"

// rows returns a godog table's cell values.
func rows(table *godog.Table) [][]string {
	out := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		cells := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cells = append(cells, cell.Value)
		}
		out = append(out, cells)
	}

	return out
}

// changes reads a "with the following changes" table: a header row and one
// row of values, keyed by column name.
func changes(rows [][]string) (map[string]string, error) {
	const headerAndValues = 2
	if len(rows) != headerAndValues {
		return nil, fmt.Errorf("a changes table needs a header and one row of values, got %d rows", len(rows))
	}

	header, values := rows[0], rows[1]
	out := make(map[string]string, len(header))
	for i, name := range header {
		if i >= len(values) {
			break
		}
		out[strings.TrimSpace(name)] = strings.TrimSpace(values[i])
	}

	return out, nil
}

// interactionsFromTable builds the interactions of a "the following HTTP
// interactions have been defined" table, numbered from 1 in row order.
// Columns other than those read here (such as "No") are ignored, as is
// "response content": the reference implementation also takes a body's
// content type from its body cell.
func interactionsFromTable(rows [][]string, fx fixtures.Loader) ([]model.Interaction, error) {
	if len(rows) == 0 {
		return nil, errors.New("the interactions table is empty")
	}

	header := rows[0]
	out := make([]model.Interaction, 0, len(rows)-1)
	for i, row := range rows[1:] {
		cells := make(map[string]string, len(header))
		for c, name := range header {
			if c < len(row) {
				cells[strings.TrimSpace(name)] = strings.TrimSpace(row[c])
			}
		}

		in := model.Interaction{
			Number: i + 1,
			// The reference implementation numbers descriptions from 0.
			Description: fmt.Sprintf("Interaction %d", i),
			Request:     model.Request{Method: cells["method"], Path: cells["path"]},
			Response:    model.Response{Status: http.StatusOK},
		}

		err := applyRequestChanges(&in.Request, map[string]string{
			"query": cells["query"], "headers": cells[headersCell], "body": cells["body"],
		}, fx)
		if err != nil {
			return nil, fmt.Errorf("interaction %d: %w", in.Number, err)
		}

		err = applyResponseChanges(&in.Response, map[string]string{
			responseColumn: cells[responseColumn], "response headers": cells["response headers"], "response body": cells["response body"],
		}, fx, false)
		if err != nil {
			return nil, fmt.Errorf("interaction %d: %w", in.Number, err)
		}

		out = append(out, in)
	}

	return out, nil
}

// applyRequestChanges sets the request fields named by cells; an empty
// value clears a field. A text body keeps the content type of the body it
// replaces, as the reference implementation does.
func applyRequestChanges(req *model.Request, cells map[string]string, fx fixtures.Loader) error {
	for name, value := range cells {
		var err error

		switch name {
		case "method":
			req.Method = value
		case "path":
			req.Path = value
		case "query":
			req.Query, err = parseQueryCell(value)
		case headersCell:
			req.Headers, err = parseHeaders(value)
		case "body":
			req.Body, err = bodyCell(value, req.Body, fx)
		default:
			return fmt.Errorf("unsupported request change %q", name)
		}

		if err != nil {
			return fmt.Errorf("request %s: %w", name, err)
		}
	}

	return nil
}

// applyResponseChanges sets the response fields named by cells. With
// mergeHeaders, new header values are added to existing ones rather than
// replacing them, as the reference implementation's provider steps do.
func applyResponseChanges(res *model.Response, cells map[string]string, fx fixtures.Loader, mergeHeaders bool) error {
	for name, value := range cells {
		var err error

		switch name {
		case responseColumn:
			if value != "" {
				res.Status, err = strconv.Atoi(value)
			}
		case "response headers":
			err = setResponseHeaders(res, value, mergeHeaders)
		case "response body":
			res.Body, err = bodyCell(value, res.Body, fx)
		default:
			return fmt.Errorf("unsupported response change %q", name)
		}

		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func setResponseHeaders(res *model.Response, value string, merge bool) error {
	headers, err := parseHeaders(value)
	if err != nil {
		return err
	}

	if !merge || res.Headers == nil {
		res.Headers = headers

		return nil
	}

	for name, values := range headers {
		for _, v := range values {
			res.Headers.Add(name, v)
		}
	}

	return nil
}

// bodyCell parses a body cell, with an empty cell meaning no body.
func bodyCell(value string, current *model.Body, fx fixtures.Loader) (*model.Body, error) {
	if value == "" {
		return nil, nil //nolint:nilnil // an empty cell is a valid "no body"
	}

	fallback := ""
	if current != nil {
		fallback = current.ContentType
	}

	body, err := fx.Body(value, fallback)
	if err != nil {
		return nil, fmt.Errorf("parsing body: %w", err)
	}

	return body, nil
}

func parseQueryCell(value string) (url.Values, error) {
	if value == "" {
		return nil, nil
	}

	query, err := url.ParseQuery(value)
	if err != nil {
		return nil, fmt.Errorf("parsing query %q: %w", value, err)
	}

	return query, nil
}

// parseHeaders reads a headers cell such as 'X-TEST: Compatibility', with
// several headers separated by commas.
func parseHeaders(value string) (http.Header, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	headers := http.Header{}
	for part := range strings.SplitSeq(value, ",") {
		part = strings.Trim(strings.TrimSpace(part), "'")
		name, v, ok := strings.Cut(part, ":")
		if !ok {
			return nil, fmt.Errorf("header %q has no ':'", part)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(v))
	}

	return headers, nil
}

// parseNumbers reads a list of interaction numbers such as "1, 2".
func parseNumbers(list string) ([]int, error) {
	parts := strings.Split(list, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("interaction list %q: %w", list, err)
		}
		out = append(out, n)
	}

	return out, nil
}

// ordinal turns "first", "second" or "third" into an index.
func ordinal(word string) (int, error) {
	for i, known := range []string{"first", "second", "third"} {
		if word == known {
			return i, nil
		}
	}

	return 0, fmt.Errorf("unknown ordinal %q", word)
}
