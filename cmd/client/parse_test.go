package main

import (
	"reflect"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"PING", []string{"PING"}},
		{"SUB foo.* s1", []string{"SUB", "foo.*", "s1"}},
		{"  UNSUB   s1 ", []string{"UNSUB", "s1"}},
		{"PUB foo hello", []string{"PUB", "foo", "hello"}},
		{"pub foo hello world", []string{"pub", "foo", "hello world"}},
		{"PUB foo   keeps   inner   spaces", []string{"PUB", "foo", "keeps   inner   spaces"}},
		// The old SplitN(line, subject, 2) approach broke on this: "P" is inside "PUB".
		{"PUB P hi", []string{"PUB", "P", "hi"}},
		// ...and on a subject that also appears in the payload.
		{"PUB foo foo foo", []string{"PUB", "foo", "foo foo"}},
		{"PUB foo", []string{"PUB", "foo"}},
		{"PUB", []string{"PUB"}},
	}
	for _, tc := range tests {
		got := parseLine(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseLine(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}
