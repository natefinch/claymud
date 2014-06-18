package emote

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
	templFile = "emotes.toml"
)

// Initialize creates the emoteTemplate map and loads emotes into it.
func Initialize() error {
	filename := filepath.Join(config.DataDir(), templFile)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Error reading emote config file: %s", err)
	}
	defer f.Close()
	em, err := decodeEmotes(f)
	if err != nil {
		return err
	}

	if err := loadEmotes(em); err != nil {
		return err
	}

	log.Printf("Loaded emotes: %v", Names)
	return nil
}

// emoteConfigs is a struct for getting the emote templates out of a config.
type emoteConfigs struct {
	// yes, Emote, singular. This lets the section header be [[emote]] instead
	// of [[emotes]] which is a lot clearer.
	Emote []emote
}

// decodeEmotes parses the data from the reader into a list of emotes.
func decodeEmotes(r io.Reader) ([]emote, error) {
	cfgs := emoteConfigs{}

	res, err := toml.DecodeReader(r, &cfgs)
	if err != nil {
		return nil, fmt.Errorf("Error parsing emote config file: %s", err)
	}

	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in emote config file: %v", und)
	}
	return cfgs.Emote, nil
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
