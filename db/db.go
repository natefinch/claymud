package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var (
	db          *bolt.DB
	ErrExists   = errors.New("value already exists")
	ErrNotFound = errors.New("value does not exist")
)

// ErrNoBucket is the error returned if we trying to reference a bucket that
// doesn't exist.
type ErrNoBucket string

// Error implemens the error interface.
func (e ErrNoBucket) Error() string {
	return "no bucket in db named " + string(e)
}

// Initialize sets up the application's configuration directory.
func Initialize(dir string) error {
	path := filepath.Join(dir, "mud.db")

	var err error
	db, err = bolt.Open(path, 0644, nil)
	if err != nil {
		return fmt.Errorf("Error opening database file %q: %s", path, err)
	}
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(players)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(users)
		return err
	})
}

// IsSetup returns true if the database has been setup.
func IsSetup() (bool, error) {
	var setup bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			// bucket doesn't exist
			return ErrNoBucket("users")
		}
		k, v := b.Cursor().First()
		setup = k != nil && v != nil
		return nil
	})
	return setup, err
}
