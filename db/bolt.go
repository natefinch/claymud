package db

import (
	"encoding/json"

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
	return true, json.Unmarshal(v, val)
}

// put gob encodes the value and puts it in the given bucket with the given key.
// This function assumes b was created from a writeable transaction.
func put(b *bolt.Bucket, key []byte, val interface{}) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return b.Put(key, bytes)
}
