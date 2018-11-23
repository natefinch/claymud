package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/world"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init sets up the application's configuration directory.
func Init() (*Config, error) {
	dataDir := getDataDir()

	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("can't find data directory: %s", dataDir)
	}
	if err != nil {
		return nil, fmt.Errorf("can't read data directory %q: %s", dataDir, err)
	}

	logfile := filepath.Join(dataDir, "logs", "mud.log")

	// set some defaults
	cfg := Config{
		BcryptCost: 10,
		Logging:    &lumberjack.Logger{Filename: logfile},
	}
	cfgFile := filepath.Join(dataDir, "mud.toml")
	md, err := toml.DecodeFile(cfgFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %v", cfgFile, err)
	}

	// ignore any data dir specified in the config... it's not really supposed to be there.
	cfg.DataDir = dataDir

	if err := configLogging(cfg.Logging); err != nil {
		return nil, err
	}
	log.Printf("Using data directory %s", dataDir)

	if len(md.Undecoded()) > 0 {
		log.Printf("WARNING: unrecognized values in mud.toml: %v", md.Undecoded())
	}

	return &cfg, nil
}

// Config contains all the general configuration parameters for the mud.
type Config struct {
	DataDir    string          // config and data directory
	StartRoom  int             // the starting room number
	MainTitle  string          // title screen
	BcryptCost int             // work factor for auth
	Commands   world.CmdConfig // command aliases
	Logging    *lumberjack.Logger
	Direction  []game.Direction
	Gender     []game.Gender
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

func configLogging(lj *lumberjack.Logger) error {
	log.Printf("Logging to %s", lj.Filename)
	log.SetOutput(io.MultiWriter(lj, os.Stdout))

	log.Println("******************* ClayMUD Starting *******************")
	return nil
}
