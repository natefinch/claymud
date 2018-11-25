package util

import (
	"testing"
)

func TestIdRoundTrip(t *testing.T) {
	var id ID = 45
	key := id.Key()
	id2, err := ToID(key)
	if err != nil {
		t.Fatal(err)
	}
	if id != id2 {
		t.Fatalf("expected %v, but got %v", id, id2)
	}
}
