package emote

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/claymud/game/gender"
)

func TestParse(t *testing.T) {
	ems, err := decodeEmotes(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
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

func TestPerformOther(t *testing.T) {
	defer patchGender(t)()

	ems, err := decodeEmotes(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	err = loadEmotes(ems)
	if err != nil {
		t.Fatal(err)
	}

	a := testActor{
		name: "fooName",
		sex:  gender.Male,
		buf:  &bytes.Buffer{},
	}
	b := testActor{
		name: "fooName2",
		sex:  gender.Female,
		buf:  &bytes.Buffer{},
	}

	others := &bytes.Buffer{}
	found := Perform("smile", a, b, others)
	if !found {
		t.Fatal("smile emote not found")
	}
	expected := "You smile at fooName2."
	if a.buf.String() != expected {
		t.Errorf("expected actor to get %q, but got %q", expected, a.buf.String())
	}
	expected = "fooName smiles at you."
	if b.buf.String() != expected {
		t.Errorf("expected actor to get %q, but got %q", expected, a.buf.String())
	}

}

/*
func (*Tests) TestParse(c *C) {
	ems, err := decodeEmotes(strings.NewReader(data))
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

func (*Tests) TestLoadEmotes(c *C) {
	ems, err := decodeEmotes(strings.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadEmotes(ems)
	c.Assert(err, IsNil)
	c.Assert(Names, HasLen, 2)
	c.Assert(emotes, HasLen, 2)
}

func (*Tests) TestDupeEmotes(c *C) {
	ems, err := decodeEmotes(strings.NewReader(dupes))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadEmotes(ems)
	c.Assert(err, ErrorMatches, `Duplicate emote defined: "smile"`)
}

func (*Tests) TestPerformSelf(c *C) {
	defer patchGender(c)()

	ems, err := decodeEmotes(strings.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(ems, HasLen, 2)

	err = loadEmotes(ems)
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

func patchGender(t *testing.T) func() {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(filepath.Join(d, "gender.toml"), genderData, 0666)
	gender.Initialize(d)
	return func() { os.RemoveAll(d) }
}

var _ Person = testActor{}

type testActor struct {
	name string
	sex  gender.Gender
	buf  *bytes.Buffer
}

func (a testActor) Name() string {
	return a.name
}

func (a testActor) Gender() gender.Gender {
	return a.sex
}

func (a testActor) Write(b []byte) (int, error) {
	return a.buf.Write(b)
}

var data = `
[[emote]]
name = "smile"

[emote.toSelf]
self = "You smile to yourself."
around = "{{.Actor}} smiles to {{.Xself}}."

[emote.toNoOne]
self = "You smile."
around = "{{.Actor}} smiles."

[emote.toOther]
self = "You smile at {{.Target}}."
target = "{{.Actor}} smiles at you."
around = "{{.Actor}} smiles at {{.Target}}."

[[emote]]
name = "jump"

# note there is no ToSelf section.

[emote.toNoOne]
self = "You jump around like a crazy person."
around = "{{.Actor}} jumps around like a crazy person."

[emote.toOther]
self = "You jump {{.Target}}."
target = "{{.Actor}} jumps you."
around = "{{.Actor}} jumps {{.Target}}."

`

var dupes = `
[[emote]]
name = "smile"

[emote.toSelf]
self = "You smile to yourself."
around = "{{.Actor}} smiles to {{.Xself}}."

[emote.toNoOne]
self = "You smile."
around = "{{.Actor}} smiles."

[emote.toOther]
self = "You smile at {{.Target}}."
target = "{{.Actor}} smiles at you."
around = "{{.Actor}} smiles at {{.Target}}."

[[emote]]
name = "smile"

[emote.toSelf]
self = "You smile to yourself."
around = "{{.Actor}} smiles to {{.Xself}}."

[emote.toNoOne]
self = "You smile."
around = "{{.Actor}} smiles."

[emote.toOther]
self = "You smile at {{.Target}}."
target = "{{.Actor}} smiles at you."
around = "{{.Actor}} smiles at {{.Target}}."

`

var genderData = []byte(`
[xself]
male = "himself"
female = "herself"
none = "itself"

[xe]
male = "he"
female = "she"
none = "it"

[xim]
male = "him"
female = "her"
none =  "it"

[xis]
male = "his"
female = "her"
none = "its"
`)
