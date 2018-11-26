package db

import "github.com/boltdb/bolt"

var credsBucket = []byte("credentials")

// An Credentials object holds a user's login credentials.
type Credentials struct {
	Username string
	PwdHash  []byte
}

// FindCreds returns the user's credentials.
func FindCreds(username string) (Credentials, error) {
	var c Credentials
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(credsBucket)
		if b == nil {
			return ErrNoBucket("credentials")
		}
		exists, err := get(b, []byte(username), &c)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound("credentials")
		}
		return nil
	})
	return c, err
}

// SaveCreds saves the user's credentials to the db.
func SaveCreds(c Credentials) error {
	return db.Update(func(tx *bolt.Tx) error {
		return saveCreds(tx, c)
	})
}

func saveCreds(tx *bolt.Tx, c Credentials) error {
	b := tx.Bucket(credsBucket)
	if b == nil {
		return ErrNoBucket("credentials")
	}
	return put(b, []byte(c.Username), c)
}
