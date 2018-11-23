package util

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	rw, buf := rw("nate\n")
	answer, err := Query(rw, "hi!")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hi!" {
		t.Fatalf(`expected output "hi!" but got %q`, buf.String())
	}
	if answer != "nate" {
		t.Fatalf(`expected answer "nate", but got %q`, answer)
	}
}

// rw returns an io.ReadWriter that reads from input and outputs to the returned
// buffer.
func rw(input string) (io.ReadWriter, *bytes.Buffer) {
	in := strings.NewReader(input)
	buf := &bytes.Buffer{}

	return struct {
		io.Reader
		io.Writer
	}{
		Reader: in,
		Writer: buf,
	}, buf
}
