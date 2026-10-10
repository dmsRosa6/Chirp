package data_structs

import (
	"sort"
	"testing"
)

func newTestTrie(patterns ...string) *Trie[string] {
	t := NewTrie[string](".", "*")
	for _, p := range patterns {
		t.Add(p, p)
	}
	return t
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMatch(t *testing.T) {
	tr := newTestTrie("foo.bar", "foo.b*", "foo.*", "foo*", "x.y", "*")

	tests := []struct {
		subject string
		want    []string
	}{
		{"foo.bar", []string{"foo.b*", "foo.bar", "foo.*"}},
		{"foo.baz", []string{"foo.b*", "foo.*"}},
		{"foo.b", []string{"foo.b*", "foo.*"}},
		{"foo.x", []string{"foo.*"}},
		{"foo", []string{"foo*", "*"}},
		{"foobar", []string{"foo*", "*"}},
		{"x.y", []string{"x.y"}},
		{"bar", []string{"*"}},
		{"foo.bar.baz", nil}, // one level only
		{"nope.bar", nil},    // missing parent
		{"x", []string{"*"}},
	}
	for _, tc := range tests {
		got := sorted(tr.Match(tc.subject))
		want := sorted(tc.want)
		if !equal(got, want) {
			t.Errorf("Match(%q) = %v, want %v", tc.subject, got, want)
		}
	}
}

func TestMatchIgnoresPathOnlyNodes(t *testing.T) {
	// "a.b" exists only as a path to "a.b.c"; it has no value of its own.
	tr := newTestTrie("a.b.c")
	if got := tr.Match("a.b"); len(got) != 0 {
		t.Fatalf("Match(a.b) = %v, want nothing", got)
	}
}

func TestFindOneIsExact(t *testing.T) {
	tr := newTestTrie("foo.b*")
	if _, ok := tr.FindOne("foo.bar"); ok {
		t.Fatal("FindOne must not interpret wildcards")
	}
	if v, ok := tr.FindOne("foo.b*"); !ok || *v != "foo.b*" {
		t.Fatalf("FindOne(foo.b*) = %v, %v", v, ok)
	}
}

func TestRemove(t *testing.T) {
	tr := newTestTrie("a.b.c", "a.b", "a.x")

	if tr.Remove("a.b.q") {
		t.Fatal("removing a missing path should report false")
	}
	if tr.Remove("a") {
		t.Fatal("removing a value-less path node should report false")
	}

	if !tr.Remove("a.b.c") {
		t.Fatal("expected a.b.c to be removed")
	}
	if _, ok := tr.FindOne("a.b.c"); ok {
		t.Fatal("a.b.c still present")
	}
	if _, ok := tr.FindOne("a.b"); !ok {
		t.Fatal("a.b has its own value and must survive")
	}
	if _, ok := tr.root.children["a"].children["b"].children["c"]; ok {
		t.Fatal("empty node c was not pruned")
	}

	// Removing the rest prunes everything up to the root.
	tr.Remove("a.b")
	tr.Remove("a.x")
	if len(tr.root.children) != 0 {
		t.Fatalf("root should be empty, has %d children", len(tr.root.children))
	}
}
