package world

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"src.natemud.org/config"
	"src.natemud.org/util"
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
func (self *Player) IP() net.Addr {
	return self.ip
}

// Close terminates the player's connection
func (self *Player) Close() {
	self.c.Close()
}

// Sex returns the sex of the user
func (self *Player) Sex() config.Sex {
	return self.sex
}

// Write writes the given string to the player, formatting if necessary
func (self *Player) Write(s string, a ...interface{}) {
	// TODO: error handling on write
	util.WriteLn(self.rw, "")
	util.WriteLn(self.rw, fmt.Sprintf(s, a...))
	self.prompt()
}

// Id returns the unique Id of this instance
func (self *Player) Id() util.Id {
	return self.id
}

// Name returns the user-visible name of the player
func (self *Player) Name() string {
	return self.name
}

// Desc returns the string that a user sees when they look at a user 
func (self *Player) Desc() string {
	return self.desc
}

// String returns a string reprentation of the player (primarily for logging)
func (self *Player) String() string {
	return fmt.Sprintf("%v [%v]", self.name, self.id)
}

// setLocation changes the player's location and adds the player to the location's map
// 
// This is the function that does the heavy lifting for moving a player from one room to another
// including keeping the user's location and the location map in sync
func (self *Player) SetLocation(loc *Location) {
	if loc != self.loc {
		// lock ordering
		if self.loc.Id() < loc.Id() {
			self.loc.RemovePlayer(self)
			loc.Add <- self
		} else {
			go func() {
				loc.Add <- self
				self.loc.Remove <- self
			}()
		}
		self.loc = loc
		self.handleCmd([]string{"look"})
	}
}

// Location returns the user's location in the world
func (self *Player) Location() *Location {
	return self.loc
}

// runLoop is a goroutine that synchronizes commands from the user and commands from the game 
func (self *Player) runLoop() {
	self.Write(fmt.Sprintf("Welcome to NateMUD, %v", self.name))
	for {
		select {
		case cmd := <-self.cmd:
			self.handleCmd(cmd)

		case loc := <-self.Move:
			self.SetLocation(loc)

		case err := <-self.exit:
			if err != nil {
				log.Printf("Removing user %v from world. Error: %v", self.name, err)
			}
			RemovePlayer(self)
			break
		}
	}
}

// readLoop is a goroutine that just passes info from the player's input to the runLoop
func (self *Player) readLoop() {
	for {
		s, err := util.ReadLn(self.rw)
		if err != nil {
			log.Printf("Error reading from user %v: %v", self.name, err)
			self.exit <- err
			break
		}

		self.cmd <- strings.Split(strings.TrimSpace(s), " ")
	}
}

// promp shows the user's prompt to the user
func (self *Player) prompt() {
	// TODO: standard/custom prompts
	util.Write(self.rw, ">")
}

// timeout times the player out of the world
func (self *Player) timeout() {
	util.WriteLn(self.rw, "")
	util.WriteLn(self.rw, "You have timed out... good bye!")
	self.exit <- ErrTimeout
}

// HandleCommand handles commands specific to the player, such as inventory and player stats
func (self *Player) HandleCommand(cmd *Command) bool {

	// TODO: implement player commands
	switch cmd.Action() {
	case "quit":
		self.handleQuit()
		return true
	}
	return false
}

// handleQuit asks the user if they really want to quit, and if they do, does so
func (self *Player) handleQuit() {
	util.Write(self.rw, "Are you sure you want to quit? (y/n)")
	tokens := <-self.cmd
	switch tokens[0] {
	case "y", "yes":
		RemovePlayer(self)
	default:
		self.prompt()
	}
}

// handleCmd converts tokens from the user into a Command object, and attempts to handle it
func (self *Player) handleCmd(tokens []string) {
	cmd := NewCommand(self, tokens)
	if !self.HandleCommand(cmd) {
		self.loc.Cmd <- cmd
	}
}
