package gender

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Initialize loads the gender configuration file from the data directory.
func Initialize(dir string) error {
	filename := filepath.Join(dir, "gender.toml")

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Error reading gender config file: %s", err)
	}
	defer f.Close()
	names, err = decode(f)
	if err != nil {
		return err
	}
	log.Printf("Loaded gender config from %q", filename)
	return nil
}

func decode(r io.Reader) (configuration, error) {
	cfg := configuration{}
	res, err := toml.DecodeReader(r, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse gender config: %s", err)
	}
	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in gender config file: %v", und)
	}

	return cfg, nil
}
