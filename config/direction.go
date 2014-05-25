package config

import (
	"github.com/natefinch/natemud/util"
	"strings"
)

var (
	dirMap    map[string]*Direction
	dirs      []*Direction
	nextDirId = util.IdGenerator()
)

// Direction is a direction that a standard exit can use
type Direction struct {
	id      util.Id
	name    string
	from    string
	aliases []string
}

// NewDirection makes a new Direction struct
//
// The direction name and aliases will become commands in the world to move around.
// Be careful what aliases you give as they will override any other commands in the world
func NewDirection(name, from string, aliases []string) *Direction {
	return &Direction{<-nextDirId, name, from, aliases}
}

// Name is the full name of the direction, which will be shown in room descriptions
func (d *Direction) Name() string {
	return d.name
}

// From is the string that is used when displaying enter/exit notifications for players
func (d *Direction) From() string {
	return d.from
}

// Aliases is a list of alternate commands that can be used to move through this direction
func (d *Direction) Aliases() []string {
	return d.aliases
}

func (d *Direction) Id() util.Id {
	return d.id
}

// FindDir will find a direction by name or alias.  This method is not case sensitives
func FindDirection(alias string) *Direction {
	return dirMap[strings.ToLower(alias)]
}

// AllExits returns a list of all the exit types that exist in the world
func AllDirections() []*Direction {
	return dirs
}

// addDir adds the given direction to the global map of directions
//
// This function uses the name and aliases as ids to find the direction
func addDir(dir *Direction) {
	dirMap[strings.ToLower(dir.Name())] = dir
	for _, alias := range dir.Aliases() {
		dirMap[strings.ToLower(alias)] = dir
	}
}

// init loads all the directions that exist in the world
func init() {

	// TODO: move this to a config file
	// note: order here will be the order in which they are displayed in room descriptions

	n := NewDirection("North", "the North", []string{"n"})
	s := NewDirection("South", "the South", []string{"s"})
	e := NewDirection("East", "the East", []string{"e"})
	w := NewDirection("West", "the West", []string{"w"})
	ne := NewDirection("Northeast", "the Northeast", []string{"ne"})
	nw := NewDirection("Northwest", "the Northwest", []string{"nw"})
	se := NewDirection("Southeast", "the Southeast", []string{"se"})
	sw := NewDirection("Southwest", "the Southwest", []string{"sw"})
	u := NewDirection("Up", "above", []string{"u"})
	d := NewDirection("Down", "below", []string{"d"})

	dirs = []*Direction{n, ne, e, se, s, sw, w, nw, u, d}

	dirMap = make(map[string]*Direction)

	// make a map of names and aliases to the directions
	// the names and aliases will become commands to move in the world
	for _, dir := range dirs {
		addDir(dir)
	}
}
