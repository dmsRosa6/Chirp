package wire

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// *3\r\n$3\r\nPUB\r\n$3\r\nfoo\r\n$5\r\nhello\r\n
func WriteCommand(w io.Writer, args ...string) error {
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
		return nil, fmt.Errorf("protocol error: expected '*', got %q", line)
	}
	n, err := strconv.Atoi(line[1:])
	if err != nil || n < 0 {
		return nil, fmt.Errorf("protocol error: bad array length %q", line)
	}

	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		head, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if len(head) == 0 || head[0] != '$' {
			return nil, fmt.Errorf("protocol error: expected '$', got %q", head)
		}
		size, err := strconv.Atoi(head[1:])
		if err != nil || size < 0 {
			return nil, fmt.Errorf("protocol error: bad bulk length %q", head)
		}
		buf := make([]byte, size+2) // payload + trailing \r\n
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
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
