package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	mainTitle string
	dataDir   string
)

// Initialize sets up the application's configuration directory.
func Initialize() error {
	dataDir = getDataDir()

	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("can't find data directory: %s", dataDir)
	}
	if err != nil {
		return fmt.Errorf("can't read data directory: %s", err)
	}

	log.Printf("Using data directory %s", dataDir)

	if err := configLogging(); err != nil {
		return err
	}
	loadMainTitle()
	return nil
}

// MainTitle returns the text that is shown to users when they connect, before
// logging in.
func MainTitle() string {
	return mainTitle
}

// DataDir returns the directory where claymud configuration and game data is
// stored.  By default it is in ~/.config/
func DataDir() string {
	return dataDir
}

// getDataDir returns the platform-specific data directory.
func getDataDir() string {
	v := os.Getenv("CLAYMUD_DATADIR")
	if v != "" {
		return v
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("USERPROFILE"), "ClayMUD")
	}

	return filepath.Join(os.Getenv("HOME"), ".config", "claymud")
}

func configLogging() error {
	path := filepath.Join(dataDir, "logs.toml")
	lj := &lumberjack.Logger{}
	md, err := toml.DecodeFile(path, lj)
	if err != nil {
		return fmt.Errorf("can't decode logging config file %q: %s", path, err)
	}
	if lj.Filename == "" {
		lj.Filename = filepath.Join(dataDir, "logs", "mud.log")
	}
	log.Printf("Logging to %s", lj.Filename)
	log.SetOutput(io.MultiWriter(lj, os.Stdout))

	log.Println("******************* ClayMUD Starting *******************")
	log.Printf("Using data directory %s", dataDir)
	if len(md.Undecoded()) > 0 {
		log.Printf("WARNING: unrecognized values in logging config: %v", md.Undecoded())
	}
	return nil
}

// loadMainTitle loads the text that is shown to users when they first connect.
func loadMainTitle() {
	path := filepath.Join(dataDir, "maintitle.txt")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("WARNING: Couldn't read maintitle from path %q: %s", path, err)
		log.Printf("Using default title.")
		mainTitle = "Welcome to ClayMUD"
	} else {
		mainTitle = string(b)
	}
}
