package util

import (
	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"
)

// Get gob decodes the value for key in the given bucket into val.  It reports
// if a value with that key exists, and any error in retrieving or decoding the
// value.
func Get(b *bolt.Bucket, key []byte, val interface{}) (exists bool, err error) {
	v := b.Get(key)
	if v == nil {
		return false, nil
	}
	d := gob.NewDecoder(bytes.NewReader(v))
	return true, d.Decode(val)
}

var buf = &bytes.Buffer{}

// Put gob encodes the value and puts it in the given bucket with the given key.
// This function assumes b was created from a writeable transaction.
func Put(b *bolt.Bucket, key []byte, val interface{}) error {
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
