package world

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/natefinch/natemud/util"
)

// A location in the mud, such as a room
type Location struct {
	id   util.Id
	Name string
	Desc string
	Exits
	Players       map[util.Id]*Player
	PlayersByName map[string]*Player
	*sync.RWMutex
}

// Creates a new location and starts its run loop
func NewLocation(name string, desc string) *Location {
	// TODO: fix chicken an egg problem with two rooms that need to be created with exits
	// that point to each other
	return &Location{
		Name:          name,
		Desc:          desc,
		id:            <-util.Ids,
		Players:       make(map[util.Id]*Player),
		PlayersByName: make(map[string]*Player),
	}
}

// Returns the unique Id of this location
func (l *Location) Id() util.Id {
	return l.id
}

// returns a string representation of this location (primarily for logging)
func (l *Location) String() string {
	return fmt.Sprintf("%v [%v]", l.Name, l.id)
}

func (l *Location) HandleCommand(cmd *Command) {
	if !l.handleCommand(cmd) {
		cmd.HandleAt(l)
	}
}

// TODO: Handle enter/exit notifications for others in the room
func (l *Location) AddPlayer(p *Player) {
	l.Players[p.Id()] = p
	l.PlayersByName[strings.ToLower(p.Name())] = p
}

func (l *Location) RemovePlayer(p *Player) {
	delete(l.Players, p.Id())
	delete(l.PlayersByName, strings.ToLower(p.Name()))
}

func (l *Location) handleCommand(cmd *Command) bool {
	// TODO: implement plugins to handle custom commands
	return false
}

// Determine the target in the room from the command's target
// returns nil if no target exists
func (l *Location) Target(cmd *Command) (p *Player) {
	// TODO: support aliases
	return l.PlayersByName[cmd.Target()]
}

// creates the room description from the point of view of the given actor
func (l *Location) RoomDesc(actor *Player) string {
	// construct the output for describing the room
	// TODO: Make this a template
	lines := make([]string, 0)

	lines = append(lines, l.Name)
	lines = append(lines, "")
	lines = append(lines, l.Desc)
	lines = append(lines, "")

	if len(l.Exits) == 0 {
		lines = append(lines, "There are no exits!")
	} else {
		lines = append(lines, "[Exits]")

		// the exits are sorted, so this always prints out in the same order
		sort.Sort(l.Exits)
		for _, exit := range l.Exits {
			lines = append(lines, fmt.Sprintf("%v - %v", exit.Name, exit.Destination.Name))
		}
	}

	first := true
	for _, p := range l.Players {
		if p.Id() != actor.Id() {
			if first {
				lines = append(lines, "")
				first = false
			}
			lines = append(lines, p.Desc)
		}
	}

	// TODO: implement showing items on the ground

	return strings.Join(lines, "\r\n")
}
