package wire

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestPushRoundTripWithCRLFInPayload(t *testing.T) {
	payload := "line one\r\nline two\r\n"
	var buf bytes.Buffer
	if err := WriteMessage(&buf, "sub1", "foo.bar", payload); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	reply, err := ReadReply(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadReply: %v", err)
	}
	if reply.Type != PushMessage {
		t.Fatalf("type = %v, want PushMessage", reply.Type)
	}
	want := []string{"MSG", "sub1", "foo.bar", payload}
	for i := range want {
		if reply.Args[i] != want[i] {
			t.Fatalf("arg %d: got %q, want %q", i, reply.Args[i], want[i])
		}
	}
}

func TestReadReplyRejectsOtherArrays(t *testing.T) {
	var buf bytes.Buffer
	WriteCommand(&buf, "PUB", "foo", "hello")
	if _, err := ReadReply(bufio.NewReader(&buf)); err == nil {
		t.Fatal("expected an error for a non-MSG array reply")
	}
}

func TestTooManyArgs(t *testing.T) {
	var buf bytes.Buffer
	args := make([]string, MaxArgs+1)
	for i := range args {
		args[i] = "x"
	}
	WriteCommand(&buf, args...)
	_, err := ReadCommand(bufio.NewReader(&buf))
	if !IsProtocolError(err) {
		t.Fatalf("got %v, want a protocol error", err)
	}
}

func TestBulkTooLarge(t *testing.T) {
	// The header lies about a huge payload; we must reject before allocating.
	in := "*1\r\n$99999999999\r\n"
	_, err := ReadCommand(bufio.NewReader(strings.NewReader(in)))
	if !IsProtocolError(err) {
		t.Fatalf("got %v, want a protocol error", err)
	}
}

func TestBulkMissingCRLF(t *testing.T) {
	in := "*1\r\n$3\r\nabcXX"
	_, err := ReadCommand(bufio.NewReader(strings.NewReader(in)))
	if !IsProtocolError(err) {
		t.Fatalf("got %v, want a protocol error", err)
	}
}

func TestBadFrameIsProtocolError(t *testing.T) {
	_, err := ReadCommand(bufio.NewReader(strings.NewReader("garbage\r\n")))
	if !IsProtocolError(err) {
		t.Fatalf("got %v, want a protocol error", err)
	}
}
