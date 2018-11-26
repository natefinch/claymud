package social

import (
	"bytes"
	"strings"
	"testing"

	"github.com/natefinch/claymud/game"
)

func TestParse(t *testing.T) {
	cfg, err := decodeConfig(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	ems := cfg.Socials
	if len(ems) != 2 {
		t.Fatalf("expected len 2, got %#v", ems)
	}

	e := ems[0]
	if e.Name != "smile" {
		t.Fatalf("expected to see smile, got %#v", e)
	}

	e = ems[1]
	if e.Name != "jump" {
		t.Fatalf("expected jump, got %#v", e)
	}
}

var female = game.Gender{
	Name:  "female",
	Xself: "herself",
	Xe:    "she",
	Xim:   "her",
	Xis:   "hers",
}

func TestPerformOther(t *testing.T) {
	cfg, err := decodeConfig(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	err = loadSocials(cfg.Socials)
	if err != nil {
		t.Fatal(err)
	}

	a := testActor{
		name:   "fooName",
		gender: female,
		buf:    &bytes.Buffer{},
	}
	b := testActor{
		name:   "fooName2",
		gender: female,
		buf:    &bytes.Buffer{},
	}

	others := &bytes.Buffer{}
	found := Perform("smile", a, b, others)
	if !found {
		t.Fatal("smile social not found")
	}
	expected := "You smile at fooName2."
	if a.buf.String() != expected {
		t.Errorf("expected actor to get %q, but got %q", expected, a.buf.String())
	}
	expected = "fooName smiles at you."
	if b.buf.String() != expected {
		t.Errorf("expected actor to get %q, but got %q", expected, b.buf.String())
	}

}

/*
func (*Tests) TestParse(c *C) {
	ems, err := decodeSocials(strings.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	e := ems[0]
	c.Assert(e.Name, Equals, "smile")
	c.Assert(e.ToSelf, NotNil)
	c.Assert(e.ToOther, NotNil)
	c.Assert(e.ToNoOne, NotNil)

	e = ems[1]
	c.Assert(e.Name, Equals, "jump")
	// note that there's no jump toself, so this should be nil
	c.Assert(e.ToSelf, IsNil)
	c.Assert(e.ToOther, NotNil)
	c.Assert(e.ToNoOne, NotNil)
}

func (*Tests) TestLoadSocials(c *C) {
	ems, err := decodeSocials(strings.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadSocials(ems)
	c.Assert(err, IsNil)
	c.Assert(Names, HasLen, 2)
	c.Assert(socials, HasLen, 2)
}

func (*Tests) TestDupeSocials(c *C) {
	ems, err := decodeSocials(strings.NewReader(dupes))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadSocials(ems)
	c.Assert(err, ErrorMatches, `Duplicate social defined: "smile"`)
}

func (*Tests) TestPerformSelf(c *C) {
	defer patchGender(c)()

	ems, err := decodeSocials(strings.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadSocials(ems)
	c.Assert(err, IsNil)

	a := testActor{
		name: "fooName",
		sex:  gender.Male,
		buf:  &bytes.Buffer{},
	}

	others := &bytes.Buffer{}
	Perform("smile", a, a, others)

	c.Assert(a.buf.String(), Equals, "You smile to yourself.")
	c.Assert(others.String(), Equals, "fooName smiles to himself.")
}
*/

var _ Person = testActor{}

type testActor struct {
	name   string
	gender game.Gender
	buf    *bytes.Buffer
}

func (a testActor) Name() string {
	return a.name
}

func (a testActor) Gender() game.Gender {
	return a.gender
}

func (a testActor) Write(b []byte) (int, error) {
	return a.buf.Write(b)
}

var data = `
[[social]]
name = "smile"

[social.toSelf]
self = "You smile to yourself."
around = "{{.Actor.Name}} smiles to {{.Actor.Gender.Xself}}."

[social.toNoOne]
self = "You smile."
around = "{{.Actor.Name}} smiles."

[social.toOther]
self = "You smile at {{.Target.Name}}."
target = "{{.Actor.Name}} smiles at you."
around = "{{.Actor.Name}} smiles at {{.Target.Name}}."

[[social]]
name = "jump"

# note there is no ToSelf section.

[social.toNoOne]
self = "You jump around like a crazy person."
around = "{{.Actor.Name}} jumps around like a crazy person."

[social.toOther]
self = "You jump {{.Target.Name}}."
target = "{{.Actor.Name}} jumps you."
around = "{{.Actor.Name}} jumps {{.Target.Name}}."

`

var dupes = `
[[social]]
name = "smile"

[social.toSelf]
self = "You smile to yourself."
around = "{{.Actor}} smiles to {{.Actor.Gender.Xself}}."

[social.toNoOne]
self = "You smile."
around = "{{.Actor}} smiles."

[social.toOther]
self = "You smile at {{.Target}}."
target = "{{.Actor}} smiles at you."
around = "{{.Actor}} smiles at {{.Target}}."

[[social]]
name = "smile"

[social.toSelf]
self = "You smile to yourself."
around = "{{.Actor}} smiles to {{.Actor.Gender.Xself}}."

[social.toNoOne]
self = "You smile."
around = "{{.Actor}} smiles."

[social.toOther]
self = "You smile at {{.Target}}."
target = "{{.Actor}} smiles at you."
around = "{{.Actor}} smiles at {{.Target}}."

`
