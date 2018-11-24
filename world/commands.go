package world

import (
	"fmt"
	"strings"
	"time"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/game/social"
)

var started = time.Now()

// todo: make this configurable
var commands = map[string]func(*Command){}

// CmdConfig lets you configure how the commands get named.  The list of strings for
// each contain the aliases, they must all be unique.
type CmdConfig struct {
	Look, Who, Tell, Quit, Say, Help, Uptime []string
	ChatMode                                 []string
}

// initCommands sets up the command names.
func initCommands(cfg CmdConfig) {
	register(look, cfg.Look)
	register(who, cfg.Who)
	register(tell, cfg.Tell)
	register(quit, cfg.Quit)
	register(say, cfg.Say)
	register(help, cfg.Help)
	register(uptime, cfg.Uptime)

	if chatMode.Mode == ChatModeAllow {
		register(chatmode, cfg.ChatMode)
	}

	// this is a special "command" that just handles when someone hits enter without typing
	// anything.
	register(prompt, []string{""})
}

func register(f func(*Command), names []string) {
	for _, n := range names {
		if _, ok := commands[n]; ok {
			panic(fmt.Errorf("duplicate command name: %v", n))
		}
		commands[n] = f
	}
}

// handles the look command

func look(c *Command) {
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

// say is when someone types the "say" command
func say(c *Command) {
	c.Actor.HandleLocal(func() {
		// message is everything but the command name
		msg := strings.Join(c.Cmd[1:], " ")
		doChat(msg, c)
	})
}

// chatModeSay is when someone types non-command text in chatmode.
func chatModeSay(c *Command) {
	c.Actor.HandleLocal(func() {
		// message includes first word
		msg := strings.Join(c.Cmd, " ")
		doChat(msg, c)
	})
}

func doChat(msg string, c *Command) {
	toOthers := c.Actor.Name() + ": " + msg
	for _, p := range c.Loc.Players {
		if !p.Is(c.Actor) {
			p.WriteString(toOthers)
		}
	}
	c.Actor.WriteString(toOthers)
}

func tell(c *Command) {
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

func help(c *Command) {
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
	lines = append(lines, "quit")
	lines = append(lines, "who")
	lines = append(lines, "uptime")
	lines = append(lines, "chatmode")
	lines = append(lines, "")
	lines = append(lines, "-- Movement --")
	for _, dir := range game.AllDirections() {
		lines = append(lines, fmt.Sprintf("%v, %v", dir.Name, strings.Join(dir.Aliases, ", ")))
	}
	lines = append(lines, "")
	lines = append(lines, "-- Socials --")
	lines = append(lines, social.Names...)
	c.Actor.Printf(strings.Join(lines, "\r\n"))
}

func who(c *Command) {
	c.Actor.HandleGlobal(func() {
		c.Actor.WriteString("[Players]\n")
		for _, p := range *playerList {
			// TODO: actually implement looking at things other than players
			c.Actor.WriteString(p.Name() + "\n")
		}
	})
}

func prompt(c *Command) {
	c.Actor.reprompt()
}

func chatmode(c *Command) {
	c.Actor.HandleLocal(func() {
		switch c.Target() {
		case "?":
			if c.Actor.chatmode {
				c.Actor.Print("chat mode is on")
			} else {
				c.Actor.Print("chat mode is off")
			}
		case "":
			c.Actor.chatmode = !c.Actor.chatmode
			if c.Actor.chatmode {
				c.Actor.Print("chat mode is now on")
			} else {
				c.Actor.Print("chat mode is now off")
			}
		default:
			c.Actor.Printf("unknown command target %v", c.Target())
		}
	})
}

func uptime(c *Command) {
	c.Actor.HandleLocal(func() {
		d := time.Since(started)
		switch {
		case d > 48*time.Hour:
			c.Actor.Printf("%d days", int(d.Hours()/24))
		case d > 2*time.Hour:
			c.Actor.Printf("%v", d.Round(time.Hour))
		case d > 5*time.Minute:
			c.Actor.Printf("%v", d.Round(time.Minute))
		default:
			c.Actor.Printf("%v", d.Round(time.Second))
		}
	})
}

func quit(c *Command) {
	// this must be done on the player's goroutine.
	c.Actor.handleQuit()
}

func (c *Command) helpdetails(command string) {
	switch command {
	case "help", "?":

	}
}
