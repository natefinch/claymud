// Package config package handles loading of configuration files that customize the mud
// for things such as exit names, supported emotes, and other strings
package config

import (
	"io/ioutil"
	"log"
)

var (
	mainTitle string
)

// LoadMainTitle loads text from config/maintitle.txt that is shown to users when they connect
func LoadMainTitle() {
	b, err := ioutil.ReadFile("config/maintitle.txt")
	if err != nil {
		log.Printf("Couldn't read maintitle, using default")
		mainTitle = "Welcome to NateMUD"
	} else {
		mainTitle = string(b)
	}
}

// MainTitle returns the text that is shown to users when they connect, before logging in
func MainTitle() string {
	return mainTitle
}

// init loads the main title text
func init() {
	LoadMainTitle()
}
