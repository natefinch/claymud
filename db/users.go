package db

import (
	"net"
	"time"

	"github.com/boltdb/bolt"
)

var (
	users = []byte("users")
)

// userDoc is the structure that is stored in the database for a User.
type userDoc struct {
	PwdHash   []byte
	LastIP    net.Addr
	LastLogin time.Time
}

// UserExists reports whether a user with the username exists.
func UserExists(username string) (bool, error) {
	exists := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return nil
		}
		exists = b.Get([]byte(username)) != nil
		return nil
	})
	return exists, err
}

// Password returns the password hash for the user, or a nil slice if the user
// does not exist.
func Password(username string) ([]byte, error) {
	var hash []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		var u userDoc
		exists, err := get(b, []byte(username), &u)
		if err != nil {
			return err
		}
		if exists {
			hash = u.PwdHash
		}
		return nil
	})
	return hash, err
}

func SaveUser(username string, ip net.Addr, hash []byte) error {
	u := userDoc{
		PwdHash:   hash,
		LastIP:    ip,
		LastLogin: time.Now(),
	}

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(users)
		if err != nil {
			return err
		}
		return put(b, []byte(username), u)
	})
}
