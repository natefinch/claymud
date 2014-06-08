package emote

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/util"
)

var (
	emotes map[string]*emote
)

// Emotes is a list of the names of the available emotes in the game
var Emotes []string

// template is a struct that lets us unmarshal directly into a template.
type template struct {
	*template.Template
}

// UnmarshalText implements TextUnmarshaler.UnmarshalText.
func (e *emoteTemplate) UnmarshalText(text []byte) error {
	var err error
	e.Template, err = template.New("t").Parse(text)
	return fmt.Errorf("can't parse emote template: %s", err)
}

// noTarget is a collection of templates for an emote that doesn't have a
// target.
type noTarget struct {
	Self   template
	Around template
}

// withTarget is a collection of templates for an emote that has a target.
type withTarget struct {
	noTarget
	Target template
}

// emote is a struct that holds data about an emote.
type emote struct {
	Name    string
	ToSelf  *noTarget
	ToNoOne *noTarget
	ToOther *withTarget
}

// Person is an interface that is used when filling the messages from an
// EmoteTemplate.
type Person interface {
	Name() string
	Gender() Sex
	io.Writer
}

// emoteData is the data we pass into the templates to generate the text.
type emoteData struct {
	Actor  string
	Target string
	XSelf  string
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
		Actor: actor.Name(),
		XSelf: actor.Gender().Xself(),
	}

	if target != nil {
		data.Target = target.Name()
	}

	switch {
	case target == nil:
		performToNoOne(emote, data, actor, others)
	case actor == target:
		performToSelf(emote, data, actor, others)
	default:
		performToOther(emote, data, actor, target, others)
	}
	return true
}

func performToSelf(emote emote, actor Person, others io.Writer) {
	if emote.ToSelf == nil {
		_, _ = actor.Write([]byte("You can't do that to yourself."))
		return
	}
	err := emote.ToSelf.Self.Execute(actor, data)
	if err != nil {
		logFillErr(cmd, "ToSelf.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}

	err = emote.ToSelf.Around.Execute(others, data)
	if err != nil {
		logFillErr(emote.Name, "ToSelf.Around", data, err)
	}
}

func performToNoOne(emote emote, actor Person, others io.Writer) {
	if emote.ToNoOne == nil {
		_, _ = actor.Write([]byte("You can't do that."))
		return
	}
	err := emote.ToNoOne.Self.Execute(actor, data)
	if err != nil {
		logFillErr(cmd, "ToNoOne.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}

	err = emote.ToNoOne.Around.Execute(others, data)
	if err != nil {
		logFillErr(emote.Name, "ToNoOne.Around", data, err)
	}
}

func performToOther(emote emote, actor Person, target Person, others io.Writer) {
	if emote.ToOther == nil {
		_, _ = actor.Write([]byte("You can't do that to someone else."))
		return
	}
	err := emote.ToOther.Self.Execute(actor, data)
	if err != nil {
		logFillErr(cmd, "ToOther.Self", data, err)
		// if there's an error running the emote to the actor, just bail early.
		return
	}

	err = emote.ToOther.Target.Execute(target, data)
	if err != nil {
		logFillErr(cmd, "ToOther.Target", data, err)
	}

	err = emote.ToOther.Around.Execute(others, data)
	if err != nil {
		logFillErr(emote.Name, "ToOther.Around", data, err)
	}
}

func logFillErr(emote, template string, data emoteData, err error) {
	log.Printf("ERROR filling emote template %q for %s with data %v: %s", emote, template, data, err)
}
