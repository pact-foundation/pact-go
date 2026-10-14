package v2

import (
	"flag"
	"testing"
)

func setFlag(t *testing.T, name, value string) {
	t.Helper()

	f := flag.Lookup(name)
	if f == nil {
		t.Skipf("flag %s is not registered", name)
	}

	old := f.Value.String()

	err := flag.Set(name, value)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = flag.Set(name, old) })
}

func TestSandboxRunsOnceWhateverTheCallersFlagsSay(t *testing.T) {
	setFlag(t, "test.count", "3")
	setFlag(t, "test.skip", "sandbox")
	setFlag(t, "test.run", "NoSuchTest")
	setFlag(t, "test.failfast", "true")

	runs := 0

	sandbox("sandbox-under-test", func(t *testing.T) {
		t.Helper()

		runs++

		t.Error("an expected failure inside the sandbox")
	})

	if runs != 1 {
		t.Fatalf("want fn to run exactly once, ran %d times", runs)
	}
	if t.Failed() {
		t.Fatal("a sandboxed failure must not fail the caller")
	}
	if got := flag.Lookup("test.count").Value.String(); got != "3" {
		t.Fatalf("want test.count restored to 3, got %s", got)
	}
}
