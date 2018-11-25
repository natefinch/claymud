package db

import (
	"math/big"
	"time"

	"github.com/boltdb/bolt"
)

var (
	users = []byte("users")
)

// User is the structure that is stored in the database for a User.
type User struct {
	PwdHash   []byte
	LastIP    string
	LastLogin time.Time
	Players   []string
	Flags     *big.Int
}

// UserExists reports whether a user with the username exists.
func UserExists(username string) (bool, error) {
	exists := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return ErrNoBucket("users")
		}
		exists = b.Get([]byte(username)) != nil
		return nil
	})
	return exists, err
}

// FindUser returns the user with the username.
func FindUser(username string) (User, error) {
	var u User
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return ErrNoBucket("users")
		}
		exists, err := get(b, []byte(username), &u)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
		return nil
	})
	return u, err
}

// SaveUser saves the user to the db.
func SaveUser(username string, u User) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return ErrNoBucket("users")
		}
		return put(b, []byte(username), u)
	})
}

// CreateUser creates the user only if it does not exist.
func CreateUser(username string, u User) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return ErrNoBucket("users")
		}
		if b.Get([]byte(username)) != nil {
			return ErrExists
		}
		return put(b, []byte(username), u)
	})
}
