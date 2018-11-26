package db

import (
	"math/big"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var playersBucket = []byte("players")

// Player is the structure that is stored in the database for a Player.
type Player struct {
	Name        string
	Description string
	ID          util.ID
	Gender      game.Gender
	Flags       *big.Int
}

// FindPlayer returns the player with the given name. This is a
// case-insensitive check.
func (st *Store) FindPlayer(name string) (*Player, error) {
	var p Player
	err := st.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(playersBucket)
		if b == nil {
			return ErrNoBucket("players")
		}
		key := []byte(strings.ToLower(name))
		exists, err := get(b, key, &p)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound("player")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PlayerExists reports whether a player with the given name exists. This is a
// case-insensitive check.
func (st *Store) PlayerExists(name string) (bool, error) {
	exists := false
	err := st.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(playersBucket)
		if b == nil {
			return ErrNoBucket("players")
		}
		val := b.Get([]byte(strings.ToLower(name)))
		exists = val != nil
		return nil
	})
	return exists, err
}

// SavePlayer saves the player's data to the db.
func (st *Store) SavePlayer(p *Player) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(playersBucket)
		if b == nil {
			return ErrNoBucket("players")
		}
		return put(b, []byte(strings.ToLower(p.Name)), p)
	})
}

// CreatePlayer saves the player only if it does not already exist.
func (st *Store) CreatePlayer(username string, p *Player) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		players := tx.Bucket(playersBucket)
		if players == nil {
			return ErrNoBucket("players")
		}
		key := []byte(strings.ToLower(p.Name))
		val := players.Get(key)
		if val != nil {
			return ErrExists("player")
		}
		id, err := players.NextSequence()
		if err != nil {
			return err
		}
		p.ID = util.ID(id)
		if err := put(players, key, p); err != nil {
			return err
		}

		u, err := getUser(tx, username)
		if err != nil {
			return err
		}
		u.Players = append(u.Players, p.Name)
		return saveUser(tx, u)
	})
}
