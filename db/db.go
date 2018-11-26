package db

import (
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// Store contains all the functionality of persistence.
type Store struct {
	db *bolt.DB
}

// Initialize sets up the application's configuration directory.
func Init(dir string) (*Store, error) {
	path := filepath.Join(dir, "mud.db")
	db, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, fmt.Errorf("Error opening database file %q: %s", path, err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
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
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// IsSetup returns true if the database has been setup.
func (st *Store) IsSetup() (bool, error) {
	var setup bool
	err := st.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(usersBucket)
		if users == nil {
			// bucket doesn't exist
			return ErrNoBucket("users")
		}
		k, v := users.Cursor().First()
		setup = k != nil && v != nil
		return nil
	})
	return setup, err
}
