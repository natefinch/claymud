// Package auth controls authenticating users and passes the connections to the world
// after authentication
package auth

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"

	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/util"
	"github.com/natefinch/natemud/world"
)

var (
	ErrAuth     = errors.New("auth: authentication failed")
	ErrDupe     = errors.New("auth: duplicate login detected")
	ErrExists   = errors.New("auth: username already exists")
	ErrNotSetup = errors.New("auth: mud not set up")

	users      = []byte("users")
	bcryptCost int
)

const (
	retries = 3
)

func Initialize() error {
	filename := filepath.Join(config.DataDir(), "auth.toml")

	cfg := struct {
		BcryptCost int
	}{}

	res, err := toml.DecodeFile(filename, &cfg)
	if err != nil {
		return fmt.Errorf("Error parsing auth config file %q: %s", filename, err)
	}

	bcryptCost = cfg.BcryptCost

	log.Printf("Auth config loaded.  Using bcrypt cost %d", bcryptCost)

	if und := res.Undecoded(); len(und) > 0 {
		log.Printf("WARNING: Unknown values in auth config file: %v", und)
	}
	return nil
}

// logs a player in from an incoming connection, creating a player
// in the world if they successfully connect
func Login(rwc io.ReadWriteCloser, ip net.Addr) {
	showTitle(rwc)
	for i := 0; i < retries; i++ {
		user, err := authenticate(rwc, ip)
		switch err {
		case nil:
			world.SpawnPlayer(rwc, user)
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
				world.SpawnPlayer(rwc, user)
				return
			}
		case ErrNotSetup:
			rwc.Close()
			return
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

// authenticate queries the user for username and password, then authenticates
// the credentials.
func authenticate(rw io.ReadWriter, ip net.Addr) (user *world.User, err error) {
	first, err := checkFirstSetup()
	if err != nil {
		return nil, fmt.Errorf("can't authenticate: %s", err)
	}
	// first time anyone has logged in.
	if first {
		host, _, err := net.SplitHostPort(ip.String())
		if err != nil {
			return nil, err
		}
		if host != "127.0.0.1" {
			showNotSetup(rw)
			return nil, ErrNotSetup
		}
		if err := showIntro(rw); err != nil {
			return nil, err
		}
		return showCreate(rw, ip)
	}

	// normal case
	a, err := util.QueryOptions(rw, nil,
		util.Opt{'c', []byte("Create account")},
		util.Opt{'l', []byte("Log in with existing account")})
	if err != nil {
		return nil, err
	}

	switch a {
	case 'c':
		return showCreate(rw, ip)
	case 'l':
		u, p, err := getUserPwd(rw)
		if err != nil {
			return nil, err
		}
		return checkPass(u, p, ip)
	default:
		panic(fmt.Errorf("Should be impossible, got %v from login options", a))
	}
}

func checkFirstSetup() (first bool, err error) {
	err = config.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			// bucket doesn't exist
			first = true
			return nil
		}
		k, v := b.Cursor().First()
		// bucket empty
		first = k == nil && v == nil
		return nil
	})
	return first, err
}

func showIntro(rw io.ReadWriter) error {
	_, err := fmt.Fprintln(rw, `
Greetings, Administrator.  Welcome to NateMUD.

Since you are the first one here, you hold all the keys.  You will be asked to
create an account, this account will be the first administrator account (you
can make other people adminstrators later).  

Do not forget your password.  There is no password reset feature (yet).`)
	return err
}

func showNotSetup(rw io.ReadWriter) {
	fmt.Fprintln(rw, `
Greetings, User.  Welcome to NateMUD.

This instance of NateMUD has not been set up and is not ready for public
consumption.  If you are the administrator of this MUD, please connect from the
machine where the MUD runs to start setup.`)
}

