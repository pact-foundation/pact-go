package steps

import "testing"

func TestNothingPublishedNeedsABroker(t *testing.T) {
	w := &world{}
	if err := w.nothingPublished(); err == nil {
		t.Fatal("want an error when no broker was used")
	}
}
