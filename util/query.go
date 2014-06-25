package util

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// Query writes the question to rw and waits for an answer.
func Query(rw io.ReadWriter, question []byte) (answer string, err error) {
	_, err = rw.Write(question)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(rw)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("Connection closed")
	}
	return scanner.Text(), nil
}

// Query writes a question to rw and waits for an answer.  It will pass the
// answer into the verify function.  Verify should check the answer for
// validity, returning a failure reason as a string, or an empty string if the
// answer is valid.
func QueryVerify(
	rw io.ReadWriter,
	question []byte,
	verify func(string) (string, error),
) (answer string, err error) {
	scanner := bufio.NewScanner(rw)
	for {
		_, err = rw.Write(question)
		if err != nil {
			return "", err
		}

		if !scanner.Scan() {
			if err = scanner.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("Connection closed")
		}
		answer = scanner.Text()
		failure, err := verify(answer)
		if err != nil {
			return "", err
		}
		if failure == "" {
			return answer, nil
		}
		_, err = fmt.Fprintln(rw, failure)
		if err != nil {
			return "", err
		}
	}
}

type Opt struct {
	Key  rune
	Text []byte
}

// Query writes a question to rw and waits for an answer.  It will pass the
// answer into the verify function.  Verify should check the answer for
// validity, returning a failure reason as a string, or an empty string if the
// answer is valid.
func QueryOptions(
	rw io.ReadWriter,
	question []byte,
	options ...Opt,
) (answer rune, err error) {
	_, err = rw.Write(question)
	if err != nil {
		return 0, err
	}

	for _, opt := range options {
		_, err = fmt.Fprintf(rw, "%c.) %s\n", opt.Key, opt.Text)
		if err != nil {
			return 0, err
		}
	}

	scanner := bufio.NewScanner(rw)
	for {
		fmt.Fprint(rw, "\nPlease choose one of the options above: ")

		if !scanner.Scan() {
			if err = scanner.Err(); err != nil {
				return utf8.RuneError, err
			}
			return utf8.RuneError, fmt.Errorf("Connection closed")
		}
		b := scanner.Bytes()
		if len(b) == 0 {
			continue
		}
		r, _ := utf8.DecodeRune(b)
		if r == utf8.RuneError {
			continue
		}
		return r, nil
	}
}
