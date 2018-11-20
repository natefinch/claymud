package emote

import (
	"io"
	"log"

	"github.com/natefinch/claymud/game/gender"
	"github.com/natefinch/claymud/util"
)

var (
	emotes map[string]emote
)

// Exists reports whether the given emote exists as a command in the game.
func Exists(name string) bool {
	_, ok := emotes[name]
	return ok
}

// Names is a list of the names of the available emotes in the game
var Names []string

// noTarget is a collection of templates for an emote that doesn't have a
// target.
type noTarget struct {
	Self   util.Template
	Around util.Template
}

// withTarget is a collection of templates for an emote that has a target.
type withTarget struct {
	noTarget
	Target util.Template
}

// emote is a struct that holds data about an emote.
type emote struct {
	Name    string
	ToSelf  *noTarget
	ToNoOne *noTarget
	ToOther *withTarget
}

func (e emote) String() string {
	return e.Name
}

// Person is an interface that is used when filling the messages from an
// EmoteTemplate.
type Person interface {
	Name() string
	Gender() gender.Gender
	io.Writer
}

// emoteData is the data we pass into the templates to generate the text.
type emoteData struct {
	Actor  Person
	Target Person
}

// Perform attempts to perform the emote named by cmd given the actor and target.
// Target may be nil if no target was specified.
// If the emote exists, the output will be written to each of the writers.
// Perform reports whether the emote was found.
func Perform(cmd string, actor Person, target Person, others io.Writer) bool {
	emote, ok := emotes[cmd]
	if !ok {
		return false
	}

	// TODO: Support more params?   Him/He/etc?
	data := emoteData{
		Actor:  actor,
		Target: target,
	}

	switch {
	case target == nil:
		performToNoOne(cmd, emote.ToNoOne, data, actor, others)
	case actor == target:
		performToSelf(cmd, emote.ToSelf, data, actor, others)
	default:
		performToOther(cmd, emote.ToOther, data, actor, target, others)
	}
	return true
}

func performToSelf(name string, toSelf *noTarget, data emoteData, actor Person, others io.Writer) {
	if toSelf == nil {
		_, _ = io.WriteString(actor, "You can't do that to yourself.")
		return
	}
	err := toSelf.Self.Template.Execute(actor, data)
	if err != nil {
		logFillErr(name, "ToSelf.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}
	around(name, ".ToSelf", toSelf.Around, data, others)
}

func around(name, to string, tmpl util.Template, data emoteData, others io.Writer) {
	err := tmpl.Execute(others, data)
	if err != nil {
		logFillErr(name, to+".Around", data, err)
		return
	}
}

func performToNoOne(name string, toNoOne *noTarget, data emoteData, actor Person, others io.Writer) {
	if toNoOne == nil {
		_, _ = io.WriteString(actor, "You can't do that.")
		return
	}
	err := toNoOne.Self.Template.Execute(actor, data)
	if err != nil {
		logFillErr(name, "ToNoOne.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}

	around(name, "ToNoOne", toNoOne.Around, data, others)
}

func performToOther(name string, toOther *withTarget, data emoteData, actor Person, target Person, others io.Writer) {
	if toOther == nil {
		_, _ = io.WriteString(actor, "You can't do that to someone else.")
		return
	}
	err := toOther.Self.Template.Execute(actor, data)
	if err != nil {
		logFillErr(name, "ToOther.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}

	err = toOther.Target.Template.Execute(target, data)
	if err != nil {
		logFillErr(name, "ToOther.Target", data, err)
	}

	around(name, "ToOther", toOther.Around, data, others)
}

func logFillErr(name string, template string, data emoteData, err error) {
	log.Printf("ERROR: filling emote template %q for %s with data %v: %s", name, template, data, err)
}
