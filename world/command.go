package world

import (
	"fmt"
	"io"
	"strings"

	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/game/emote"
)

// Represents a command sent by a player
type Command struct {
	Actor *Player
	Cmd   []string
}

// NewCommand creates a new command from the given player and the given tokens
func NewCommand(actor *Player, cmd []string) *Command {
	return &Command{actor, cmd}
}

// Action returns the keyword of the command (always lowercase)
//
// This is just a shortcut to access the first token in the command
func (c *Command) Action() string {
	if len(c.Cmd) > 0 {
		return strings.ToLower(c.Cmd[0])
	}
	return ""
}

// Target returns the string that indicates the target of the command (if any) (always lowercase)
//
// This is just a shortcut to access the second token in the command
func (c *Command) Target() string {
	if len(c.Cmd) > 1 {
		return strings.ToLower(c.Cmd[1])
	}
	return ""
}

// Text returns the text of the command, not including the keyword (or target if hasTarget is false)
func (c *Command) Text(hasTarget bool) string {
	if hasTarget {
		return strings.Join(c.Cmd[2:], " ")
	}
	return strings.Join(c.Cmd[1:], " ")
}

func (c *Command) HandleAt(loc *Location) {
	// order is important here, since earlier parsing overrides later
	// exits should always override anything else
	// emotes are least important (so if you configure an emote named "north" you won't
	// prevent yourc from going north... your emote just won't work

	if c.exit(loc) {
		return
	}

	switch c.Action() {
	case "look", "l":
		c.look(loc)
	case "say", "'":
		c.say(loc)
	case "tell", "t":
		c.tell()
	case "help", "?":
		c.help()
	default:
		if !c.emote(loc) {
			c.Actor.Writef("That is not a valid command.")
		}
	}
}

// Check to see if the command corresponds to an exit in the room
func (c *Command) exit(loc *Location) (handled bool) {
	// TODO: Handle custom exits
	// TODO: do we reject directions with a target, like "north Bob"?
	valid, room := loc.Exits.Find(c.Action())
	if !valid {
		return false
	}

	if room != nil {
		c.Actor.Move(room)
	} else {
		c.Actor.Writef("You can't go that way!")
	}
	return true
}

// checks if the command is an existing emote, and if so, handles it
func (c *Command) emote(loc *Location) (handled bool) {
	target := loc.Target(c)
	others := []io.Writer{}
	for _, p := range loc.Players {
		if p.Id() != c.Actor.Id() {
			others = append(others, p)
		}
	}
	return emote.Perform(c.Action(), c.Actor, target, io.MultiWriter(others...))
}

// handles the look command
func (c *Command) look(loc *Location) {
	if c.Target() != "" {
		p := loc.Target(c)
		if p != nil {
			c.Actor.Writef(p.Desc)
		} else {
			// TODO: actually implement looking at things other than players
			c.Actor.Writef("You don't see that here.")
		}
		return
	}

	loc.ShowRoom(c.Actor)
}

func (c *Command) say(loc *Location) {
	msg := strings.Join(c.Cmd[1:], " ")
	toOthers := fmt.Sprintf("%v says %v", c.Actor.Name, msg)
	for _, p := range loc.Players {
		if p.Id() != c.Actor.Id() {
			p.Writef(toOthers)
		}
	}
	c.Actor.Writef("You say %v", msg)
}

func (c *Command) tell() {
	p := FindPlayer(c.Target())
	if p != nil {
		msg := strings.Join(c.Cmd[2:], " ")
		p.Writef("%v tells you %v", c.Actor.Name(), msg)
		c.Actor.Writef("You tell %v %v", p.Name(), msg)
	} else {
		p.Writef("No one with that name exists.")
	}
}

func (c *Command) help() {
	if c.Target() != "" {
		c.helpdetails(c.Target())
		return
	}
	var lines []string
	lines = append(lines, "List of available commands")
	lines = append(lines, "")
	lines = append(lines, "-- Standard Commands --")
	lines = append(lines, "help, ?")
	lines = append(lines, "look, l")
	lines = append(lines, "say, '")
	lines = append(lines, "tell, t")
	lines = append(lines, "")
	lines = append(lines, "-- Movement --")
	for _, dir := range config.AllDirections() {
		lines = append(lines, fmt.Sprintf("%v, %v", dir.Name, strings.Join(dir.Aliases, ", ")))
	}
	lines = append(lines, "")
	lines = append(lines, "-- Emotes --")
	lines = append(lines, emote.Names...)
	c.Actor.Writef(strings.Join(lines, "\r\n"))
}

func (c *Command) helpdetails(command string) {
	switch command {
	case "help", "?":

	}
}
