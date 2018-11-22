package social

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
)

const (
	templFile = "socials.toml"
)

var arrival *noTarget

// DoArrival runs the standard social that occurs when you
func DoArrival(actor Person, others io.Writer) {
	performToNoOne("arrival", arrival, socialData{Actor: actor}, actor, others)
}

// Initialize creates the socialTemplate map and loads socials into it.
func Initialize(dir string) error {
	filename := filepath.Join(dir, templFile)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Error reading social config file: %s", err)
	}
	defer f.Close()
	cfg, err := decodeConfig(f)
	if err != nil {
		return err
	}
	arrival = &cfg.Arrival
	if arrival.Self.Template == nil {
		arrival.Self.Template = template.Must(template.New("arrival.self").Parse("You arrive in a puff of smoke."))
	}
	if arrival.Around.Template == nil {
		arrival.Around.Template = template.Must(template.New("arrival.around").Parse("{{Actor}} arrives in a puff of smoke."))
	}

	if err := loadSocials(cfg.Socials); err != nil {
		return err
	}

	log.Printf("Loaded socials: %v", Names)
	return nil
}

// socialConfig is a struct for getting the social templates out of a config.
type socialConfig struct {
	// the social for a player arriving in the world at the starting location.
	Arrival noTarget

	// yes, the toml is social, singular. This lets the section header be
	// [[social]] instead of [[socials]] which is a lot clearer.
	Socials []social `toml:"social"`
}

// decodeSocials parses the data from the reader into a list of socials.
func decodeConfig(r io.Reader) (*socialConfig, error) {
	cfg := socialConfig{}

	res, err := toml.DecodeReader(r, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error parsing social config file: %s", err)
	}

	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in social config file: %v", und)
	}
	return &cfg, nil
}

// loadSocials populates the game's list of socials and checks for duplicates.
func loadSocials(em []social) error {
	socials = make(map[string]social, len(em))
	Names = make([]string, len(em))
	for i, social := range em {
		if _, ok := socials[social.Name]; ok {
			return fmt.Errorf("Duplicate social defined: %q", social.Name)
		}

		Names[i] = social.Name
		socials[social.Name] = social
	}
	return nil
}
