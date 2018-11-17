package util

import (
	"bytes"
	"encoding/binary"
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
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, int64(i))

	return buf
}

func ToId(key []byte) (Id, error) {
	i, err := binary.ReadVarint(bytes.NewReader(key))
	if err != nil {
		return 0, err
	}
	return Id(i), nil
}
