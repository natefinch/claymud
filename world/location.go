package world

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/util"
)

var locTemplate *template.Template

// A location in the mud, such as a room
type Location struct {
	ID   util.Id
	Name string
	Desc string
	Exits
	Area    *Area
	Players map[string]*Player
}

// NewLocation creates a new location.
func NewLocation(name string, desc string, area *Area) *Location {
	l := &Location{
		Name:    name,
		Desc:    desc,
		ID:      <-ids,
		Players: make(map[string]*Player),
	}
	area.Add(l)
	return l
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

// Determine the target in the room from the command's target
// returns nil if no target exists
func (l *Location) Target(cmd *Command) (p *Player) {
	// TODO: support aliases
	return l.Players[cmd.Target()]
}

// ShowRoom displays the room description from the point of view of the given
// actor.
func (l *Location) ShowRoom(actor *Player) {
	locTemplate.Execute(actor, locData{actor, l})
}

func loadLocTempl() error {
	path := filepath.Join(config.DataDir(), "location.template")
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
