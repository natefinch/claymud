package world

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/util"
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

func genWorld(zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) {
	log.Printf("Generating world locations")

	zone := SpawnZone("The Town of Momentary", zoneLock, shutdown, wg)

	momentary := &Area{
		ID:   <-ids,
		Name: "Downtown",
	}
	zone.Add(momentary)

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
				fmt.Sprintf("This is the description of room Room X%d Y%d", x, y),
				momentary,
			)
			locMap[l.ID] = l
			locs[x][y] = l
		}
	}
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			loc := locs[x][y]
			exits := make([]Exit, 0)

			if x > 0 {
				exits = append(exits, Exit{dir("w"), locs[x-1][y]})
			}
			if x < (size - 1) {
				exits = append(exits, Exit{dir("e"), locs[x+1][y]})
			}
			if y > 0 {
				exits = append(exits, Exit{dir("n"), locs[x][y-1]})
			}
			if y < (size - 1) {
				exits = append(exits, Exit{dir("s"), locs[x][y+1]})
			}
			e := Exits(exits)
			sort.Sort(e)
			loc.Exits = e
		}
	}
	start = locs[0][0]
}

func dir(d string) config.Direction {
	dir, ok := config.FindDirection(d)
	if !ok {
		panic(fmt.Errorf("Can't find direction %s", d))
	}
	return dir
}
