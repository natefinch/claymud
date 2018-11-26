package util

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	scanner, buf := scanner("nate\n")
	answer, err := Query(scanner, "hi!")
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

// scanner returns an io.ReadWriter that reads from input and outputs to the returned
// buffer.
func scanner(input string) (WriteScanner, *bytes.Buffer) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	buf := &bytes.Buffer{}

	return struct {
		*bufio.Scanner
		io.Writer
	}{
		Scanner: scanner,
		Writer:  buf,
	}, buf
}
