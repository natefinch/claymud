package world

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"

	"github.com/natefinch/claymud/game/social"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var (
	// ErrTimeout is returned when a player times out.
	ErrTimeout = errors.New("Player timed out")
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

var playerMap = map[string]*Player{}
var playerList = &sortedPlayers{}

type sortedPlayers []*Player

func (s *sortedPlayers) add(p *Player) {
	*s = append(*s, p)
	sort.Slice(*s, func(i, j int) bool { return (*s)[i].Name() < (*s)[j].Name() })
}
func (s *sortedPlayers) remove(p *Player) {
	for i, pl := range *s {
		if pl.Is(p) {
			*s = append((*s)[:i], (*s)[i+1:]...)
			break
		}
	}
}

// addPlayer adds a new player to the world list.
func addPlayer(p *Player) {
	playerMap[p.Name()] = p
	playerList.add(p)
}

// removePlayer removes a player from the world list.
func removePlayer(p *Player) {
	delete(playerMap, p.Name())
	playerList.remove(p)
}

// FindPlayer returns the player for the given name.
func FindPlayer(name string) (*Player, bool) {
	p, ok := playerMap[name]
	return p, ok
}

// A User is an account with a login.
type User struct {
	IP       net.Addr
	Username string
	writer   util.SafeWriter
	rwc      io.ReadWriteCloser
	*bufio.Scanner
}

// Player represents a player-character in the world.
type Player struct {
	ID     util.Id
	name   string
	Desc   string
	loc    *Location
	gender game.Gender
	global *game.Worker
	*User
	needsLF  bool
	exiting  bool
	chatmode bool
}

// SpawnPlayer attaches the connection to a player and inserts it into the world.  This
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
		Desc:     user.Username + " is hanging out here.",
		ID:       id,
		loc:      loc,
		gender:   game.Genders[0],
		global:   global,
		User:     user,
		needsLF:  true,
		chatmode: chatMode.Default,
	}
	p.writer = util.SafeWriter{Writer: rwc, OnErr: p.exit}

	// intentionally directly call the global handler so we skip the autoprompt
	// here.
	p.global.Handle(func() {
		addPlayer(p)
	})
	p.HandleLocal(func() {
		loc.AddPlayer(p)
		others := make([]io.Writer, 0, len(loc.Players))
		for _, other := range loc.Players {
			if !p.Is(other) {
				others = append(others, other)
			}
		}
		social.DoArrival(p, io.MultiWriter(others...))
		loc.ShowRoom(p)
	})
	if err := p.readLoop(); err != nil {
		p.exit(err)
	}
}

// Printf is a helper function to write the formatted string to the player.
func (p *Player) Printf(format string, args ...interface{}) {
	p.maybeNewline()
	fmt.Fprintf(p.writer, format, args...)
}

var newline = []byte("\n")

func (p *Player) maybeNewline() {
	if p.needsLF {
		p.writer.Write(newline)
		p.needsLF = false
	}
}

// WriteString implements io.StringWriter.  It will never return an error.
func (p *Player) WriteString(s string) (int, error) {
	p.maybeNewline()
	io.WriteString(p.writer, s)
	return len(s), nil
}

// Write implements io.Writer.  It will never return an error.
func (p *Player) Write(b []byte) (int, error) {
	p.maybeNewline()
	p.writer.Write(b)
	return len(b), nil
}

// Is reports whether the other player is the same as this player.
func (p *Player) Is(other *Player) bool {
	return p.ID == other.ID
}

// Name returns the player's Name.
func (p *Player) Name() string {
	return p.name
}

// String returns a string reprentation of the player (primarily for logging)
func (p *Player) String() string {
	return fmt.Sprintf("%s [%v]", p.name, p.ID)
}

// HandleLocal runs the given event for the player on its zone-local thread.
func (p *Player) HandleLocal(event func()) {
	p.loc.Handle(func() {
		event()
		p.prompt()
	})
}

// HandleGlobal runs the given event for the player on the global thread.
func (p *Player) HandleGlobal(event func()) {
	p.global.Handle(func() {
		event()
		p.prompt()
	})
}

// Move changes the player's location and adds the player to the location's map
//
// This is the function that does the heavy lifting for moving a player from one
// room to another including keeping the user's location and the location map in
// sync.  It will run on the appropriate thread depending on if this is a local
// move or a move between zones.
func (p *Player) Move(to *Location) {
	if to.ID == p.loc.ID {
		return
	}

	if p.loc.LocalTo(to) {
		p.HandleLocal(func() {
			moveEvent(p, to)
		})
	} else {
		p.HandleGlobal(func() {
			moveEvent(p, to)
		})
	}
}

func moveEvent(p *Player, to *Location) {
	p.loc.RemovePlayer(p)
	to.AddPlayer(p)
	p.loc = to
	to.ShowRoom(p)
}

// Location returns the user's location in the world.
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
	p.exiting = true
}

// Gender returns the player's gender.
func (p *Player) Gender() game.Gender {
	return p.gender
}

// readLoop is a goroutine that just passes info from the player's input to the
// runLoop.
func (p *Player) readLoop() (err error) {
	// need this because scan can panic if you send it too much stuff
	defer func() {
		panicErr := recover()
		if panicErr == nil {
			return
		}
		if e, ok := panicErr.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("%v", panicErr)
	}()
	for p.Scan() {
		// The user entered a command, so by definition has hit enter.
		p.needsLF = false
		p.handleCmd(p.Text())
		if p.exiting {
			p.HandleGlobal(func() {
				p.loc.RemovePlayer(p)
				removePlayer(p)
			})
			p.rwc.Close()
			break
		}
	}
	return p.Err()
}

// prompt shows the player's prompt to the user.
func (p *Player) prompt() {
	// TODO: standard/custom prompts
	io.WriteString(p.writer, "\n>")
	p.needsLF = true
}

// reprompt shows the player's prompt to the user, but without the preceding
// newline. This only occurs when the user hits enter with no command.
func (p *Player) reprompt() {
	// TODO: standard/custom prompts
	io.WriteString(p.writer, ">")
	p.needsLF = true
}

// timeout times the player out of the world.
func (p *Player) timeout() {
	p.WriteString("You have timed out... good bye!")
	p.exit(ErrTimeout)
}

// handleQuit asks the user if they really want to quit, and if they say yes,
// does so.
func (p *Player) handleQuit() {
	answer, err := p.Query("Are you sure you want to quit? (y/N) ")
	if err != nil {
		return
	}
	tokens := strings.Fields(answer)
	if len(tokens) == 0 {
		return
	}
	switch tokens[0] {
	case "y", "yes":
		p.exit(nil)
	}
}

// handleCmd converts tokens from the user into a Command object, and attempts
// to handle it.  It reports whether the readloop should exit
func (p *Player) handleCmd(s string) {
	cmd := Command{Actor: p, Cmd: strings.Fields(s), Loc: p.loc}
	cmd.Handle()
}

// Query asks the player a question and receives an answer
func (p *Player) Query(q string) (answer string, err error) {
	defer func() {
		if err != nil {
			p.exit(err)
		}
	}()

	return util.Query(p.rwc, q)
}
