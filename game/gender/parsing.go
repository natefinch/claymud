package gender

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/natefinch/natemud/config"
)

const (
	templFile = "gender.toml"
)

// Initialize loads the gender configuration file from the data directory.
func Initialize() {
	filename := filepath.Join(config.DataDir(), templFile)

	f, err := os.Open(filename)
	if err != nil {
		panic(fmt.Errorf("Error reading gender config file: %s", err))
	}
	defer f.Close()
	names, err = decode(f)
	if err != nil {
		panic(err)
	}
	log.Print("Loaded gender config")
}

func decode(r io.Reader) (configuration, error) {
	cfg := configuration{}
	res, err := toml.DecodeReader(r, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("Error parsing gender config: %s", err)
	}
	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in gender config file: %v", und)
	}

	return cfg, nil
}
