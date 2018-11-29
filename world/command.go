package world

import (
	"io"
	"log"
	"strings"

	"github.com/natefinch/claymud/game/social"
)

var chatMode ChatMode

// Command represents a command sent by a player.
type Command struct {
	Actor *Player
	Loc   *Location
	Cmd   []string
}

// Action returns the keyword of the command (always lowercase).
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

// Handle executes the command, if possible.
func (c *Command) Handle() {
	// everything in this function MUST be run through either the location's
	// handler or the actor's global handler.

	// order is important here, since earlier parsing overrides later
	// exits should always override anything else
	// socials are least important (so if you configure an social named "north" you won't
	// prevent yourc from going north... your social just won't work

	actionName := strings.Join(c.Cmd, " ")
	action, ok := c.Actor.loc.Actions[actionName]
	if ok {
		f := func() {
			if err := runLocAction(action.Filename, c.Actor, c.Actor.loc); err != nil {
				log.Printf("error running loc action %q in room %v: %v", actionName, c.Actor.loc.ID, err)
			}
		}
		if action.IsGlobal {
			c.Actor.HandleGlobal(f)
		} else {
			c.Actor.HandleLocal(f)
		}
		return
	}

	if !c.Actor.Flag(PFlagChatmode) {
		if c.handleExit() {
			return
		}
		if c.run() {
			return
		}
		if !c.handleSocial() {
			c.Actor.HandleLocal(func() {
				c.Actor.WriteString(`"` + c.Action() + `"` + " is not a valid command.")
			})
		}
		return
	}
	// chatmode is on, so we run directions if they are standalone,
	// otherwise all commands must be prefixed by the chatmode prefix

	isCmd := strings.HasPrefix(c.Action(), chatMode.Prefix)
	if !isCmd {
		if c.Target() == "" && c.handleExit() {
			return
		}
		// default to "say"
		chatModeSay(c)
		return
	}
	// strip prefix off the command name, so the rest of our string checks work
	c.Cmd[0] = c.Cmd[0][len(chatMode.Prefix):]
	if c.run() {
		return
	}
	if !c.handleSocial() {
		c.Actor.HandleLocal(func() {
			c.Actor.WriteString(`"` + c.Action() + `"` + " is not a valid command.")
		})
	}
}

func (c *Command) run() bool {
	f, ok := commands[c.Action()]
	if !ok {
		return false
	}
	f(c)
	return true
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

// handleSocial checks if the command is an existing social, and if so, handles
// it.
func (c *Command) handleSocial() (handled bool) {
	if !social.Exists(c.Action()) {
		return false
	}
	c.Actor.HandleLocal(func() {
		target := c.Loc.Target(c.Target())
		others := []io.Writer{}
		for _, p := range c.Loc.Players {
			if !p.Is(c.Actor) {
				others = append(others, p)
			}
		}
		var t social.Person
		if target != nil {
			t = target
		}
		social.Perform(c.Action(), c.Actor, t, io.MultiWriter(others...))
	})
	return true
}
