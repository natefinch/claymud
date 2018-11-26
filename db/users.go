package db

import (
	"math/big"
	"time"

	"github.com/boltdb/bolt"
)

var usersBucket = []byte("users")

// User is the structure that is stored in the database for a User.
type User struct {
	Username  string
	LastIP    string
	LastLogin time.Time
	Players   []string
	Flags     *big.Int
}

// UserExists reports whether a user with the username exists.
func (st *Store) UserExists(username string) (bool, error) {
	exists := false
	err := st.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return ErrNoBucket("users")
		}
		exists = b.Get([]byte(username)) != nil
		return nil
	})
	return exists, err
}

// FindUser returns the user with the username.
func (st *Store) FindUser(username string) (User, error) {
	var u User
	err := st.db.View(func(tx *bolt.Tx) error {
		var err error
		u, err = getUser(tx, username)
		return err
	})
	return u, err
}

func getUser(tx *bolt.Tx, username string) (User, error) {
	var u User
	b := tx.Bucket(usersBucket)
	if b == nil {
		return u, ErrNoBucket("users")
	}
	exists, err := get(b, []byte(username), &u)
	if err != nil {
		return u, err
	}
	if !exists {
		return u, ErrNotFound("user")
	}
	return u, nil
}

// SaveUser saves the user to the db.
func (st *Store) SaveUser(u User) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		return saveUser(tx, u)
	})
}

func saveUser(tx *bolt.Tx, u User) error {
	b := tx.Bucket(usersBucket)
	if b == nil {
		return ErrNoBucket("users")
	}
	return put(b, []byte(u.Username), u)
}

// CreateUser creates the user only if it does not exist.
func (st *Store) CreateUser(u User, pwdHash []byte) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(usersBucket)
		if users == nil {
			return ErrNoBucket("users")
		}
		if users.Get([]byte(u.Username)) != nil {
			return ErrExists("user")
		}
		if err := put(users, []byte(u.Username), u); err != nil {
			return err
		}
		return saveCreds(tx, Credentials{Username: u.Username, PwdHash: pwdHash})
	})
}
