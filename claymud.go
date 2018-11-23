// Package main is the entry point for ClayMUD.  This defines the command line
// args and listens for incoming connections.
package main

import (
	"log"
	"os"

	"github.com/natefinch/claymud/server"
)

func main() {
	if err := server.Main(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
