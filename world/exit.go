package world

import (
	"src.natemud.org/config"
)

// An exit is a direction that connects to a location
type Exit struct {
	Direction   *config.Direction
	Destination *Location
}

// Exits is a list of exits for a location
type Exits []Exit

func (self Exits) Len() int {
	return len(self)
}

func (self Exits) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type ExitsById struct{ Exits }

func (self ExitsById) Less(i, j int) bool {
	return self.Exits[i].Direction.Id() < self.Exits[j].Direction.Id()
}

// Returns the room that exists in the given direction.  Returns valid == false if the alias is 
// not a valid direction alias. Returns dest == nil if there's no exit in that direction
func (self Exits) Find(alias string) (valid bool, dest *Location) {
	dir := config.FindDirection(alias)
	if dir == nil {
		return false, nil
	}
	for _, exit := range self {
		if exit.Direction.Id() == dir.Id() {
			return true, exit.Destination
		}
	}
	return true, nil
}
