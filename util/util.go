// Package util holds basic utility methods
package util

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"text/template"
)

var (
	// Error returned when a reader or writer is closed when you try to use it
	ErrClosed = errors.New("utils.ErrClosed: reader closed")
)

const (
	// Telnet color codes
	BLACK   = "\033[30m"
	RED     = "\033[31m"
	GREEN   = "\033[32m"
	YELLOW  = "\033[33m"
	BLUE    = "\033[34m"
	MAGENTA = "\033[35m"
	CYAN    = "\033[36m"
	WHITE   = "\033[37m"
)

// Id is a type that allows for unique identification of an object
type Id uint64

// IdGenerator creates a channel that will continuously generate new unique Ids
//
// Thanks to Andrew Rolfe of WolfMUD fame for this snippet of code - http://www.wolfmud.org
func IdGenerator() <-chan (Id) {
	// TODO: initialize at startup with highest Id created
	next := make(chan Id)
	id := Id(0)
	go func() {
		for {
			next <- id
			id++
		}
	}()
	return next
}

// ReadLn is a helper function to encapsulate reading a line from a reader
func ReadLn(r *bufio.ReadWriter) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from reader: %v", err)
		return s, ErrClosed
	}
	// drop the trailing \r\n (all telnet lines end with \r\n per the RFCs)
	s = s[:len(s)-2]
	return s, nil
}

// Write is a helper function to encapsulate writing and flushing text to a writer
func Write(w *bufio.ReadWriter, s string) error {
	_, err := w.WriteString(s)
	if err != nil {
		return ErrClosed
	}
	err = w.Flush()
	if err != nil {
		return ErrClosed
	}
	return nil
}

// WriteLn is a helper function to encapsulate writing and flusing a full line to a writer 
//
// a carriage return and newline will be appended to the given string
func WriteLn(w *bufio.ReadWriter, s string) error {
	_, err := w.WriteString(s)
	if err != nil {
		return ErrClosed
	}
	_, err = w.WriteString("\r\n")
	if err != nil {
		return ErrClosed
	}
	err = w.Flush()
	if err != nil {
		return ErrClosed
	}
	return nil
}

// FillTemplate is a helper function that parses and fills a template with values from the map
func FillTemplate(templ string, params map[string]string) (result string, err error) {
	t, err := template.New("template").Parse(templ)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, params)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return "", err
	}
	return b.String(), nil
}
