// Package world holds most of the MUD code, including locations, players, etc
package world

import (
	"sort"
	"sync"
)

var playerMap = map[string]*Player{}
var playerList = &sortedPlayers{}

type sortedPlayers []*Player

func (s *sortedPlayers) add(p *Player) {
	*s = append(*s, p)
	sort.Slice(*s, func(i, j int) bool { return (*s)[i].Name() < (*s)[j].Name() })
}
func (s *sortedPlayers) remove(p *Player) {
	for i, pl := range *s {
		if pl.Is(p) {
			*s = append((*s)[:i], (*s)[i+1:]...)
			break
		}
	}
}

// addPlayer adds a new player to the world list.
func addPlayer(p *Player) {
	playerMap[p.Name()] = p
	playerList.add(p)
}

// removePlayer removes a player from the world list.
func removePlayer(p *Player) {
	delete(playerMap, p.Name())
	playerList.remove(p)
}

// FindPlayer returns the player for the given name.
func FindPlayer(name string) (*Player, bool) {
	p, ok := playerMap[name]
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
