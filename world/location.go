package world

import (
	"fmt"
	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/util"
	"log"
	"sort"
	"strings"
)

var (
	nextLocId = util.IdGenerator()
	locMap    = make(map[util.Id]*Location)
	start     *Location
)

// The start room of the MUD, where players appear
// TODO: multiple / configurable start rooms
func Start() *Location {
	return start
}

// A location in the mud, such as a room
type Location struct {
	name    string
	desc    string
	exits   Exits
	id      util.Id
	Players map[util.Id]*Player
	pNames  map[string]*Player
	Add     chan *Player
	Remove  chan *Player
	Cmd     chan *Command
}

// Creates a new location and starts its run loop
func NewLocation(name string, desc string) (loc *Location) {
	// TODO: fix chicken an egg problem with two rooms that need to be created with exits
	// that point to each other
	loc = &Location{
		name,
		desc,
		Exits(make([]Exit, 0)),
		<-nextLocId,
		make(map[util.Id]*Player),
		make(map[string]*Player),
		make(chan *Player),
		make(chan *Player),
		make(chan *Command)}
	go loc.runLoop()
	return
}

// Returns the unique Id of this location
func (l *Location) Id() util.Id {
	return l.id
}

// Returns the user visible name of this location
func (l *Location) Name() string {
	return l.name
}

// Sets the user visible name of this location
func (l *Location) SetName(name string) {
	l.name = name
}

// Returns the location's description that you see when you look in the room
func (l *Location) Desc() string {
	return l.desc
}

// Sets the location's description
func (l *Location) SetDesc(desc string) {
	l.desc = desc
}

// Exits returns a copy of the exits this locations contains
func (l *Location) Exits() Exits {
	return Exits(l.exits)
}

// SetExits changes the exits for this Location
func (l *Location) SetExits(exits Exits) {
	l.exits = exits
	sort.Sort(ExitsById{l.exits})
}

// returns a string representation of this location (primarily for logging)
func (l *Location) String() string {
	return fmt.Sprintf("%v [%v]", l.name, l.id)
}

// loop that controls access to the location
func (l *Location) runLoop() {
	for {
		select {
		case p := <-l.Add:
			l.AddPlayer(p)
		case p := <-l.Remove:
			l.RemovePlayer(p)
		case cmd := <-l.Cmd:
			if !l.handleCommand(cmd) {
				cmd.HandleAt(l)
			}
		}
	}
}

// TODO: Handle enter/exit notifications for others in the room
func (l *Location) AddPlayer(p *Player) {
	l.Players[p.Id()] = p
	l.pNames[strings.ToLower(p.Name())] = p
}

func (l *Location) RemovePlayer(p *Player) {
	delete(l.Players, p.Id())
	delete(l.pNames, strings.ToLower(p.Name()))
}

func (l *Location) handleCommand(cmd *Command) bool {
	// TODO: implement plugins to handle custom commands
	return false
}

// Determine the target in the room from the command's target
// returns nil if no target exists
func (l *Location) Target(cmd *Command) (p *Player) {
	// TODO: support aliases
	return l.pNames[cmd.Target()]
}

// creates the room description from the point of view of the given actor
func (l *Location) RoomDesc(actor *Player) string {
	// construct the output for describing the room
	lines := make([]string, 0)

	lines = append(lines, l.name)
	lines = append(lines, "")
	lines = append(lines, l.desc)
	lines = append(lines, "")

	if len(l.exits) == 0 {
		lines = append(lines, "There are no exits!")
	} else {
		lines = append(lines, "[Exits]")

		// the exits are sorted, so this always prints out in the same order
		for _, exit := range l.exits {
			lines = append(lines, fmt.Sprintf("%v - %v", exit.Direction.Name(), exit.Destination.Name()))
		}
	}

	first := true
	for _, p := range l.Players {
		if p.Id() != actor.Id() {
			if first {
				lines = append(lines, "")
				first = false
			}
			lines = append(lines, p.Desc())
		}
	}

	// TODO: implement showing items on the ground

	return strings.Join(lines, "\r\n")
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
