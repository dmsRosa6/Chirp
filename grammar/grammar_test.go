package grammar

import "testing"

func TestLookupKnownCommand(t *testing.T) {
	spec, ok := Lookup("pub") // case-insensitive on purpose
	if !ok {
		t.Fatal("expected PUB to be found")
	}
	if spec.Name != "PUB" || spec.Args != 2 {
		t.Fatalf("got %+v, want Name=PUB Args=2", spec)
	}
}

func TestLookupUnknownCommand(t *testing.T) {
	if _, ok := Lookup("BOGUS"); ok {
		t.Fatal("expected BOGUS to be unknown")
	}
}

func TestEveryCommandHasUsage(t *testing.T) {
	for name, spec := range Commands {
		if spec.Usage == "" {
			t.Errorf("%s has no usage string", name)
		}
		if spec.Name != name {
			t.Errorf("map key %q doesn't match spec.Name %q", name, spec.Name)
		}
	}
}
