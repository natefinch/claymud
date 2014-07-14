package db

import (
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"

	"github.com/natefinch/natemud/config"
)

var (
	db *bolt.DB
)

// Initialize sets up the application's configuration directory.
func Initialize() error {
	path := filepath.Join(config.DataDir(), "natemud.db")

	var err error
	db, err = bolt.Open(path, 0644, nil)
	if err != nil {
		return fmt.Errorf("Error opening database file %q: %s", path, err)
	}
	return nil
}

// IsSetup returns true if the database has been setup.
func IsSetup() (bool, error) {
	var setup bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			// bucket doesn't exist
			return nil
		}
		k, v := b.Cursor().First()
		setup = k != nil && v != nil
		return nil
	})
	return setup, err
}
