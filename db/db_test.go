package db

import (
	"io/ioutil"
	"os"
	"testing"
)

func tmpStore(t *testing.T) (st *Store, cleanup func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	store, err := Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store, func() {
		st.db.Close()
		os.RemoveAll(dir)
	}
}
