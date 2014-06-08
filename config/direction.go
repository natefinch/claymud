package config

import (
	"strings"
)

var (
	dirMap map[string]Direction
	dirs   []Direction
)

// Direction is a direction that a standard exit can use
type Direction struct {
	// Order determines the sort order for the exits in room descriptions.
	Order int
	// Name is what will be shown in room descriptions
	Name string
	// From what is used when displaying enter/exit notifications for players
	From string
	// Aliases is a list of alternate names for the direction
	Aliases []string
}

// FindDir will find a direction by name or alias.  This method is not case
// sensitive.
func FindDirection(alias string) Direction {
	return dirMap[strings.ToLower(alias)]
}

// AllExits returns a list of all the exit types that exist in the world
func AllDirections() []Direction {
	return dirs
}

// addDir adds the given direction to the global map of directions
//
// This function uses the name and aliases as ids to find the direction
func addDir(dir Direction) {
	dirMap[strings.ToLower(dir.Name)] = dir
	for _, alias := range dir.Aliases {
		dirMap[strings.ToLower(alias)] = dir
	}
}

// init loads all the directions that exist in the world
func init() {

	// TODO: move this to a config file

	n := Direction{0, "North", "the North", []string{"n"}}
	s := Direction{1, "South", "the South", []string{"s"}}
	e := Direction{2, "East", "the East", []string{"e"}}
	w := Direction{3, "West", "the West", []string{"w"}}
	ne := Direction{4, "Northeast", "the Northeast", []string{"ne"}}
	nw := Direction{5, "Northwest", "the Northwest", []string{"nw"}}
	se := Direction{6, "Southeast", "the Southeast", []string{"se"}}
	sw := Direction{7, "Southwest", "the Southwest", []string{"sw"}}
	u := Direction{8, "Up", "above", []string{"u"}}
	d := Direction{9, "Down", "below", []string{"d"}}

	dirs = []Direction{n, ne, e, se, s, sw, w, nw, u, d}

	dirMap = make(map[string]Direction)

	// make a map of names and aliases to the directions
	// the names and aliases will become commands to move in the world
	for _, dir := range dirs {
		addDir(dir)
	}
}
