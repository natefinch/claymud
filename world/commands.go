package world

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/game/social"
	"github.com/natefinch/claymud/util"
)

var started = time.Now()

var commands = map[string]func(*Command){}
var allCommands []CommandCfg

// Commands lets you configure how the commands get named.  The list of strings for
// each contain the aliases, they must all be unique.
type Commands struct {
	Look,
	Who,
	Tell,
	Quit,
	Say,
	Help,
	Uptime,
	ChatMode,
	Goto CommandCfg
}

// CommandCfg defines how a command gets run,  its aliases, and its help text.
type CommandCfg struct {
	Command string
	Aliases []string
	Help    string
}

var helptext string

// initCommands sets up the command names.
func initCommands(cfg Commands) {
	register(look, cfg.Look)
	register(who, cfg.Who)
	register(tell, cfg.Tell)
	register(quit, cfg.Quit)
	register(say, cfg.Say)
	register(help, cfg.Help)
	register(uptime, cfg.Uptime)
	register(gotoCmd, cfg.Goto)

	// this is a special "command" that just handles when someone hits enter without typing
	// anything.
	commands[""] = prompt

	// only register the chatmode command if it's allowed to be run
	if chatMode.Mode == ChatModeAllow {
		register(chatmode, cfg.ChatMode)
	}

	sort.SliceStable(allCommands, func(i, j int) bool { return allCommands[i].Command < allCommands[j].Command })

	var lines []string
	lines = append(lines, "List of available commands")
	lines = append(lines, "")
	lines = append(lines, "-- Standard Commands --")

	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 0, ' ', 0)

	for _, c := range allCommands {
		aliases := append([]string{c.Command}, c.Aliases...)
		fmt.Fprintf(w, "%s\t  %s\n", strings.Join(aliases, ", "), c.Help)
	}
	if err := w.Flush(); err != nil {
		// should be impossible
		panic(err)
	}
	lines = append(lines, buf.String())
	lines = append(lines, "")
	lines = append(lines, "-- Movement --")
	for _, dir := range game.AllDirections() {
		lines = append(lines, fmt.Sprintf("%v, %v", dir.Name, strings.Join(dir.Aliases, ", ")))
	}
	lines = append(lines, "")
	lines = append(lines, "-- Socials --")
	lines = append(lines, social.Names...)
	helptext = strings.Join(lines, "\n")
}

func register(f func(*Command), cmd CommandCfg) {
	names := append(cmd.Aliases, cmd.Command)
	for _, n := range names {
		if _, ok := commands[n]; ok {
			panic(fmt.Errorf("duplicate command name: %v %#v", n, cmd))
		}
		commands[n] = f
	}
	allCommands = append(allCommands, cmd)
}

// look handles the look command
func look(c *Command) {
	c.Actor.HandleLocal(func() {
		if c.Target() == "" {
			c.Loc.ShowRoom(c.Actor)
			return
		}
		desc, ok := c.Loc.LookTarget(c.Target())
		if ok {
			c.Actor.WriteString(desc)
		} else {
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

// gotoCmd is an admin command that lets you go to any room by room number or to the
// room a person is in.
func gotoCmd(c *Command) {
	if c.Target() == "" {
		c.Actor.HandleLocal(func() {
			c.Actor.WriteString("Goto where?")
		})
		return
	}
	num, err := strconv.Atoi(c.Target())
	if err == nil {
		loc, ok := locMap[util.ID(num)]
		if !ok {
			c.Actor.HandleLocal(func() {
				c.Actor.WriteString("There is no room with that number.")
			})
			return
		}
		c.Actor.Move(loc)
		return
	}
	c.Actor.HandleGlobal(func() {
		if p, ok := playerMap[c.Target()]; ok {
			moveEvent(c.Actor, p.loc)
			return
		}
		c.Actor.WriteString("There is player with that name.")
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
	c.Actor.HandleLocal(func() {
		if c.Target() != "" {
			c.helpdetails(c.Target())
		} else {
			c.Actor.WriteString(helptext)
		}
	})
}

func who(c *Command) {
	c.Actor.HandleGlobal(func() {
		c.Actor.WriteString("[Players]\n")
		for _, p := range *playerList {
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
			if c.Actor.Flag(PFlagChatmode) {
				c.Actor.WriteString("chat mode is on")
			} else {
				c.Actor.WriteString("chat mode is off")
			}
		case "":
			if c.Actor.Flag(PFlagChatmode) {
				c.Actor.UnsetFlag(PFlagChatmode)
				c.Actor.WriteString("chat mode is now off")
			} else {
				c.Actor.SetFlag(PFlagChatmode)
				c.Actor.WriteString("chat mode is now on")
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
	command = strings.ToLower(command)
	for _, cmd := range allCommands {
		if command == cmd.Command || contains(command, cmd.Aliases) {
			c.Actor.WriteString(cmd.Help)
		}
	}
}

func contains(s string, vals []string) bool {
	for i := range vals {
		if s == vals[i] {
			return true
		}
	}
	return false
}
