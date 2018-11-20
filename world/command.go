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
	Loc   *Location
	Cmd   []string
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

func (c *Command) Handle() {
	// everything in this function MUST be run through either the location's
	// handler or the actor's global handler.

	// order is important here, since earlier parsing overrides later
	// exits should always override anything else
	// emotes are least important (so if you configure an emote named "north" you won't
	// prevent yourc from going north... your emote just won't work

	if c.handleExit() {
		return
	}

	switch c.Action() {
	case "look", "l":
		c.look()
	case "say", "'":
		c.say()
	case "tell", "t":
		c.tell()
	case "help", "?":
		c.help()
	case "quit":
		c.Loc.Handle(c.Actor.handleQuit)
	case "":
		c.Actor.reprompt()
	default:
		if !c.handleEmote() {
			c.Actor.HandleLocal(func() {
				c.Actor.WriteString(`"` + c.Action() + `"` + " is not a valid command.")
			})
		}
	}
}

// Check to see if the command corresponds to an exit in the room
func (c *Command) handleExit() (handled bool) {
	// TODO: Handle custom exits
	// TODO: do we reject directions with a target, like "north Bob"?
	valid, room := c.Loc.Exits.Find(c.Action())
	if !valid {
		return false
	}

	if room != nil {
		c.Actor.Move(room)
	} else {
		c.Actor.HandleLocal(func() {
			c.Actor.WriteString("You can't go that way!")
		})
	}
	return true
}

// handleEmote checks if the command is an existing emote, and if so, handles
// it.
func (c *Command) handleEmote() (handled bool) {
	if !emote.Exists(c.Action()) {
		return false
	}
	c.Actor.HandleLocal(func() {
		target := c.Loc.Target(c)
		others := []io.Writer{}
		for _, p := range c.Loc.Players {
			if !p.Is(c.Actor) {
				others = append(others, p)
			}
		}
		var t emote.Person
		if target != nil {
			t = target
		}
		emote.Perform(c.Action(), c.Actor, t, io.MultiWriter(others...))
	})
	return true
}

// handles the look command
func (c *Command) look() {
	c.Actor.HandleLocal(func() {
		if c.Target() == "" {
			c.Loc.ShowRoom(c.Actor)
			return
		}
		p := c.Loc.Target(c)
		if p != nil {
			c.Actor.WriteString(p.Desc)
		} else {
			// TODO: actually implement looking at things other than players
			c.Actor.WriteString("You don't see that here.\n")
		}
	})
}

func (c *Command) say() {
	c.Actor.HandleLocal(func() {
		msg := strings.Join(c.Cmd[1:], " ")
		toOthers := c.Actor.Name() + " says " + msg
		for _, p := range c.Loc.Players {
			if p.Is(c.Actor) {
				p.WriteString(toOthers)
			}
		}
		c.Actor.WriteString("You say " + msg)
	})
}

func (c *Command) tell() {
	c.Actor.HandleGlobal(func() {
		target, ok := FindPlayer(c.Target())
		if ok {
			msg := strings.Join(c.Cmd[2:], " ")
			target.Printf("%v tells you: %v", c.Actor.Name(), msg)
			target.prompt()
			c.Actor.Printf("You tell %v: %v", target.Name(), msg)
		} else {
			c.Actor.WriteString("No one with that name exists.")
		}
	})
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
	c.Actor.Printf(strings.Join(lines, "\r\n"))
}

func (c *Command) helpdetails(command string) {
	switch command {
	case "help", "?":

	}
}
