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
	"github.com/natefinch/lumberjack"
)

var (
	mainTitle string
	dataDir   string
)

// Initialize sets up the application's configuration directory.
func Initialize() error {
	if err := setupDataDir(); err != nil {
		return err
	}
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

// DataDir returns the directory where natemud configuration and game data is
// stored.  By default it is in ~/.config/
func DataDir() string {
	return dataDir
}

func setupDataDir() error {
	dataDir = getDataDir()
	log.Printf("Using data directory %s", dataDir)

	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("can't create datadir: %s", err)
		}
	} else if err != nil {
		// some other error
		return fmt.Errorf("can't get info about data dir: %s", err)
	}

	return copyGopath()
}

func getDataDir() string {
	v := os.Getenv("NATEMUD_DATADIR")
	if v != "" {
		return v
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("USERPROFILE"), "NateMUD")
	}

	return filepath.Join(os.Getenv("HOME"), ".config", "natemud")
}

func copyFile(name, dir string) error {
	src, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("can't read file from config dir: %s", err)
	}
	defer src.Close()
	newname := filepath.Join(dir, filepath.Base(name))

	log.Printf("Copying %s", filepath.Base(name))

	dst, err := os.OpenFile(newname, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("can't create new file for config dir: %s", err)
	}
	defer src.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("can't create new file for config dir: %s", err)
	}
	return nil
}

func copyGopath() error {
	// TODO: this is a dirty hack, find a better way to do this
	// let's copy the data from the repo to the expected directory.
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("can't get file list from data dir: %s", err)
	}
	if len(files) > 0 {
		// folder already populated
		return nil
	}

	gopath := os.Getenv("GOPATH")

	if gopath == "" {
		return nil
	}

	paths := filepath.SplitList(gopath)
	for _, p := range paths {
		p = filepath.Join(p, "src", "github.com", "natefinch", "natemud", "data")
		if _, err := os.Stat(p); err != nil {
			continue
		}
		log.Printf("Dev setup: copying files from repo dir %q to dataDir\n", p)
		infos, err := ioutil.ReadDir(p)
		if err != nil {
			return fmt.Errorf("can't read data files from config dir %q: %s", p, err)
		}
		for _, info := range infos {
			filename := filepath.Join(p, info.Name())
			if err := copyFile(filename, dataDir); err != nil {
				return err
			}
		}
		break
	}
	return nil
}

func configLogging() error {
	path := filepath.Join(dataDir, "logs.toml")
	lj := &lumberjack.Logger{}
	md, err := toml.DecodeFile(path, lj)
	if err != nil {
		return fmt.Errorf("can't decode logging config file %q: %s", path, err)
	}
	if lj.Dir == "" {
		lj.Dir = filepath.Join(dataDir, "logs")
	}
	log.Printf("Logging to %s", lj.Dir)
	log.SetOutput(io.MultiWriter(lj, os.Stdout))

	log.Println("******************* NateMUD Starting *******************")
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
		mainTitle = "Welcome to NateMUD"
	} else {
		mainTitle = string(b)
	}
}
