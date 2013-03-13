package world

import (
	"fmt"
	"github.com/natefinch/natemud/config"
	"strings"
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
func (self *Command) Action() string {
	if len(self.Cmd) > 0 {
		return strings.ToLower(self.Cmd[0])
	}
	return ""
}

// Target returns the string that indicates the target of the command (if any) (always lowercase)
//
// This is just a shortcut to access the second token in the command
func (self *Command) Target() string {
	if len(self.Cmd) > 1 {
		return strings.ToLower(self.Cmd[1])
	}
	return ""
}

// Text returns the text of the command, not including the keyword (or target if hasTarget is false)
func (self *Command) Text(hasTarget bool) string {
	if hasTarget {
		return strings.Join(self.Cmd[2:], " ")
	}
	return strings.Join(self.Cmd[1:], " ")
}

func (self *Command) HandleAt(loc *Location) {
	// order is important here, since earlier parsing overrides later
	// exits should always override anything else
	// emotes are least important (so if you configure an emote named "north" you won't
	// prevent yourself from going north... your emote just won't work

	if self.exit(loc) {
		return
	}

	switch self.Action() {
	case "look", "l":
		self.look(loc)
	case "say", "'":
		self.say(loc)
	case "tell", "t":
		self.tell()
	case "help", "?":
		self.help()
	default:
		if !self.emote(loc) {
			self.Actor.Write("That is not a valid command.")
		}
	}
}

// Check to see if the command corresponds to an exit in the room
func (self *Command) exit(loc *Location) (handled bool) {
	// TODO: Handle custom exits
	// TODO: do we reject directions with a target, like "north Bob"?
	valid, room := loc.Exits().Find(self.Action())
	if !valid {
		return false
	}

	if room != nil {
		self.Actor.SetLocation(room)
	} else {
		self.Actor.Write("You can't go that way!")
	}
	return true
}

// checks if the command is an existing emote, and if so, handles it
func (self *Command) emote(loc *Location) (handled bool) {
	templ := config.FindEmote(self.Action())
	if templ != nil {
		target := loc.Target(self)
		var em *config.Emote
		if target == nil {
			em = config.MakeGlobalEmote(self.Actor, templ)
		} else {
			em = config.MakeTargetEmote(self.Actor, target, templ)
		}
		if em != nil {
			self.Actor.Write(em.ToSelf)
			for _, p := range loc.Players {
				if p.Id() != self.Actor.Id() {
					p.Write(em.ToOthers)
				}
			}
			if target != nil && em.ToTarget != "" {
				target.Write(em.ToTarget)
			}
			return true
		}
	}
	return false
}

// handles the look command
func (self *Command) look(loc *Location) {
	if self.Target() != "" {
		p := loc.Target(self)
		if p != nil {
			self.Actor.Write(p.Desc())
		} else {
			// TODO: actually implement looking at things other than players
			self.Actor.Write("You don't see that here.")
		}
		return
	}

	self.Actor.Write(loc.RoomDesc(self.Actor))
}

func (self *Command) say(loc *Location) {
	msg := strings.Join(self.Cmd[1:], " ")
	toOthers := fmt.Sprintf("%v says %v", self.Actor.Name(), msg)
	for _, p := range loc.Players {
		if p.Id() != self.Actor.Id() {
			p.Write(toOthers)
		}
	}
	self.Actor.Write("You say %v", msg)
}

func (self *Command) tell() {
	p := FindPlayer(self.Target())
	if p != nil {
		msg := strings.Join(self.Cmd[2:], " ")
		p.Write("%v tells you %v", self.Actor.Name(), msg)
		self.Actor.Write("You tell %v %v", p.Name(), msg)
	} else {
		p.Write("No one with that name exists.")
	}
}

func (self *Command) help() {
	if self.Target() != "" {
		self.helpdetails(self.Target())
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
		lines = append(lines, fmt.Sprintf("%v, %v", dir.Name(), strings.Join(dir.Aliases(), ", ")))
	}
	lines = append(lines, "")
	lines = append(lines, "-- Emotes --")
	lines = append(lines, config.GetEmoteNames()...)
	self.Actor.Write(strings.Join(lines, "\r\n"))
}

func (self *Command) helpdetails(command string) {
	switch command {
	case "help", "?":

	}
}
