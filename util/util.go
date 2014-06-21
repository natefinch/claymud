// Package util holds basic utility methods
package util

import (
	"fmt"
	"io"
	"text/template"
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

// templ is a struct that lets us unmarshal directly into a template.
type Template struct {
	*template.Template
}

// UnmarshalText implements TextUnmarshaler.UnmarshalText.
func (e *Template) UnmarshalText(text []byte) error {
	var err error
	e.Template, err = template.New("template").Parse(string(text))
	if err != nil {
		return fmt.Errorf("can't parse template %q: %#v", text, err)
	}
	return nil
}

type SafeWriter struct {
	Writer io.Writer
	OnErr  func(error)
}

func (s SafeWriter) Write(b []byte) (int, error) {
	n, err := s.Writer.Write(b)
	if err != nil {
		s.OnErr(err)
	}
	return n, nil
}
