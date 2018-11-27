package world

import (
	"fmt"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

// Area is a small collection of related locations, such as the rooms in a
// hotel.
type Area struct {
	ID        util.ID
	Name      string
	Zone      *Zone
	Locations []*Location
	LocByID   map[util.ID]*Location
}

// Add adds the location to this area.
func (a *Area) Add(l *Location) {
	l.Area = a
	a.Locations = append(a.Locations, l)
	a.LocByID[l.ID] = l
}

// Zone is a collection of Areas that represent one large and logically distinct
// section of the mud, such as a town.
type Zone struct {
	ID     util.ID
	Name   string
	Closed bool
	Areas  []*Area
	*game.Worker
}

func (z *Zone) String() string {
	if z == nil {
		return "<Zone nil>"
	}
	return fmt.Sprintf("<Zone %q (%v)>", z.Name, z.ID)
}

// Add adds the area to this zone.
func (z *Zone) Add(a *Area) {
	z.Areas = append(z.Areas, a)
	a.Zone = z
}
