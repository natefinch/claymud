package world

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/util"
	"io"
	"log"
	"net"
	"strings"
)

var (
	nextPlayerId = util.IdGenerator()
	ErrTimeout   = errors.New("Player timed out")
)

// Player is a struct representing a user in the world
type Player struct {
	rw    *bufio.ReadWriter
	name  string
	desc  string
	ip    net.Addr
	c     io.Closer
	id    util.Id
	loc   *Location
	Move  chan *Location
	input chan string
	exit  chan error
	cmd   chan []string
	sex   config.Sex
}

// Attaches the connection to a player and inserts it into the world
func SpawnPlayer(rwc io.ReadWriteCloser, name string, ip net.Addr) {
	loc := Start()
	player := &Player{
		bufio.NewReadWriter(bufio.NewReader(rwc), bufio.NewWriter(rwc)),
		name,
		fmt.Sprintf("You see %v here.", name),
		ip,
		rwc,
		<-nextPlayerId,
		loc,
		make(chan *Location),
		make(chan string),
		make(chan error),
		make(chan []string),
		config.SEX_NONE}
	AddPlayer(player)
	go player.runLoop()
	go player.readLoop()
	loc.AddPlayer(player)
}

// IP returns the player's connecting address
func (p *Player) IP() net.Addr {
	return p.ip
}

// Close terminates the player's connection
func (p *Player) Close() {
	p.c.Close()
}

// Sex returns the sex of the user
func (p *Player) Sex() config.Sex {
	return p.sex
}

// Write writes the given string to the player, formatting if necessary
func (p *Player) Write(s string, a ...interface{}) {
	// TODO: error handling on write
	util.WriteLn(p.rw, "")
	util.WriteLn(p.rw, fmt.Sprintf(s, a...))
	p.prompt()
}

// Id returns the unique Id of this instance
func (p *Player) Id() util.Id {
	return p.id
}

// Name returns the user-visible name of the player
func (p *Player) Name() string {
	return p.name
}

// Desc returns the string that a user sees when they look at a user
func (p *Player) Desc() string {
	return p.desc
}

// String returns a string reprentation of the player (primarily for logging)
func (p *Player) String() string {
	return fmt.Sprintf("%v [%v]", p.name, p.id)
}

// setLocation changes the player's location and adds the player to the location's map
//
// This is the function that does the heavy lifting for moving a player from one room to another
// including keeping the user's location and the location map in sync
func (p *Player) SetLocation(loc *Location) {
	if loc != p.loc {
		// lock ordering
		if p.loc.Id() < loc.Id() {
			p.loc.RemovePlayer(p)
			loc.Add <- p
		} else {
			go func() {
				loc.Add <- p
				p.loc.Remove <- p
			}()
		}
		p.loc = loc
		p.handleCmd([]string{"look"})
	}
}

// Location returns the user's location in the world
func (p *Player) Location() *Location {
	return p.loc
}

// runLoop is a goroutine that synchronizes commands from the user and commands from the game
func (p *Player) runLoop() {
	p.Write(fmt.Sprintf("Welcome to NateMUD, %v", p.name))
	for {
		select {
		case cmd := <-p.cmd:
			p.handleCmd(cmd)

		case loc := <-p.Move:
			p.SetLocation(loc)

		case err := <-p.exit:
			if err != nil {
				log.Printf("Removing user %v from world. Error: %v", p.name, err)
			}
			RemovePlayer(p)
			break
		}
	}
}

// readLoop is a goroutine that just passes info from the player's input to the runLoop
func (p *Player) readLoop() {
	for {
		s, err := util.ReadLn(p.rw)
		if err != nil {
			log.Printf("Error reading from user %v: %v", p.name, err)
			p.exit <- err
			break
		}

		p.cmd <- strings.Split(strings.TrimSpace(s), " ")
	}
}

// promp shows the user's prompt to the user
func (p *Player) prompt() {
	// TODO: standard/custom prompts
	util.Write(p.rw, ">")
}

// timeout times the player out of the world
func (p *Player) timeout() {
	util.WriteLn(p.rw, "")
	util.WriteLn(p.rw, "You have timed out... good bye!")
	p.exit <- ErrTimeout
}

// HandleCommand handles commands specific to the player, such as inventory and player stats
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
	util.Write(p.rw, "Are you sure you want to quit? (y/n)")
	tokens := <-p.cmd
	switch tokens[0] {
	case "y", "yes":
		RemovePlayer(p)
	default:
		p.prompt()
	}
}

// handleCmd converts tokens from the user into a Command object, and attempts to handle it
func (p *Player) handleCmd(tokens []string) {
	cmd := NewCommand(p, tokens)
	if !p.HandleCommand(cmd) {
		p.loc.Cmd <- cmd
	}
}
