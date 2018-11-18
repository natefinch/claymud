package world

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"text/template"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/game/gender"
	"github.com/natefinch/claymud/util"
)

var (
	TimeoutError = errors.New("Player timed out")
)

var ids = make(chan util.Id)

func init() {
	go func() {
		var id util.Id
		for {
			ids <- id
			id++
		}
	}()
}

type User struct {
	IP       net.Addr
	Username string
	writer   util.SafeWriter
	closer   io.Closer
	rwc      io.ReadWriteCloser
	*bufio.Scanner
}

// Player represents a player-character in the world.
type Player struct {
	id      util.Id
	name    string
	Desc    string
	Actions chan func()
	loc     *Location
	gender  gender.Gender
	global  *game.Worker
	*User
}

// Attaches the connection to a player and inserts it into the world.  This
// function runs for as long as the player is in the world.
func SpawnPlayer(rwc io.ReadWriteCloser, user *User, global *game.Worker) {
	id := <-ids
	log.Printf("Spawning player %s (%v) id: %v", user.Username, user.IP, id)
	user.rwc = rwc
	user.Scanner = bufio.NewScanner(rwc)
	// TODO: Persistence
	loc := Start()
	p := &Player{
		name: user.Username,
		// TODO: make this a template
		Desc:    user.Username + " is hanging out here.",
		Actions: make(chan func()),
		id:      id,
		loc:     loc,
		gender:  gender.None,
		global:  global,
		User:    user,
	}
	p.writer = util.SafeWriter{Writer: rwc, OnErr: p.exit}
	global.Handle(func() {
		addPlayer(p)
		loc.AddPlayer(p)
		loc.ShowRoom(p)
	})
	p.readLoop()
}

// Writef is a helper function to write the formatted string to the player.
func (p *Player) Writef(format string, args ...interface{}) {
	p.Write([]byte(fmt.Sprintf(format, args...)))
}

// Execute executes the given template and writes the output to the player as a
// single locked write. If we try to run the template from outside the player's
// lock, you get multiple writes which behaves badly.
func (p *Player) Execute(t *template.Template, data interface{}) error {
	p.writer.Write([]byte("\n"))
	err := t.Execute(p.writer, data)
	p.prompt()
	return err
}

// Write implements io.Writer.  It will never return an error.
func (p *Player) Write(b []byte) (int, error) {
	p.writer.Write([]byte("\n"))
	p.writer.Write(b)

	return len(b), nil
}

// Id returns the unique Id of this Player
func (p *Player) Id() util.Id {
	return p.id
}

// Name returns the player's Name.
func (p *Player) Name() string {
	return p.name
}

// String returns a string reprentation of the player (primarily for logging)
func (p *Player) String() string {
	return fmt.Sprintf("%v [%v]", p.name, p.id)
}

// Move changes the player's location and adds the player to the location's map
//
// This is the function that does the heavy lifting for moving a player from one
// room to another including keeping the user's location and the location map in
// sync
func (p *Player) Move(to *Location) {
	if to.ID == p.loc.ID {
		return
	}

	move := func() {
		p.loc.RemovePlayer(p)
		to.AddPlayer(p)
		p.loc = to
		to.ShowRoom(p)
	}
	if p.loc.LocalTo(to) {
		p.loc.Handle(move)
	} else {
		p.global.Handle(move)
	}
}

// Location returns the user's location in the world
func (p *Player) Location() *Location {
	return p.loc
}

// exit removes the player from the world, logging the error if not nil.
func (p *Player) exit(err error) {
	if err != nil {
		log.Printf("EXIT: Removing user %v from world. Error: %v", p, err)
	} else {
		log.Printf("EXIT: Removing user %v from world.", p)
	}
	p.global.Handle(func() { removePlayer(p) })
}

// Gender returns the player's gender.
func (p *Player) Gender() gender.Gender {
	return p.gender
}

// readLoop is a goroutine that just passes info from the player's input to the
// runLoop.
func (p *Player) readLoop() {
	for p.Scan() {
		p.handleCmd(p.Text())
	}
	err := p.Err()
	if err != nil {
		log.Printf("Error reading from user %v: %v", p.Name(), err)
	}
	p.exit(err)
}

// prompt shows the player's prompt to the user.
func (p *Player) prompt() {
	// TODO: standard/custom prompts
	p.writer.Write([]byte("\n>"))
}

// timeout times the player out of the world.
func (p *Player) timeout() {
	p.Writef("You have timed out... good bye!")
	p.exit(TimeoutError)
}

// HandleCommand handles commands specific to the player, such as inventory and
// player stats.
func (p *Player) HandleCommand(cmd *Command) bool {

	// TODO: implement player commands
	switch cmd.Action() {
	case "quit":
		p.handleQuit()
		return true
	}
	return false
}

// handleQuit asks the user if they really want to quit, and if they say yes,
// does so.
func (p *Player) handleQuit() {
	answer, err := p.Query([]byte("Are you sure you want to quit? (y/n) "))
	if err != nil {
		return
	}
	tokens := tokenize(answer)
	switch tokens[0] {
	case "y", "yes":
		p.exit(nil)
	default:
		p.prompt()
	}
}

// handleCmd converts tokens from the user into a Command object, and attempts
// to handle it.
func (p *Player) handleCmd(s string) {
	cmd := NewCommand(p, tokenize(s))
	if !p.HandleCommand(cmd) {
		p.loc.HandleCommand(cmd)
	}
	p.prompt()
}

// tokenize returns a list of space separated tokens from the given string.
func tokenize(s string) []string {
	return strings.Split(strings.TrimSpace(s), " ")
}

// Query asks the player a question and receives an answer
func (p *Player) Query(q []byte) (answer string, err error) {
	defer func() {
		if err != nil {
			p.exit(err)
		}
	}()

	return util.Query(p.rwc, q)
}
