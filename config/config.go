package config

import (
	"os"
)

const (
	NATEMUD_DATADIR = "NATEMUD_DATADIR"
)

var (
	DataDir = "data"
)

func init() {
	override(&DataDir, NATEMUD_DATADIR)
}

// override will set the value of val to the environment variable value of env
// if the environment variable is set
func override(val *string, env string) {
	if v := os.Getenv(env); v != "" {
		*val = v
	}
}
