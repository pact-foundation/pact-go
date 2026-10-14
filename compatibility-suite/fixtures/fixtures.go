// Package fixtures reads the upstream suite's fixture files and the body
// cells of its feature tables.
package fixtures

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// Loader reads fixtures from the upstream suite's fixtures directory.
type Loader struct {
	Dir string
}

// Read returns a fixture file's raw contents.
func (l Loader) Read(name string) ([]byte, error) {
	//nolint:gosec // G304: name comes from the vendored upstream feature files.
	data, err := os.ReadFile(filepath.Join(l.Dir, name))
	if err != nil {
		return nil, fmt.Errorf("reading fixture %s: %w", name, err)
	}

	return data, nil
}

// Body parses a body cell:
//   - "JSON: <document>" and "XML: <document>" give that document;
//   - "file: <name>" loads a fixture (see File);
//   - anything else is literal text, with "EMPTY" meaning an empty body.
//
// fallback is the content type for text whose type cannot be inferred. A
// step changing a body passes the existing body's content type, as the
// reference implementation does.
func (l Loader) Body(cell, fallback string) (*model.Body, error) {
	cell = strings.TrimSpace(cell)

	switch {
	case cell == "":
		return nil, errors.New("a body cell is empty")
	case strings.HasPrefix(cell, "JSON:"):
		return &model.Body{
			ContentType: "application/json",
			Contents:    []byte(strings.TrimSpace(strings.TrimPrefix(cell, "JSON:"))),
		}, nil
	case strings.HasPrefix(cell, "XML:"):
		return &model.Body{
			ContentType: "application/xml",
			Contents:    []byte(strings.TrimSpace(strings.TrimPrefix(cell, "XML:"))),
		}, nil
	case strings.HasPrefix(cell, "file:"):
		return l.File(strings.TrimSpace(strings.TrimPrefix(cell, "file:")), fallback)
	}

	text := cell
	if cell == "EMPTY" {
		text = ""
	}

	return &model.Body{ContentType: contentType(cell, fallback), Contents: []byte(text)}, nil
}

// File loads a fixture as a body. A "*-body.xml" fixture describes a body
// (its content type and contents); any other file is the body itself, typed
// by its file name.
func (l Loader) File(name, fallback string) (*model.Body, error) {
	data, err := l.Read(name)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(name, "-body.xml") {
		return parseBodyXML(name, data)
	}

	return &model.Body{ContentType: contentType(name, fallback), Contents: data}, nil
}

// contentType infers a content type from a file name or text the way the
// reference implementation does, falling back to fallback, then text/plain.
func contentType(name, fallback string) string {
	switch {
	case strings.HasSuffix(name, ".json"):
		return "application/json"
	case strings.HasSuffix(name, ".xml"):
		return "application/xml"
	case strings.HasSuffix(name, ".jpg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".pdf"):
		return "application/pdf"
	case fallback != "":
		return fallback
	default:
		return "text/plain"
	}
}

type bodyXML struct {
	ContentType string `xml:"contentType"`
	Contents    struct {
		EOL  string `xml:"eol,attr"`
		Text string `xml:",chardata"`
	} `xml:"contents"`
}

// parseBodyXML reads a body fixture: <body><contentType/><contents/></body>.
func parseBodyXML(name string, data []byte) (*model.Body, error) {
	var doc bodyXML

	err := xml.Unmarshal(data, &doc)
	if err != nil {
		return nil, fmt.Errorf("parsing body fixture %s: %w", name, err)
	}

	contents := strings.TrimSpace(doc.Contents.Text)
	// Normalise first, so a checkout with CRLF line endings is not doubled up.
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	if doc.Contents.EOL == "CRLF" {
		contents = strings.ReplaceAll(contents, "\n", "\r\n")
	}

	ct := strings.TrimSpace(doc.ContentType)
	if ct == "" {
		ct = "text/plain"
	}

	return &model.Body{ContentType: ct, Contents: []byte(contents)}, nil
}
