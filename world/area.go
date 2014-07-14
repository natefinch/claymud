package world

import (
	"github.com/natefinch/claymud/util"
)

// Area is a small collection of related locations, such as the rooms in a
// hotel.
type Area struct {
	id        util.Id
	name      string
	Locations []*Location
}

func (a Area) Id() util.Id {
	return a.id
}

func (a Area) Name() string {
	return a.name
}

// Zone is a collection of Areas that represent one large and logically distinct
// section of the mud, such as a town.
type Zone struct {
	id        util.Id
	name      string
	Areas     []*Area
	Locations []*Location
}
