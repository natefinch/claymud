package db

import "fmt"

// ErrNotFound indicates the thing doesn't exist.
type ErrNotFound string

// Error implements the error interface.
func (e ErrNotFound) Error() string {
	return string(e) + " doesn't exist"
}

// ErrExists indicates the thing already exists.
type ErrExists string

// Error implements the error interface.
func (e ErrExists) Error() string {
	return string(e) + " already exists"
}

// ErrNoBucket is the error returned if we trying to reference a bucket that
// doesn't exist.
type ErrNoBucket string

// Error implements the error interface.
func (e ErrNoBucket) Error() string {
	return fmt.Sprintf("bucket %q doesn't exist", string(e))
}
