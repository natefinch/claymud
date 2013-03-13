package config

import (
	"github.com/natefinch/natemud/util"
	"log"
	"strings"
)

var (
	emoteTemplates map[string]*EmoteTemplate
)

// Emote is a struct that holds messages that get displayed to players at a location
//
// There is a message that gets sent to the sender, one to the target of the emote (if any)
// and one to everyone else in the room.
type Emote struct {
	ToSelf   string
	ToTarget string
	ToOthers string
}

// EmoteTemplate is a struct that holds the messages that get displayed due to an emote
//
// There's one string for each message that should get displayed depending on how the emote is used
type EmoteTemplate struct {
	ToSelf_Self   string
	ToSelf_Around string

	ToNoOne_Self   string
	ToNoOne_Around string

	ToOther_Self   string
	ToOther_Target string
	ToOther_Around string
}

// FindEmote returns an emote template that corresponds to the given command
func FindEmote(name string) *EmoteTemplate {
	// TODO: implement abbreviations
	return emoteTemplates[strings.ToLower(name)]
}

func GetEmoteNames() []string {
	names := make([]string, len(emoteTemplates))
	x := 0
	for name, _ := range emoteTemplates {
		names[x] = name
		x++
	}
	return names
}

// Person is an interface that is used when filling the messages from an EmoteTemplate
type Person interface {
	Name() string
	Sex() Sex
}

// MakeTargetEmote creates an emote from the given template, actor, and target
//
// For example, if Bill types "smile Bob", the smile EmoteTemplate would be filled out with
// Bill as the actor and Bob as the target
func MakeTargetEmote(actor Person, target Person, templ *EmoteTemplate) (emote *Emote) {
	emote = &Emote{}

	// TODO: Support more params?   Him/He/etc?
	params := map[string]string{
		"Actor":  actor.Name(),
		"Target": target.Name(),
		"XSelf":  actor.Sex().Xself()}

	// to self
	if actor == target {
		s, err := util.FillTemplate(templ.ToSelf_Self, params)
		if err != nil {
			return nil
		}
		emote.ToSelf = s
		s, err = util.FillTemplate(templ.ToSelf_Around, params)
		if err != nil {
			return nil
		}
		emote.ToOthers = s
		return
	}

	// to someone else
	s, err := util.FillTemplate(templ.ToOther_Self, params)
	if err != nil {
		return nil
	}
	emote.ToSelf = s
	s, err = util.FillTemplate(templ.ToOther_Target, params)
	if err != nil {
		return nil
	}
	emote.ToTarget = s
	s, err = util.FillTemplate(templ.ToOther_Around, params)
	if err != nil {
		return nil
	}
	emote.ToOthers = s
	return
}

// MakeGlobalEmote creates an emote from the given template and originator
//
// It should be used when an emote is not targeted at any particular person
func MakeGlobalEmote(actor Person, templ *EmoteTemplate) (emote *Emote) {
	emote = &Emote{}

	// TODO: Support more params?   Him/He/etc?
	params := map[string]string{
		"Actor": actor.Name(),
		"XSelf": actor.Sex().Xself()}

	s, err := util.FillTemplate(templ.ToNoOne_Self, params)
	if err != nil {
		return nil
	}
	emote.ToSelf = s
	s, err = util.FillTemplate(templ.ToNoOne_Around, params)
	if err != nil {
		return nil
	}
	emote.ToOthers = s
	return
}

// init creates the emoteTemplate map and loads emotes into it
func init() {
	emoteTemplates = make(map[string]*EmoteTemplate)

	// TODO: put this in a config file
	smile := &EmoteTemplate{
		// to self
		"You smile to yourself.",
		"{{.Actor}} smiles to {{.Xself}}.",

		// to no one
		"You smile.",
		"{{.Actor}} smiles.",

		// to someone else
		"You smile at {{.Target}}.",
		"{{.Actor}} smiles at you.",
		"{{.Actor}} smiles at {{.Target}}."}
	emoteTemplates["smile"] = smile

	log.Printf("Loaded %v emotes", len(emoteTemplates))
}
