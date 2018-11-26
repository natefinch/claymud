package auth

import (
	"io"
	"math/big"

	"github.com/natefinch/claymud/util"
)

// UFlag represents a flag (bit) set on a User.
type UFlag int

// All possible user flags.
//
// DO NOT REARRANGE OR COMMENT OUT VALUES.  If you need to deprecate a value,
// append _DEPRECATED on the end of the name.  New values must be appended to
// this list, never inserted anywhere else.
const (
	UFlagAdmin UFlag = iota
)

// An user is a username and password and connection info.
type User struct {
	ID       util.ID
	Username string
	Players  []string
	bits     *big.Int
	io.Closer
	util.WriteScanner
}

// Flag reports if the given flag has been set to true for the user.
func (u *User) Flag(f UFlag) bool {
	return u.bits.Bit(int(f)) == 1
}

// SetFlag sets the given flag to true for the user.
func (u *User) SetFlag(f UFlag) {
	u.bits.SetBit(u.bits, int(f), 1)
}

// UnsetFlag sets the given flag to false for the user.
func (u *User) UnsetFlag(f UFlag) {
	u.bits.SetBit(u.bits, int(f), 0)
}
