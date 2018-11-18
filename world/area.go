package world

import (
	"sync"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

// Area is a small collection of related locations, such as the rooms in a
// hotel.
type Area struct {
	ID        util.Id
	Name      string
	Zone      *Zone
	Locations []*Location
}

// Add adds the location to this area.
func (a *Area) Add(l *Location) {
	l.Area = a
	a.Locations = append(a.Locations, l)
}

// SpawnZone creates a new Zone and spawns a worker goroutine for which handles
// events in that zone.
func SpawnZone(name string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) *Zone {
	w := game.SpawnWorker(zoneLock, shutdown, wg)
	return &Zone{
		ID:     <-ids,
		Name:   name,
		Worker: w,
	}
}

// Zone is a collection of Areas that represent one large and logically distinct
// section of the mud, such as a town.
type Zone struct {
	ID    util.Id
	Name  string
	Areas []*Area
	*game.Worker
}

// Add adds the area to this zone.
func (z *Zone) Add(a *Area) {
	z.Areas = append(z.Areas, a)
	a.Zone = z
}
