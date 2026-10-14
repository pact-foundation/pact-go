package steps

import "testing"

func TestNothingPublishedNeedsABroker(t *testing.T) {
	w := &world{}
	err := w.nothingPublished()
	if err == nil {
		t.Fatal("want an error when no broker was used")
	}
}