// showCreate leads the user through the process of creating a user.
func showCreate(rw io.ReadWriter, ip net.Addr) (user *world.User, err error) {
	_, err = fmt.Fprintln(rw, `
Please enter a username.  Note that this is only for use in logging into the MUD
and will not be visible to non-admins.
`)
	if err != nil {
		return nil, err
	}

	for {
		u, pw, err := newUserPwd(rw)
		if err != nil {
			return nil, err
		}
		user, err = createUser(u, pw, ip)
		if err == ErrExists {
			_, err := fmt.Fprintln(rw, "That username already exists, please choose another.")
			if err != nil {
				return nil, err
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		return user, nil
	}
}

// getUserPwd asks the user for their username and password.
func getUserPwd(rw io.ReadWriter) (user, pwd string, err error) {
	user, err = util.Query(rw, []byte("Username: "))
	if err != nil {
		return "", "", err
	}
	pwd, err = util.Query(rw, []byte("Password: "))
	if err != nil {
		return "", "", err
	}
	return user, pwd, nil
}

// newUserPwd asks the user to create a new username and password.
func newUserPwd(rw io.ReadWriter) (user, pwd string, err error) {
	user, err = util.QueryVerify(rw, []byte("Username: "),
		func(user string) (string, error) {
			ex, err := exists(user)
			if err != nil {
				return "", fmt.Errorf("error checking for existence of username: %s", err)
			}
			if !ex {
				return "", nil
			}
			return "That username already exists, please choose another.", nil
		})
	if err != nil {
		return "", "", err
	}
	pwd, err = util.QueryVerify(rw, []byte("Password: "),
		func(pw string) (string, error) {
			if len(pw) > 1024 {
				return "The maximum length for a password is 1024 characters.", nil
			}
			return "", nil
		})
	if err != nil {
		return "", "", err
	}
	return user, pwd, nil
}

// userDoc is the structure that is stored in the database for a User.
type userDoc struct {
	PwdHash   []byte
	LastIP    net.Addr
	LastLogin time.Time
}

// createUser creates the user account in the db.
func createUser(username, pw string, ip net.Addr) (user *world.User, err error) {
	// do the expensive hash outside the critical section!!
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	if err != nil {
		return nil, err
	}
	err = config.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(users)
		if err != nil {
			return err
		}

		// make sure the username doesn't exist.  Since we're in a write
		// transaction, we have exclusive access, so don't need to worry about
		// race conditions.
		if b.Get([]byte(username)) != nil {
			return ErrExists
		}

		u := userDoc{
			PwdHash:   hash,
			LastIP:    ip,
			LastLogin: time.Now(),
		}

		return util.Put(b, []byte(username), u)
	})
	if err != nil {
		return nil, err
	}
	user = &world.User{
		Username: username,
	}
	return user, err
}

// exists reports whether the username exists.
func exists(username string) (bool, error) {
	exists := false
	err := config.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		if b == nil {
			return nil
		}
		exists = b.Get([]byte(username)) != nil
		return nil
	})
	return exists, err
}

// checkPass verifies that the given user exists and that the password matches.
func checkPass(username, pass string, ip net.Addr) (user *world.User, err error) {
	if world.FindPlayer(username) != nil {
		return nil, ErrDupe
	}
	var hash []byte
	err = config.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(users)
		var u userDoc
		exists, err := util.Get(b, []byte(username), &u)
		if err != nil {
			return err
		}
		if !exists {
			return ErrAuth
		}
		hash = u.PwdHash
		return nil
	})
	if err == ErrAuth {
		// fake out the time we would otherwise take to run the hash
		bcrypt.CompareHashAndPassword([]byte("somefakehash"), []byte(pass))
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	// TODO: handle bcryptcost changes

	err = bcrypt.CompareHashAndPassword(hash, []byte(pass))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, ErrAuth
	}
	if err != nil {
		return nil, err
	}

	// login successful, update info
	err = config.DB.Update(func(tx *bolt.Tx) error {
		u := userDoc{
			PwdHash:   hash,
			LastIP:    ip,
			LastLogin: time.Now(),
		}

		return util.Put(tx.Bucket(users), []byte(username), u)
	})

	user = &world.User{Username: username, IP: ip}
	return user, nil
}

func handleDupe(user *world.User, w io.Writer) (bool, error) {
	// TODO: actually handle duplicate logins
	_, err := w.Write([]byte("This account is already logged in."))
	return false, err
}

func kick(user *world.User) {
	// TODO: implement
}
