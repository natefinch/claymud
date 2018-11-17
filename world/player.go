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
	"text/template"

	"github.com/natefinch/claymud/game/gender"
	"github.com/natefinch/claymud/lock"
	"github.com/natefinch/claymud/util"
)

var (
	TimeoutError = errors.New("Player timed out")
)

type User struct {
	IP       net.Addr
	Username string
	writer   util.SafeWriter
	closer   io.Closer
	rwc      io.ReadWriteCloser
	*bufio.Scanner
	sync.RWMutex
}

// Player is a struct representing a user in the world
type Player struct {
	id     util.Id
	name   string
	Desc   string
	loc    *Location
	gender gender.Gender

	*User
}

// Attaches the connection to a player and inserts it into the world
func SpawnPlayer(rwc io.ReadWriteCloser, user *User) {
	log.Printf("Spawning player %s (%v)", user.Username, user.IP)
	user.rwc = rwc
	user.Scanner = bufio.NewScanner(rwc)
	// TODO: Persistence
	loc := Start()
	p := &Player{
		name: user.Username,
		// TODO: make this a template
		Desc:   fmt.Sprintf("You see %v here.", user.Username),
		id:     0,
		loc:    loc,
		gender: gender.None,

		User: user,
	}
	p.writer = util.SafeWriter{Writer: rwc, OnErr: p.exit}
	AddPlayer(p)
	go p.readLoop()
	loc.AddPlayer(p)
	p.prompt()
}

// Writef is a helper function to write the formatted string to the player.
func (p *Player) Writef(format string, args ...interface{}) {
	p.Write([]byte(fmt.Sprintf(format, args...)))
}

// Execute executes the given template and writes the output to the player as a
// single locked write. If we try to run the template from outside the player's
// lock, you get multiple writes which behaves badly.
func (p *Player) Execute(t *template.Template, data interface{}) {
	p.writer.Write([]byte("\n"))
	if err := t.Execute(p.writer, data); err != nil {
		log.Printf("ERROR: problem writing template to user: %s", err)
	}
}

// Write implements io.Writer.  It will never return an error.
func (p *Player) Write(b []byte) (int, error) {
	p.Lock()
	defer p.Unlock()

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
func (p *Player) Move(loc *Location) {
	if loc.Id() != p.loc.Id() {
		locks := []lock.IdLocker{p, loc, p.loc}
		lock.All(locks)
		defer lock.UnlockAll(locks)

		p.loc.RemovePlayer(p)
		p.loc = loc
		loc.AddPlayer(p)
	}
}

// Location returns the user's location in the world
func (p *Player) Location() *Location {
	return p.loc
}

// exit removes the player from the world, logging the error if not nil.
func (p *Player) exit(err error) {
	p.Lock()
	defer p.Unlock()
	if err != nil {
		log.Printf("EXIT: Removing user %v from world. Error: %v", p, err)
	} else {
		log.Printf("EXIT: Removing user %v from world.", p)
	}
	RemovePlayer(p)
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
		RemovePlayer(p)
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
	p.Lock()
	defer p.Unlock()
	defer func() {
		if err != nil {
			p.exit(err)
		}
	}()

	return util.Query(p.rwc, q)
}
