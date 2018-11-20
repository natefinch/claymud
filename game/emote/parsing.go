package emote

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
	templFile = "emotes.toml"
)

var arrival *noTarget

// DoArrival runs the standard emote that occurs when you
func DoArrival(actor Person, others io.Writer) {
	performToNoOne("arrival", arrival, emoteData{Actor: actor}, actor, others)
}

// Initialize creates the emoteTemplate map and loads emotes into it.
func Initialize(dir string) error {
	filename := filepath.Join(dir, templFile)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Error reading emote config file: %s", err)
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

	if err := loadEmotes(cfg.Emotes); err != nil {
		return err
	}

	log.Printf("Loaded emotes: %v", Names)
	return nil
}

// emoteConfig is a struct for getting the emote templates out of a config.
type emoteConfig struct {
	// the emote for a player arriving in the world at the starting location.
	Arrival noTarget

	// yes, the toml is emote, singular. This lets the section header be
	// [[emote]] instead of [[emotes]] which is a lot clearer.
	Emotes []emote `toml:"emote"`
}

// decodeEmotes parses the data from the reader into a list of emotes.
func decodeConfig(r io.Reader) (*emoteConfig, error) {
	cfg := emoteConfig{}

	res, err := toml.DecodeReader(r, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error parsing emote config file: %s", err)
	}

	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in emote config file: %v", und)
	}
	return &cfg, nil
}

// loadEmotes populates the game's list of emotes and checks for duplicates.
func loadEmotes(em []emote) error {
	emotes = make(map[string]emote, len(em))
	Names = make([]string, len(em))
	for i, emote := range em {
		if _, ok := emotes[emote.Name]; ok {
			return fmt.Errorf("Duplicate emote defined: %q", emote.Name)
		}

		Names[i] = emote.Name
		emotes[emote.Name] = emote
	}
	return nil
}
