package util

import (
	"bytes"
	"encoding/binary"
)

// ID is a type that allows for unique identification of an object
type ID uint64

// Key returns a byte representation of this ID for use with a key value database.
func (i ID) Key() []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, uint64(i))

	return buf
}

// ToID converts a key/value db key into an Id.
func ToID(key []byte) (ID, error) {
	i, err := binary.ReadUvarint(bytes.NewReader(key))
	if err != nil {
		return 0, err
	}
	return ID(i), nil
}
