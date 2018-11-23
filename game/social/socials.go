package social

import (
	"io"
	"log"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var (
	socials map[string]social
)

// Exists reports whether the given social exists as a command in the game.
func Exists(name string) bool {
	_, ok := socials[name]
	return ok
}

// Names is a list of the names of the available socials in the game
var Names []string

// noTarget is a collection of templates for an social that doesn't have a
// target.
type noTarget struct {
	Self   util.Template
	Around util.Template
}

// withTarget is a collection of templates for an social that has a target.
type withTarget struct {
	noTarget
	Target util.Template
}

// social is a struct that holds data about an social.
type social struct {
	Name    string
	ToSelf  *noTarget
	ToNoOne *noTarget
	ToOther *withTarget
}

func (e social) String() string {
	return e.Name
}

// Person is an interface that is used when filling the messages from an
// SocialTemplate.
type Person interface {
	Name() string
	Gender() game.Gender
	io.Writer
}

// socialData is the data we pass into the templates to generate the text.
type socialData struct {
	Actor  Person
	Target Person
}

// Perform attempts to perform the social named by cmd given the actor and target.
// Target may be nil if no target was specified.
// If the social exists, the output will be written to each of the writers.
// Perform reports whether the social was found.
func Perform(cmd string, actor Person, target Person, others io.Writer) bool {
	social, ok := socials[cmd]
	if !ok {
		return false
	}

	// TODO: Support more params?   Him/He/etc?
	data := socialData{
		Actor:  actor,
		Target: target,
	}

	switch {
	case target == nil:
		performToNoOne(cmd, social.ToNoOne, data, actor, others)
	case actor == target:
		performToSelf(cmd, social.ToSelf, data, actor, others)
	default:
		performToOther(cmd, social.ToOther, data, actor, target, others)
	}
	return true
}

func performToSelf(name string, toSelf *noTarget, data socialData, actor Person, others io.Writer) {
	if toSelf == nil {
		_, _ = io.WriteString(actor, "You can't do that to yourself.")
		return
	}
	err := toSelf.Self.Template.Execute(actor, data)
	if err != nil {
		logFillErr(name, "ToSelf.Self", data, err)
		// if there's an error running the social to the actor, just bail early.
		return
	}
	around(name, ".ToSelf", toSelf.Around, data, others)
}

func around(name, to string, tmpl util.Template, data socialData, others io.Writer) {
	err := tmpl.Execute(others, data)
	if err != nil {
		logFillErr(name, to+".Around", data, err)
		return
	}
	io.WriteString(others, "\n")
}

func performToNoOne(name string, toNoOne *noTarget, data socialData, actor Person, others io.Writer) {
	if toNoOne == nil {
		_, _ = io.WriteString(actor, "You can't do that.\n")
		return
	}
	err := toNoOne.Self.Template.Execute(actor, data)
	io.WriteString(actor, "\n")
	if err != nil {
		logFillErr(name, "ToNoOne.Self", data, err)
		// if there's an error running the social to the actor, just bail early.
		return
	}

	around(name, "ToNoOne", toNoOne.Around, data, others)
}

func performToOther(name string, toOther *withTarget, data socialData, actor Person, target Person, others io.Writer) {
	if toOther == nil {
		_, _ = io.WriteString(actor, "You can't do that to someone else.")
		return
	}
	err := toOther.Self.Template.Execute(actor, data)
	if err != nil {
		logFillErr(name, "ToOther.Self", data, err)
		// if there's an error running the social to the actor, just bail early.
		return
	}

	err = toOther.Target.Template.Execute(target, data)
	if err != nil {
		logFillErr(name, "ToOther.Target", data, err)
	}

	around(name, "ToOther", toOther.Around, data, others)
}

func logFillErr(name string, template string, data socialData, err error) {
	log.Printf("ERROR: filling social template %q for %s with data %v: %s", name, template, data, err)
}
