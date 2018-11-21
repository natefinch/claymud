// Package world holds most of the MUD code, including locations, players, etc
package world

import (
	"sync"
)

var players = map[string]*Player{}

// addPlayer adds a new player to the world list.
func addPlayer(p *Player) {
	players[p.Name()] = p
}

// removePlayer removes a player from the world list.
func removePlayer(p *Player) {
	delete(players, p.Name())
}

// FindPlayer returns the player for the given name.
func FindPlayer(name string) (*Player, bool) {
	p, ok := players[name]
	return p, ok
}

// Initialize spawns the zones and their attendant workers, creates all areas
// and locations.
func Initialize(datadir string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) error {
	if err := loadLocTempl(); err != nil {
		return err
	}

	return genWorld(datadir, zoneLock, shutdown, wg)
}
