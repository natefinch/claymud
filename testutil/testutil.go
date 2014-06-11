package testutil

import (
	"os"
)

func PatchEnv(key, value string) func() {
	orig := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		os.Setenv(key, orig)
	}
}
