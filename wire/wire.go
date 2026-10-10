package wire

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Hard caps so a client can't make the server allocate gigabytes.
const (
	MaxArgs    = 8
	MaxBulkLen = 1 << 20 // 1 MiB
)

// ProtocolError means the peer sent bytes that are not valid framing. The
// connection can't be trusted afterwards, so callers should reply (if they
// want) and close.
type ProtocolError struct{ msg string }

func (e *ProtocolError) Error() string { return "protocol error: " + e.msg }

func protoErr(format string, args ...any) error {
	return &ProtocolError{msg: fmt.Sprintf(format, args...)}
}

// *3\r\n$3\r\nPUB\r\n$3\r\nfoo\r\n$5\r\nhello\r\n
func WriteCommand(w io.Writer, args ...string) error {
	return writeArray(w, args)
}

// WriteMessage writes the push frame delivered to a subscriber:
// *4\r\n$3\r\nMSG\r\n$<n>\r\n<sub_id>\r\n$<n>\r\n<subject>\r\n$<n>\r\n<payload>\r\n
func WriteMessage(w io.Writer, subID, subject, payload string) error {
	return writeArray(w, []string{"MSG", subID, subject, payload})
}

func writeArray(w io.Writer, args []string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func ReadCommand(r *bufio.Reader) ([]string, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 || line[0] != '*' {
		return nil, protoErr("expected '*', got %q", line)
	}
	return readArray(r, line)
}

// readArray parses an array whose "*<n>" header line has already been read.
func readArray(r *bufio.Reader, header string) ([]string, error) {
	n, err := strconv.Atoi(header[1:])
	if err != nil || n < 0 {
		return nil, protoErr("bad array length %q", header)
	}
	if n > MaxArgs {
		return nil, protoErr("too many arguments (%d, max %d)", n, MaxArgs)
	}

	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s, err := readBulk(r)
		if err != nil {
			return nil, err
		}
		args = append(args, s)
	}
	return args, nil
}

func readBulk(r *bufio.Reader) (string, error) {
	head, err := readLine(r)
	if err != nil {
		return "", err
	}
	if len(head) == 0 || head[0] != '$' {
		return "", protoErr("expected '$', got %q", head)
	}
	return readBulkBody(r, head)
}

// readBulkBody reads the payload for a "$<n>" header that is already consumed.
func readBulkBody(r *bufio.Reader, head string) (string, error) {
	size, err := strconv.Atoi(head[1:])
	if err != nil || size < 0 {
		return "", protoErr("bad bulk length %q", head)
	}
	if size > MaxBulkLen {
		return "", protoErr("bulk too large (%d bytes, max %d)", size, MaxBulkLen)
	}
	buf := make([]byte, size+2) // payload + trailing \r\n
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	if buf[size] != '\r' || buf[size+1] != '\n' {
		return "", protoErr("bulk payload not terminated by CRLF")
	}
	return string(buf[:size]), nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func WriteOK(w io.Writer) error { return WriteSimple(w, "OK") }

func WriteSimple(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "+%s\r\n", s)
	return err
}

func WriteError(w io.Writer, msg string) error {
	_, err := fmt.Fprintf(w, "-ERR %s\r\n", msg)
	return err
}

func WriteBulk(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(s), s)
	return err
}

type ReplyType int

const (
	SimpleString ReplyType = iota // +OK
	ErrorReply                    // -ERR ...
	BulkString                    // $<len>\r\n<bytes>\r\n
	PushMessage                   // *4 MSG <sub_id> <subject> <payload>
)

type Reply struct {
	Type  ReplyType
	Value string   // SimpleString, ErrorReply, BulkString
	Args  []string // PushMessage: ["MSG", sub_id, subject, payload]
}

func ReadReply(r *bufio.Reader) (Reply, error) {
	line, err := readLine(r)
	if err != nil {
		return Reply{}, err
	}
	if len(line) == 0 {
		return Reply{}, protoErr("empty reply line")
	}

	switch line[0] {
	case '+':
		return Reply{Type: SimpleString, Value: line[1:]}, nil
	case '-':
		return Reply{Type: ErrorReply, Value: line[1:]}, nil
	case '$':
		s, err := readBulkBody(r, line)
		if err != nil {
			return Reply{}, err
		}
		return Reply{Type: BulkString, Value: s}, nil
	case '*':
		args, err := readArray(r, line)
		if err != nil {
			return Reply{}, err
		}
		if len(args) != 4 || args[0] != "MSG" {
			return Reply{}, protoErr("unexpected array reply %q", args)
		}
		return Reply{Type: PushMessage, Args: args}, nil
	default:
		return Reply{}, protoErr("unknown reply type %q", line[0])
	}
}

// IsProtocolError reports whether err is (or wraps) a *ProtocolError.
func IsProtocolError(err error) bool {
	var pe *ProtocolError
	return errors.As(err, &pe)
}
