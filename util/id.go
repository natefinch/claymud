package util

import (
	"errors"
)

// somebody shoot me if I ever need more than 256 buckets

// BucketType makes ids unique across buckets by scoping them with another byte.
type BucketType byte

const (
	User BucketType = iota
)

var ErrBadId = errors.New("Invalid Id")

// Id is a type that allows for unique identification of an object
type Id int64

func (i Id) Key() []byte {
	key := make([]byte, 8)
	for n := range key {
		key[n] = byte(i >> uint(56-8*n))
	}

	return key
}

func ToId(key []byte) (Id, error) {
	if len(key) != 8 {
		return 0, ErrBadId
	}

	var i Id
	for n := range key {
		i |= Id(key[n]) << uint(56-8*n)
	}
	return i, nil
}
