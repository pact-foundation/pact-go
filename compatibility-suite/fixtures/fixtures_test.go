package fixtures_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
)

var fx = fixtures.Loader{Dir: "../pact-compatibility-suite/fixtures"}

func TestBodyFromJSONCell(t *testing.T) {
	got, err := fx.Body(`JSON: {"one": "a", "two": "c"}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/json" || string(got.Contents) != `{"one": "a", "two": "c"}` {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromXMLCell(t *testing.T) {
	got, err := fx.Body(`XML: <?xml version="1.0"?><values/>`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/xml" || string(got.Contents) != `<?xml version="1.0"?><values/>` {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromJSONFixture(t *testing.T) {
	got, err := fx.Body("file: basic.json", "")
	if err != nil {
		t.Fatal(err)
	}
	want, err := fx.Read("basic.json")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/json" || !bytes.Equal(got.Contents, want) {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromTextBodyFixture(t *testing.T) {
	got, err := fx.Body("file: text-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "text/plain" || string(got.Contents) != "Hello World!" {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromXMLBodyFixtureKeepsTheDocument(t *testing.T) {
	got, err := fx.Body("file: xml-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(got.Contents)
	if got.ContentType != "application/xml" || !strings.HasPrefix(contents, "<?xml") ||
		!strings.HasSuffix(contents, "</values>") {
		t.Fatalf("got %q %q", got.ContentType, contents)
	}
}

func TestBodyFromMultipartFixtureUsesCRLF(t *testing.T) {
	got, err := fx.Body("file: multipart-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(got.Contents)
	if !strings.HasPrefix(got.ContentType, "multipart/mixed; boundary=gc0p4Jq0M2Yt08jU534c0p") {
		t.Fatalf("content type %q", got.ContentType)
	}
	if !strings.HasPrefix(contents, "--gc0p4Jq0M2Yt08jU534c0p\r\n") {
		t.Fatalf("contents start %q", contents[:min(40, len(contents))])
	}
	if strings.Count(contents, "\n") != strings.Count(contents, "\r\n") {
		t.Fatalf("contents have bare LF line endings: %q", contents)
	}
}

func TestBodyFromBinaryFixture(t *testing.T) {
	got, err := fx.Body("file: rat.jpg", "")
	if err != nil {
		t.Fatal(err)
	}
	want, err := fx.Read("rat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "image/jpeg" || !bytes.Equal(got.Contents, want) {
		t.Fatalf("got %q with %d bytes", got.ContentType, len(got.Contents))
	}
}

func TestBodyFromTextUsesFallback(t *testing.T) {
	got, err := fx.Body("a=1&b=2&c=33&d=4", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/x-www-form-urlencoded" || string(got.Contents) != "a=1&b=2&c=33&d=4" {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}

	plain, err := fx.Body("Hello Mars!", "")
	if err != nil {
		t.Fatal(err)
	}
	if plain.ContentType != "text/plain" {
		t.Fatalf("want text/plain without a fallback, got %q", plain.ContentType)
	}
}

func TestBodyEmptyMarker(t *testing.T) {
	got, err := fx.Body("EMPTY", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Contents) != 0 {
		t.Fatalf("want an empty body, got %q", got.Contents)
	}
}

func TestBodyRejectsEmptyCell(t *testing.T) {
	_, err := fx.Body("  ", "")
	if err == nil {
		t.Fatal("want an error for an empty cell")
	}
}

func TestBodyMissingFixtureNamesIt(t *testing.T) {
	_, err := fx.Body("file: nope.json", "")
	if err == nil || !strings.Contains(err.Error(), "nope.json") {
		t.Fatalf("want an error naming nope.json, got %v", err)
	}
}

func TestBodyXMLFixtureCheckedOutWithCRLF(t *testing.T) {
	dir := t.TempDir()
	doc := "<body>\r\n<contentType>text/plain</contentType>\r\n" +
		"<contents eol=\"CRLF\"><![CDATA[line one\r\nline two]]></contents>\r\n</body>\r\n"
	err := os.WriteFile(filepath.Join(dir, "crlf-body.xml"), []byte(doc), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	got, err := fixtures.Loader{Dir: dir}.Body("file: crlf-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Contents) != "line one\r\nline two" {
		t.Fatalf("got %q", got.Contents)
	}
}
