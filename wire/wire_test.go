package wire

import (
	"bufio"
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCommand(&buf, "PUB", "foo", "hello world"); err != nil {
		t.Fatalf("WriteCommand: %v", err)
	}

	got, err := ReadCommand(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadCommand: %v", err)
	}
	want := []string{"PUB", "foo", "hello world"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPayloadWithSpacesAndNewlines(t *testing.T) {
	// This is exactly the case a whitespace-splitting lexer has to special
	// case ("Remainder" args). Length-prefixing makes it a non-issue.
	payload := "line one\nline two   with   spaces"
	var buf bytes.Buffer
	WriteCommand(&buf, "PUB", "foo", payload)

	got, err := ReadCommand(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadCommand: %v", err)
	}
	if got[2] != payload {
		t.Fatalf("payload mismatch: got %q, want %q", got[2], payload)
	}
}

func TestBadFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("not-a-frame\r\n")
	if _, err := ReadCommand(bufio.NewReader(&buf)); err == nil {
		t.Fatal("expected a protocol error, got nil")
	}
}
