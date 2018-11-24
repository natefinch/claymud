//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating offial releases and is really only
// useful to project admins.
package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Runs "go build" and generates the version info the binary.
func Build() error {
	return sh.RunV("go", "build", "-ldflags="+flags(), "github.com/natefinch/claymud")
}

// Runs the binary after building.
func Run() error {
	mg.Deps(Build)
	return sh.RunV("./claymud")
}

var releaseTag = regexp.MustCompile(`^v1\.[0-9]+\.[0-9]+$`)

// Generates a new release.  Expects the TAG environment variable to be set,
// which will create a new tag with that name.
func Release() (err error) {
	tag := os.Getenv("TAG")
	if !releaseTag.MatchString(tag) {
		return errors.New("TAG environment variable must be in semver v1.x.x format, but was " + tag)
	}

	if err := sh.RunV("git", "tag", "-a", tag, "-m", tag); err != nil {
		return err
	}
	if err := sh.RunV("git", "push", "origin", tag); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			sh.RunV("git", "tag", "--delete", "$TAG")
			sh.RunV("git", "push", "--delete", "origin", "$TAG")
		}
	}()
	return sh.RunV("goreleaser")
}

// Remove the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	hash := hash()
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	return fmt.Sprintf(`-X "github.com/natefinch/claymud/server.timestamp=%s" -X "github.com/natefinch/claymud/server.commitHash=%s" -X "github.com/natefinch/claymud/server.gitTag=%s"`, timestamp, hash, tag)
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	return s
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}
