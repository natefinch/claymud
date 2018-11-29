package world

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/natefinch/claymud/util"
)

var (
	locMap = map[util.ID]*Location{}
	start  *Location
)

// SetStart sets the starting room of the mud.
func SetStart(room util.ID) error {
	loc, exists := locMap[room]
	if !exists {
		return fmt.Errorf("starting room %v does not exist", room)
	}
	start = loc
	return nil
}

// Start returns the start room of the MUD, where players appear
// TODO: multiple / configurable start rooms
func Start() *Location {
	return start
}

var locTemplate *template.Template

// A Location in the mud, such as a room
type Location struct {
	ID   util.ID
	Name string
	Desc string
	Exits
	Area         *Area
	Players      map[string]*Player
	Descriptions map[string]string

	// LocalActions is a map of command phrases to script names that get run in a zone-local thread.
	Actions map[string]Action
}

// returns a string representation of this location (primarily for logging)
func (l *Location) String() string {
	return fmt.Sprintf("%v [%v]", l.Name, l.ID)
}

// LocalTo returns true if the other location uses the same Worker as this
// location.
func (l *Location) LocalTo(other *Location) bool {
	return l.Area.Zone.Worker == other.Area.Zone.Worker
}

// AddPlayer syncs the player's location with the location's player list.
// TODO: Handle enter/exit notifications for others in the room
func (l *Location) AddPlayer(p *Player) {
	l.Players[strings.ToLower(p.Name())] = p
}

// Handle handles a zone-local event.
func (l *Location) Handle(event func()) {
	l.Area.Zone.Handle(event)
}

// RemovePlayer removes a player from this room.
func (l *Location) RemovePlayer(p *Player) {
	delete(l.Players, strings.ToLower(p.Name()))
}

// Target returns a target from the room with the given name or nil if none.
func (l *Location) Target(target string) *Player {
	return l.Players[target]
}

// LookTarget returns the description of the target in the room, and if a target was
// found.
func (l *Location) LookTarget(target string) (string, bool) {
	p, ok := l.Players[target]
	if ok {
		return p.Desc, true
	}
	desc, ok := l.Descriptions[target]
	if ok {
		return desc, true
	}
	return "", false
}

// ShowRoom displays the room description from the point of view of the given
// actor.
func (l *Location) ShowRoom(actor *Player) {
	locTemplate.Execute(actor, locData{actor, l})
}

func loadLocTempl(datadir string) error {
	path := filepath.Join(datadir, "location.template")
	log.Printf("Loading location template from %s", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading location template file: %s", err)
	}

	locTemplate, err = template.New("location.template").Parse(string(b))
	if err != nil {
		return fmt.Errorf("can't parse location template: %s", err)
	}
	return nil
}

type locData struct {
	Actor *Player
	*Location
}
