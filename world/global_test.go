package world

import (
	"fmt"
	"testing"
)

func TestSortedPlayers(t *testing.T) {
	sp := sortedPlayers{}
	bob := &Player{ID: 1, name: "Bob"}
	alice := &Player{ID: 2, name: "Alice"}
	chris := &Player{ID: 3, name: "Chris"}
	sp.add(bob)
	sp.add(alice)
	sp.add(chris)

	order := fmt.Sprintf("%v%v%v", sp[0].ID, sp[1].ID, sp[2].ID)
	expected := "213"
	if order != expected {
		t.Fatalf("after add, should be %v, was %v", expected, order)
	}
	sp.remove(bob)
	if len(sp) != 2 {
		t.Errorf("len should be 2, was %v", len(sp))
	}
	order = fmt.Sprintf("%v%v", sp[0].ID, sp[1].ID)
	expected = "23"
	if order != expected {
		t.Fatalf("after remove, should be %v, was %v", expected, order)
	}
}
