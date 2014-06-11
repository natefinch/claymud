package config

import (
	"os"
)

const (
	NATEMUD_DATADIR = "NATEMUD_DATADIR"
)

func DataDir() string {
	v := os.Getenv(NATEMUD_DATADIR)
	if v == "" {
		return "data"
	}
	return v
}
