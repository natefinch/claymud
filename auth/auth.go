// Package auth controls authenticating users and passes the connections to the world
// after authentication
package auth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/world"
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
	showTitle(rwc)
	for i := 0; i < retries; i++ {
		user, err := authenticate(rwc)
		switch err {
		case nil:
			world.SpawnPlayer(rwc, user, ip)
			return

		case ErrAuth:
			log.Printf("Failed login from %s", ip)
			_, err = rwc.Write([]byte("Incorrect username or password, please try again\n"))
			if err != nil {
				break
			}
		case ErrDupe:
			ok, err := handleDupe(user, rwc)
			if ok && err == nil {
				kick(user)
				world.SpawnPlayer(rwc, user, ip)
				return
			}
		}
		if err != nil {
			log.Printf("Error during login of user from %s: %s", ip, err)
			return
		}
	}
}

func showTitle(w io.Writer) error {
	_, err := w.Write([]byte(config.MainTitle()))
	return err
}

// Queries the user for username and password, then authenticates the credentials
func authenticate(rw io.ReadWriter) (user string, err error) {
	_, err = rw.Write([]byte("Username: "))
	if err != nil {
		return user, err
	}

	scanner := bufio.NewScanner(rw)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return user, err
		}
		return user, fmt.Errorf("Connection closed")
	}
	user = scanner.Text()

	// TODO: remove this before production!
	log.Printf("User logging in: %v", user)

	_, err = rw.Write([]byte("Password: "))
	if err != nil {
		return user, err
	}

	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return user, err
		}
		return user, fmt.Errorf("Connection closed")
	}
	pass := scanner.Text()
	err = checkPass(user, pass)
	return user, err
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

func handleDupe(user string, w io.Writer) (bool, error) {
	// TODO: actually handle duplicate logins
	_, err := w.Write([]byte("This account is already logged in."))
	return false, err
}

func kick(user string) {
	// TODO: implement
}
