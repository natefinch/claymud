// Package util holds basic utility methods
package util

import (
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

// FillTemplate is a helper function that parses and fills a template with values from the map
func FillTemplate(t *template.Template, data interface{}) (result string, err error) {
	var b bytes.Buffer
	err = t.Execute(&b, data)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
