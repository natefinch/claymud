// Package auth controls authenticating users and passes the connections to the world
// after authentication
package auth

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"src.natemud.org/config"
	"src.natemud.org/util"
	"src.natemud.org/world"
)

var (
	ErrAuth = errors.New("auth: authentication failed")
	ErrDupe = errors.New("auth: duplciate login detected")
)

const (
	retries = 3
)

// logs a player in from an incoming connection, creating a player
// in the world if they successfully connect
func Login(rwc io.ReadWriteCloser, ip net.Addr) {
	rw := bufio.NewReadWriter(bufio.NewReader(rwc), bufio.NewWriter(rwc))
	showTitle(rw)
	for i := 0; i < retries; i++ {
		user, err := authenticate(rw)
		switch err {
		case nil:
			world.SpawnPlayer(rwc, user, ip)
			return

		case ErrAuth:
			log.Printf("Failed login from %s", ip)
			err = util.WriteLn(rw, "Incorrect username or password, please try again")
			continue

		case ErrDupe:
			ok, err := handleDupe(user, rw)
			if ok && err == nil {
				kick(user)
				world.SpawnPlayer(rwc, user, ip)
				return
			}
			continue
		}
		if err != nil {
			log.Printf("Error during login of user from %s: %s", ip, err)
			return
		}
	}
}

func showTitle(rw *bufio.ReadWriter) {
	util.WriteLn(rw, config.MainTitle())
}

// Queries the user for username and password, then authenticates the credentials
func authenticate(rw *bufio.ReadWriter) (user string, err error) {
	err = util.Write(rw, "Username: ")
	if err != nil {
		return
	}

	user, err = util.ReadLn(rw)
	if err != nil {
		return
	}
	// TODO: remove this before production!
	log.Printf("User logging in: %v", user)

	err = util.Write(rw, "Password: ")
	if err != nil {
		return user, util.ErrClosed
	}

	pass, err := util.ReadLn(rw)
	if err != nil {
		return
	}
	err = checkPass(user, pass)
	return
}

func checkPass(user, pass string) error {
	// TODO: actually authenticate
	if pass == "TheBadPassword" {
		return ErrAuth
	}

	if world.FindPlayer(user) != nil {
		return ErrDupe
	}
	return nil
}

func handleDupe(user string, rw *bufio.ReadWriter) (bool, error) {
	// TODO: actually handle duplicate logins
	util.WriteLn(rw, "This account is already logged in.")
	return false, nil
}

func kick(user string) {
	// TODO: implement
}
