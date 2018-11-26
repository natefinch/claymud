package db

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestSaveLoadUser(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u := fakeUser(t)
	hash := []byte("secret")
	if err := st.CreateUser(u, hash); err != nil {
		t.Fatal(err)
	}
	found, err := st.FindUser(u.Username)
	if err != nil {
		t.Fatal(err)
	}
	usersEqual(t, u, found)
	creds, err := st.FindCreds(u.Username)
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username != u.Username {
		t.Errorf("Expected creds username %q but got %q", u.Username, creds.Username)
	}
	if !bytes.Equal(hash, creds.PwdHash) {
		t.Errorf("Expected password hash %x, but got %x", hash, creds.PwdHash)
	}
}

func TestDupeUser(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u := fakeUser(t)
	hash := []byte("secret")
	if err := st.CreateUser(u, hash); err != nil {
		t.Fatal(err)
	}
	err := st.CreateUser(u, hash)
	if _, ok := err.(ErrExists); !ok {
		t.Fatalf("expected to get ErrExists, but got %#v", err)
	}
}

func TestFindUserNotFound(t *testing.T) {
	st, cleanup := tmpStore(t)
	defer cleanup()
	u, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.FindUser(u.String())
	if _, ok := err.(ErrNotFound); !ok {
		t.Fatalf("expected to get ErrNotFound but got %#v", err)
	}
}

func usersEqual(t *testing.T, expected, got *User) {
	expectedT := expected.LastLogin
	gotT := got.LastLogin

	// zero out the comparison for times because they don't round trip in a comparable fashion.
	expected.LastLogin = got.LastLogin
	if !expectedT.Equal(gotT) {
		t.Errorf("expected time %v, but got %v", expectedT, gotT)
	}
	if !reflect.DeepEqual(*expected, *got) {
		t.Fatalf("expected %#v, got %#v", *expected, *got)
	}
}

func fakeUser(t *testing.T) *User {
	u, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	return &User{
		Username:  u.String(),
		LastIP:    "bar",
		LastLogin: time.Now(),
		Players:   []string{"bill", "jane"},
		Flags:     big.NewInt(12),
	}
}

func createFakeUser(t *testing.T, st *Store) *User {
	u := fakeUser(t)
	hash := []byte("secret")
	if err := st.CreateUser(u, hash); err != nil {
		t.Fatal(err)
	}
	return u
}
