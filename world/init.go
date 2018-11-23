// Package world holds most of the MUD code, including locations, players, etc
package world

import (
	"sync"
)

// Config determines the configuration of the world.  This affects the
type Config struct {
	StartRoom int
	Commands  CmdConfig
}

// Initialize spawns the zones and their attendant workers, creates all areas
// and locations.
func Init(cfg Config, datadir string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) error {
	if err := loadLocTempl(datadir); err != nil {
		return err
	}
	initCommands(cfg.Commands)
	return genWorld(datadir, zoneLock, shutdown, wg)
}
