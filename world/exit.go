package world

import (
	"github.com/natefinch/natemud/config"
)

// Exit is a direction that connects to a location.
type Exit struct {
	config.Direction
	Destination *Location
}

// Exits is a sorted list of exits for a location.
type Exits []Exit

// Len implements sort.Interface.Len.
func (e Exits) Len() int {
	return len(e)
}

// Swap implements sort.Interface.Swap.
func (e Exits) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// Less implements sort.Interface.Less.
func (e Exits) Less(i, j int) bool {
	return e[i].Id < e[j].Id
}

// Returns the room that exists in the given direction.  Returns valid == false
// if the alias is not a valid direction alias. Returns dest == nil if there's
// no exit in that direction.
func (e Exits) Find(alias string) (valid bool, dest *Location) {
	dir, found := config.FindDirection(alias)
	if !found {
		return false, nil
	}
	for _, exit := range e {
		if exit.Direction.Id == dir.Id {
			return true, exit.Destination
		}
	}
	return true, nil
}
