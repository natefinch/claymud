// Package util holds basic utility methods
package util

import (
	"errors"
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
