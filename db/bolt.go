package db

import (
	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"
)

// get gob decodes the value for key in the given bucket into val.  It reports
// if a value with that key exists, and any error in retrieving or decoding the
// value.
func get(b *bolt.Bucket, key []byte, val interface{}) (bool, error) {
	v := b.Get(key)
	if v == nil {
		return false, nil
	}
	d := gob.NewDecoder(bytes.NewReader(v))
	return true, d.Decode(val)
}

var buf = &bytes.Buffer{}

// put gob encodes the value and puts it in the given bucket with the given key.
// This function assumes b was created from a writeable transaction.
func put(b *bolt.Bucket, key []byte, val interface{}) error {
	// we reuse the buffer here since this is, by definition locked by being in
	// a bolt write action.
	defer buf.Reset()
	e := gob.NewEncoder(buf)
	err := e.Encode(val)
	if err != nil {
		return err
	}
	return b.Put(key, buf.Bytes())
}
