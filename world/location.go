package world

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/util"
)

var locTemplate *template.Template

// A location in the mud, such as a room
type Location struct {
	id   util.Id
	Name string
	Desc string
	Exits
	Area          *Area
	Zone          *Zone
	Players       map[util.Id]*Player
	PlayersByName map[string]*Player
	sync.RWMutex
}

// Creates a new location and starts its run loop
func NewLocation(name string, desc string) *Location {
	// TODO: fix chicken an egg problem with two rooms that need to be created with exits
	// that point to each other
	return &Location{
		Name:          name,
		Desc:          desc,
		id:            1,
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
	l.ShowRoom(p)
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

// ShowRoom displays the room description from the point of view of the given
// actor.
func (l *Location) ShowRoom(actor *Player) {
	actor.Execute(locTemplate, locData{actor, l})
}

func loadLocTempl() error {
	path := filepath.Join(config.DataDir(), "location.template")
	log.Printf("Loading location template from %s", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading location template file: %s", err)
	}

	locTemplate, err = template.New("location.template").Parse(preprocess(b))
	if err != nil {
		return fmt.Errorf("can't parse location template: %s", err)
	}
	return nil
}

type locData struct {
	Actor *Player
	*Location
}

var strip = regexp.MustCompile("}}\n")
var repl = []byte("}}")

func preprocess(b []byte) string {
	return string(strip.ReplaceAll(b, repl))
}
