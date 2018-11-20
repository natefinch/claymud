package util

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// Query writes the question to rw and waits for an answer.
func Query(rw io.ReadWriter, question string) (answer string, err error) {
	// need this because scan can panic if you send it too much stuff
	defer func() {
		panicErr := recover()
		if panicErr == nil {
			return
		}
		if e, ok := panicErr.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("%v", panicErr)
	}()
	_, err = io.WriteString(rw, question)
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
	question string,
	verify func(string) (string, error),
) (answer string, err error) {
	// need this because scan can panic if you send it too much stuff
	defer func() {
		panicErr := recover()
		if panicErr == nil {
			return
		}
		if e, ok := panicErr.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("%v", panicErr)
	}()
	scanner := bufio.NewScanner(rw)
	for {
		_, err = io.WriteString(rw, question)
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
	Text string
}

// Query writes a question to rw and waits for an answer.  It will pass the
// answer into the verify function.  Verify should check the answer for
// validity, returning a failure reason as a string, or an empty string if the
// answer is valid.
func QueryOptions(
	rw io.ReadWriter,
	question string,
	options ...Opt,
) (answer rune, err error) {
	// need this because scan can panic if you send it too much stuff
	defer func() {
		panicErr := recover()
		if panicErr == nil {
			return
		}
		if e, ok := panicErr.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("%v", panicErr)
	}()

	_, err = io.WriteString(rw, question)
	if err != nil {
		return utf8.RuneError, err
	}

	for _, opt := range options {
		_, err = fmt.Fprintf(rw, "%c.) %s\n", opt.Key, opt.Text)
		if err != nil {
			return utf8.RuneError, err
		}
	}

	scanner := bufio.NewScanner(rw)
	for {
		_, err := io.WriteString(rw, "\nPlease choose one of the options above: ")
		if err != nil {
			return utf8.RuneError, err
		}
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
		for _, o := range options {
			if r == o.Key {
				return r, nil
			}
		}
	}
}
