package game

import (
	"fmt"
	"strings"
)

var (
	dirMap  = map[string]Direction{}
	dirList []Direction
)

// InitDirs initializes the configured directions.  Order matters, the order becomes
// how they'll get displayed in the list of exits.
func InitDirs(dirs []Direction) {
	if len(dirs) == 0 {
		panic(fmt.Errorf("no directions defined in config"))
	}
	for i := range dirs {
		dirs[i].ID = i
	}
	// make a map of names and aliases to the directions
	// the names and aliases will become commands to move in the world
	for _, dir := range dirs {
		addDir(dir)
	}
	dirList = dirs
}

// Direction is a direction that a standard exit can use
type Direction struct {
	ID int
	// Name is what will be shown in room descriptions.
	Name string
	// From is used when displaying enter/exit notifications for players.
	From string
	// Aliases is a list of alternate names for the direction.
	Aliases []string
}

// FindDir will find a direction by name or alias.  This method is not case
// sensitive.
func FindDirection(alias string) (dir Direction, found bool) {
	dir, found = dirMap[strings.ToLower(alias)]
	return dir, found
}

// AllDirections returns a list of all the directions that exist in the world.
func AllDirections() []Direction {
	return dirList
}

// addDir adds the given direction to the global map of directions
//
// This function uses the name and aliases as ids to find the direction
func addDir(dir Direction) {
	name := strings.ToLower(dir.Name)
	if _, ok := dirMap[name]; ok {
		panic(fmt.Errorf("direction with name/alias %q already exists", name))
	}
	dirMap[name] = dir
	for _, alias := range dir.Aliases {
		alias = strings.ToLower(alias)
		if _, ok := dirMap[alias]; ok {
			panic(fmt.Errorf("direction with name/alias %q already exists", alias))
		}
		dirMap[alias] = dir
	}
}
