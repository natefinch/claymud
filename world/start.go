package world

import (
	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/util"
)

var (
	locMap = make(map[util.Id]*Location)
	start  *Location
)

// The start room of the MUD, where players appear
// TODO: multiple / configurable start rooms
func Start() *Location {
	return start
}

func init() {
	log.Printf("Generating world locations")

	// generate some rooms so we have somewhere to walk around
	// TODO: actually implement creation of rooms in-game
	size := 10
	locs := make([][]*Location, size)

	for x := 0; x < size; x++ {
		locs[x] = make([]*Location, size)
	}

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {

			l := NewLocation(
				fmt.Sprintf("Room X%d Y%d", x, y),
				fmt.Sprintf("This is the description of room Room X%d Y%d", x, y))
			locMap[l.Id()] = l
			locs[x][y] = l
		}
	}
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			loc := locs[x][y]
			exits := make([]Exit, 0)

			if x > 0 {
				exits = append(exits, Exit{config.FindDirection("w"), locs[x-1][y]})
			}
			if x < (size - 1) {
				exits = append(exits, Exit{config.FindDirection("e"), locs[x+1][y]})
			}
			if y > 0 {
				exits = append(exits, Exit{config.FindDirection("n"), locs[x][y-1]})
			}
			if y < (size - 1) {
				exits = append(exits, Exit{config.FindDirection("s"), locs[x][y+1]})
			}
			loc.SetExits(exits)
		}
	}
	start = locs[0][0]
}
