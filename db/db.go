package db

import (
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
)

// Initialize sets up the application's configuration directory.
func Initialize(dir string) error {
	path := filepath.Join(dir, "mud.db")

	var err error
	db, err = bolt.Open(path, 0644, nil)
	if err != nil {
		return fmt.Errorf("Error opening database file %q: %s", path, err)
	}
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(playersBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(usersBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(credsBucket)
		return err
	})
}

// IsSetup returns true if the database has been setup.
func IsSetup() (bool, error) {
	var setup bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
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
