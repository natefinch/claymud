package world

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/lock"
	"github.com/natefinch/natemud/util"
)

var (
	ErrTimeout = errors.New("Player timed out")
)

// Player is a struct representing a user in the world
type Player struct {
	Name string
	Desc string
	IP   net.Addr
	Id   util.Id
	Loc  *Location
	Sex  config.Sex

	writer io.Writer
	closer io.Closer
	*bufio.Scanner

	*sync.RWMutex
}

// Attaches the connection to a player and inserts it into the world
func SpawnPlayer(rwc io.ReadWriteCloser, name string, ip net.Addr) {
	// TODO: Persistence
	loc := Start()
	p := &Player{
		Name: name,
		Desc: fmt.Sprintf("You see %v here.", name),
		IP:   ip,
		Id:   <-util.Ids,
		Loc:  loc,
		Sex:  config.SEX_NONE,

		closer:  rwc,
		writer:  rwc,
		Scanner: bufio.NewScanner(rwc),
	}
	AddPlayer(p)
	go p.runLoop()
	go p.readLoop()
	loc.AddPlayer(p)
}

// Write writes out text to the user.
func (p *Player) Write(msg string, args ...string) {
	if _, err := p.Writer.Write("\n"); err != nil {
		p.exit(err)
	}

	if _, err := p.Writer.Write([]byte(fmt.Sprintf(msg, args...))); err != nil {
		p.exit(err)
	}
	p.prompt()
}

// Id returns the unique Id of this Player
func (p *Player) Id() util.Id {
	return p.Id
}

// String returns a string reprentation of the player (primarily for logging)
func (p *Player) String() string {
	return fmt.Sprintf("%v [%v]", p.Name, p.Id)
}

// Move changes the player's location and adds the player to the location's map
//
// This is the function that does the heavy lifting for moving a player from one
// room to another including keeping the user's location and the location map in
// sync
func (p *Player) Move(loc *Location) {
	if loc != p.Loc {
		locks := []IdLocker{p, loc, p.Loc}
		lock.All(locks)

		p.loc.Remove(p)
		loc.Add(p)
		p.Loc = loc

		lock.UnlockAll(locks)

		p.handleCmd([]string{"look"})
	}
}

// Location returns the user's location in the world
func (p *Player) Location() *Location {
	return p.loc
}

func (p *Player) exit(err error) {
	p.Lock()
	if err != nil {
		log.Printf("Removing user %v from world. Error: %v", p.name, err)
	}
	RemovePlayer(p)
	p.Unlock()
}

// Title implements config.Person.Title()
func (p *Player) Title() string {
	return p.Name
}

// Gender implements config.Person.Gender()
func (p *Player) Gender() config.Sex {
	return p.Sex
}

// readLoop is a goroutine that just passes info from the player's input to the
// runLoop
func (p *Player) readLoop() {
	for p.Scan() {
		p.handleCmd(p.Text())
	}
	err := p.in.Err()
	if err != nil {
		log.Printf("Error reading from user %v: %v", p.Name, err)
	}
	p.exit(err)
}

// promp shows the user's prompt to the user
func (p *Player) prompt() {
	// TODO: standard/custom prompts
	if _, err := p.out.Write(">"); err != nil {
		log.Print("Failed writing prompt to %s: %s", p, err)
		p.exit(err)
	}
}

// timeout times the player out of the world
func (p *Player) timeout() {
	util.WriteLn(p.out, "")
	util.WriteLn(p.out, "You have timed out... good bye!")
	p.exit <- ErrTimeout
}

// HandleCommand handles commands specific to the player, such as inventory and
// player stats
func (p *Player) HandleCommand(cmd *Command) bool {

	// TODO: implement player commands
	switch cmd.Action() {
	case "quit":
		p.handleQuit()
		return true
	}
	return false
}

// handleQuit asks the user if they really want to quit, and if they do, does so
func (p *Player) handleQuit() {
	_, err := p.Write("Are you sure you want to quit? (y/n)")

	if p.Scan() {
		tokens := tokenize(p.Text())
		switch tokens[0] {
		case "y", "yes":
			RemovePlayer(p)
		default:
			p.prompt()
		}
	}
	if err := p.Err(); ert != nil {
		p.exit(err)
	}
}

// handleCmd converts tokens from the user into a Command object, and attempts
// to handle it
func (p *Player) handleCmd(s string) {
	cmd := NewCommand(p, tokenize(s))
	if !p.HandleCommand(cmd) {
		p.loc.HandleCommand(cmd)
	}
}

func tokenize(s string) []string {
	return strings.Split(strings.TrimSpace(s), " ")
}
