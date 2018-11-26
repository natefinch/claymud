package db

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/natefinch/claymud/game"
)

func TestSaveLoadPlayer(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u := createFakeUser(t, st)
	p := fakePlayer(t)
	err := st.CreatePlayer(u.Username, p)
	if err != nil {
		t.Fatal(err)
	}
	found, err := st.FindPlayer(p.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p, found) {
		t.Fatalf("expected %#v, got %#v", p, found)
	}
}

func TestDupePlayer(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u := createFakeUser(t, st)
	p := fakePlayer(t)
	if err := st.CreatePlayer(u.Username, p); err != nil {
		t.Fatal(err)
	}
	err := st.CreatePlayer(u.Username, p)
	if _, ok := err.(ErrExists); !ok {
		t.Fatalf("expected to get ErrExists, but got %#v", err)
	}
}

func TestFindPlayerNotFound(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.FindPlayer(u.String())
	if _, ok := err.(ErrNotFound); !ok {
		t.Fatalf("expected to get ErrNotFound but got %#v", err)
	}
}

func fakePlayer(t *testing.T) *Player {
	name, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	return &Player{
		Name:        name.String(),
		Description: "bar",
		Gender: game.Gender{
			Name:  "foo",
			Xself: "fooself",
			Xe:    "fooe",
			Xim:   "fooim",
			Xis:   "foois",
		},
		Flags: big.NewInt(17),
	}
}
