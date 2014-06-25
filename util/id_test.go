package util

import (
	. "gopkg.in/check.v1"
)

func (*Tests) TestIdToKey(c *C) {
	var id Id = 0x12345678
	key := id.Key()
	c.Assert(key, Equals, []byte{1, 2, 3, 4, 5, 6, 7, 8})
}

func (*Tests) TestKeyToId(c *C) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	id, err := ToId(key)
	c.Assert(key, Equals, []byte{1, 2, 3, 4, 5, 6, 7, 8})
}
