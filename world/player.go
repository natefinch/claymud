package world

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gofrs/uuid"

	"github.com/natefinch/claymud/auth"
	"github.com/natefinch/claymud/db"
	"github.com/natefinch/claymud/game/social"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var (
	// ErrTimeout is returned when a player times out.
	ErrTimeout = errors.New("Player timed out")
)

// PFlag represents a flag (bit) set on a player.
type PFlag int

// All possible player flags.
//
// DO NOT REARRANGE OR COMMENT OUT VALUES.  If you need to deprecate a value,
// append _DEPRECATED on the end of the name.  New values must be appended to
// this list, never inserted anywhere else.
const (
	PFlagChatmode PFlag = iota
)

var (
	playerMap  = map[string]*Player{}
	playerList = &sortedPlayers{}
	userMap    = map[string]*auth.User{}
)

type sortedPlayers []*Player

func (s *sortedPlayers) add(p *Player) {
	*s = append(*s, p)
	sort.SliceStable(*s, func(i, j int) bool { return (*s)[i].Name() < (*s)[j].Name() })
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

// FindUser returns the user for the given username.
func FindUser(name string) (*auth.User, bool) {
	p, ok := userMap[name]
	return p, ok
}

// Player represents a player-character in the world.
type Player struct {
	ID     util.ID
	uuid   uuid.UUID
	name   string
	Desc   string
	loc    *Location
	gender game.Gender
	global *game.Worker
	*auth.User
	util.SafeWriter
	bits    *big.Int
	needsLF bool
	exiting bool
}

// SpawnPlayer attaches the connection to a player and inserts it into the world.  This
// function runs for as long as the player is in the world.
func SpawnPlayer(st *db.Store, user *auth.User, global *game.Worker) error {
	dbp, err := chooseDBPlayer(st, user)
	if err != nil {
		return err
	}

	log.Printf("Spawning user %s's player %s with id: %v", user.Username, dbp.Name, dbp.ID)

	loc := Start()
	p := &Player{
		name:    dbp.Name,
		Desc:    dbp.Description,
		ID:      dbp.ID,
		loc:     loc,
		gender:  dbp.Gender,
		global:  global,
		User:    user,
		needsLF: true,
		bits:    dbp.Flags,
	}
	p.SafeWriter = util.SafeWriter{Writer: user, OnErr: p.exit}

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
	return nil
}

func chooseDBPlayer(st *db.Store, user *auth.User) (*db.Player, error) {
	if len(user.Players) == 0 {
		_, err := io.WriteString(user, "You have no players, let's create one.\n")
		if err != nil {
			return nil, err
		}
		return createPlayer(st, user, game.Genders)
	}
	if user.Flag(auth.UFlagAdmin) {
		// Admins get exactly one player.
		return st.FindPlayer(user.Players[0])
	}
	newChar := "(create new)"
	choices := append([]string{newChar}, user.Players...)
	i, err := util.QueryStrings(user, "Choose a player, or c to create a new one:\n", -1, choices...)
	if err != nil {
		return nil, err
	}
	if i == 0 {
		return createPlayer(st, user, game.Genders)
	}
	return st.FindPlayer(choices[i])
}

func verifyName(name string) (string, error) {
	if name == "" {
		return "!! Names cannot be empty.", nil
	}
	if utf8.RuneCount([]byte(name)) > 32 {
		return "!! Names cannot be longer than 32 characters.", nil
	}
	if utf8.RuneCount([]byte(name)) < 3 {
		return "!! Names must be at least 3 characters.", nil
	}
	for _, r := range []rune(name) {
		if !unicode.IsLetter(r) {
			return "!! Names may only contain letters.", nil
		}
	}
	return "", nil
}

func createPlayer(st *db.Store, user *auth.User, genders []game.Gender) (*db.Player, error) {
	const queryName = "By what name do you wish your character to be known? "
	name, err := util.QueryVerify(user, queryName, verifyName)
	if err != nil {
		return nil, err
	}
	options := make([]string, len(genders))
	for i := range genders {
		options[i] = fmt.Sprintf("%s (%s/%s)", genders[i].Name, genders[i].Xe, genders[i].Xim)
	}
	i, err := util.QueryStrings(user, "What should this character's gender be?\n", -1, options...)
	if err != nil {
		return nil, err
	}
	gender := genders[i]
	p := &db.Player{
		Name:        name,
		Description: name + " is standing here.",
		Gender:      gender,
		Flags:       big.NewInt(0),
	}
	for {
		err = st.CreatePlayer(user.Username, p)
		if _, ok := err.(db.ErrExists); ok {
			_, err := io.WriteString(user, "A character with that name already exists.\n")
			if err != nil {
				return nil, err
			}

			name, err := util.QueryVerify(user, queryName, verifyName)
			if err != nil {
				return nil, err
			}
			p.Name = name
			continue
		}
		if err != nil {
			return nil, err
		}
		return p, nil
	}
}

// Printf is a helper function to write the formatted string to the player.
func (p *Player) Printf(format string, args ...interface{}) {
	p.maybeNewline()
	fmt.Fprintf(p.Writer, format, args...)
}

var newline = []byte("\n")

func (p *Player) maybeNewline() {
	if p.needsLF {
		p.Writer.Write(newline)
		p.needsLF = false
	}
}

// WriteString implements io.StringWriter.  It will never return an error.
func (p *Player) WriteString(s string) (int, error) {
	p.maybeNewline()
	io.WriteString(p.Writer, s)
	return len(s), nil
}

// Write implements io.Writer.  It will never return an error.
func (p *Player) Write(b []byte) (int, error) {
	p.maybeNewline()
	p.Writer.Write(b)
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
			p.Relocate(to)
		})
	} else {
		p.HandleGlobal(func() {
			p.Relocate(to)
		})
	}
}

// Relocate moves the character to a new lcoation. This is NOT run in a worker,
// so you need to handle that yourself.
func (p *Player) Relocate(to *Location) {
	p.loc.RemovePlayer(p)
	to.AddPlayer(p)
	p.loc = to
	to.ShowRoom(p)
}

// Flag reports if the given flag has been set to true for the user.
func (p *Player) Flag(f PFlag) bool {
	return p.bits.Bit(int(f)) == 1
}

// SetFlag sets the given flag to true for the user
func (p *Player) SetFlag(f PFlag) {
	p.bits.SetBit(p.bits, int(f), 1)
}

// UnsetFlag sets the given flag to false for the user
func (p *Player) UnsetFlag(f PFlag) {
	p.bits.SetBit(p.bits, int(f), 0)
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
	// defer func() {
	// 	panicErr := recover()
	// 	if panicErr == nil {
	// 		return
	// 	}
	// 	if e, ok := panicErr.(error); ok {
	// 		err = e
	// 		return
	// 	}
	// 	err = fmt.Errorf("%v", panicErr)
	// }()
	for p.Scan() {
		// The user entered a command, so by definition has hit enter.
		p.needsLF = false
		p.handleCmd(p.Text())
		if p.exiting {
			p.HandleGlobal(func() {
				p.loc.RemovePlayer(p)
				removePlayer(p)
			})
			p.Close()
			break
		}
	}
	return p.Err()
}

// prompt shows the player's prompt to the user.
func (p *Player) prompt() {
	// TODO: standard/custom prompts
	io.WriteString(p.Writer, "\n>")
	p.needsLF = true
}

// reprompt shows the player's prompt to the user, but without the preceding
// newline. This only occurs when the user hits enter with no command.
func (p *Player) reprompt() {
	// TODO: standard/custom prompts
	io.WriteString(p.Writer, ">")
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

	return util.Query(p, q)
}
