package world

import (
	"github.com/natefinch/natemud/config"
)

// An exit is a direction that connects to a location
type Exit struct {
	Direction   *config.Direction
	Destination *Location
}

// Exits is a list of exits for a location
type Exits []Exit

func (e Exits) Len() int {
	return len(e)
}

func (e Exits) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type ExitsById struct{ Exits }

func (e ExitsById) Less(i, j int) bool {
	return e.Exits[i].Direction.Id() < e.Exits[j].Direction.Id()
}

// Returns the room that exists in the given direction.  Returns valid == false if the alias is
// not a valid direction alias. Returns dest == nil if there's no exit in that direction
func (e Exits) Find(alias string) (valid bool, dest *Location) {
	dir := config.FindDirection(alias)
	if dir == nil {
		return false, nil
	}
	for _, exit := range e {
		if exit.Direction.Id() == dir.Id() {
			return true, exit.Destination
		}
	}
	return true, nil
}
