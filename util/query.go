package util

import (
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

// WriteScanner merges an io.Writer and a bufio.Scanner.
type WriteScanner interface {
	io.Writer
	Scanner
}

// Scanner represents a bufio.Scanner.
type Scanner interface {
	Scan() bool
	Err() error
	Text() string
	Bytes() []byte
}

// Query writes the question to rw and waits for an answer.
func Query(ws WriteScanner, question string) (answer string, err error) {
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
	_, err = io.WriteString(ws, question)
	if err != nil {
		return "", err
	}

	if !ws.Scan() {
		if err = ws.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("Connection closed")
	}
	return ws.Text(), nil
}

// Query writes a question to rw and waits for an answer.  It will pass the
// answer into the verify function.  Verify should check the answer for
// validity, returning a failure reason as a string, or an empty string if the
// answer is valid.
func QueryVerify(
	ws WriteScanner,
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
	for {
		_, err = io.WriteString(ws, question)
		if err != nil {
			return "", err
		}

		if !ws.Scan() {
			if err = ws.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("Connection closed")
		}
		answer = ws.Text()
		failure, err := verify(answer)
		if err != nil {
			return "", err
		}
		if failure == "" {
			return answer, nil
		}
		_, err = fmt.Fprintln(ws, failure)
		if err != nil {
			return "", err
		}
	}
}

// Opt represents an option you can choose from a list
type Opt struct {
	Key  rune
	Text string
}

// QueryStrings generates an automatic options list from the given options
// and returns the chosen index.
func QueryStrings(
	ws WriteScanner,
	question string,
	defaultIndex int,
	options ...string,
) (index int, err error) {
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
	_, err = io.WriteString(ws, question)
	if err != nil {
		return -1, err
	}

	for i, s := range options {
		if i == defaultIndex {
			_, err = fmt.Fprintf(ws, "%d - %s (default)\n", i+1, s)
			if err != nil {
				return -1, err
			}
		} else {
			_, err = fmt.Fprintf(ws, "%d - %s\n", i+1, s)
			if err != nil {
				return -1, err
			}
		}
	}

	for {
		_, err := io.WriteString(ws, "\nPlease choose one of the options above: ")
		if err != nil {
			return -1, err
		}
		if !ws.Scan() {
			if err = ws.Err(); err != nil {
				return -1, err
			}
			return -1, fmt.Errorf("Connection closed")
		}
		choice := ws.Text()
		if len(choice) == 0 {
			if defaultIndex > -1 {
				return defaultIndex, nil
			}
			continue
		}
		i, err := strconv.Atoi(choice)
		if err != nil {
			continue
		}
		if i < 1 || i > len(options) {
			continue
		}
		return i - 1, nil
	}
}

// QueryOptions writes a question to rw and waits for an answer.  If Default is
// not 0, the given option is returns if the user hits enter.
func QueryOptions(
	ws WriteScanner,
	question string,
	Default rune,
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

	_, err = io.WriteString(ws, question)
	if err != nil {
		return utf8.RuneError, err
	}

	for _, opt := range options {
		if opt.Key == Default {
			_, err = fmt.Fprintf(ws, "%c - %s (default)\n", opt.Key, opt.Text)
			if err != nil {
				return utf8.RuneError, err
			}
		} else {
			_, err = fmt.Fprintf(ws, "%c - %s\n", opt.Key, opt.Text)
			if err != nil {
				return utf8.RuneError, err
			}
		}
	}

	for {
		_, err := io.WriteString(ws, "\nPlease choose one of the options above: ")
		if err != nil {
			return utf8.RuneError, err
		}
		if !ws.Scan() {
			if err = ws.Err(); err != nil {
				return utf8.RuneError, err
			}
			return utf8.RuneError, fmt.Errorf("Connection closed")
		}
		b := ws.Bytes()
		if len(b) == 0 {
			if Default != 0 {
				return Default, nil
			}
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
