// Package auth controls authenticating users and passes the connections to the world
// after authentication
package auth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/natefinch/claymud/db"
	"github.com/natefinch/claymud/util"
)

var (
	ErrAuth     = errors.New("auth: authentication failed")
	ErrDupe     = errors.New("auth: duplicate login detected")
	ErrExists   = errors.New("auth: username already exists")
	ErrNotSetup = errors.New("auth: mud not set up")

	bcryptCost int
	mainTitle  []byte

	// fakehash is a fake hashed password created with the current bcryptcost.
	// It exists to allow us to fake out password hashing time when a username
	// doesn't exist.
	fakehash []byte
)

const (
	retries = 3
)

// Init sets the bcryptcost for hashing passwords and sets up authentication.
func Init(title string, cost int) {
	mainTitle = []byte(title)
	bcryptCost = cost
	log.Printf("Using bcrypt cost %d", bcryptCost)

	var err error
	fakehash, err = bcrypt.GenerateFromPassword([]byte("password"), bcryptCost)
	if err != nil {
		panic(err)
	}
}

// Login logs a user in from an incoming connection, creating a player
// in the world if they successfully connect
func Login(st *db.Store, rwc io.ReadWriteCloser, ip net.Addr) (*User, error) {
	if err := showTitle(rwc); err != nil {
		return nil, err
	}
	var ws util.WriteScanner = struct {
		io.Writer
		*bufio.Scanner
	}{
		Writer:  rwc,
		Scanner: bufio.NewScanner(rwc),
	}
	for i := 0; i < retries; i++ {
		user, err := authenticate(st, ws, ip)
		switch err {
		case nil:
			user.WriteScanner = ws
			user.Closer = rwc
			return user, nil
		case ErrAuth:
			log.Printf("Failed login from %s", ip)
			_, err := io.WriteString(rwc, "Incorrect username or password, please try again\n")
			if err != nil {
				return nil, err
			}
			continue
		case ErrDupe:
			_, err = io.WriteString(rwc, "This account is already logged in.\n")
			if err != nil {
				return nil, err
			}
			continue
		case ErrNotSetup:
			_ = rwc.Close()
			return nil, ErrNotSetup
		default:
			log.Printf("Failed to log in user: %v", err)
			return nil, err
		}
	}
	io.WriteString(rwc, "Too many failures, good bye.\n")
	rwc.Close()
	return nil, ErrAuth
}

func showTitle(w io.Writer) error {
	_, err := w.Write(mainTitle)
	return err
}

// authenticate queries the user for username and password, then authenticates
// the credentials.
func authenticate(st *db.Store, ws util.WriteScanner, ip net.Addr) (*User, error) {
	setup, err := st.IsSetup()
	if err != nil {
		return nil, fmt.Errorf("can't authenticate: %s", err)
	}
	// first time anyone has logged in.
	if !setup {
		host, _, err := net.SplitHostPort(ip.String())
		if err != nil {
			return nil, err
		}
		if host != "127.0.0.1" && host != "::1" {
			showNotSetup(ws)
			return nil, ErrNotSetup
		}
		if err := showIntro(ws); err != nil {
			return nil, err
		}
		user, err := showCreate(st, ws, ip)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	// normal case
	a, err := util.QueryOptions(ws, "", 'l',
		util.Opt{Key: 'c', Text: "Create account"},
		util.Opt{Key: 'l', Text: "Log in with existing account"})
	if err != nil {
		return nil, err
	}

	switch a {
	case 'c':
		return showCreate(st, ws, ip)
	case 'l':
		u, p, err := queryCreds(ws)
		if err != nil {
			return nil, err
		}
		return checkPass(st, u, p, ip)
	default:
		panic(fmt.Errorf("Should be impossible, got %v from login options", a))
	}
}

func showIntro(w io.Writer) error {
	_, err := fmt.Fprintln(w, `
Greetings, Administrator.  Welcome to ClayMUD.

Since you are the first one here, you hold all the keys.  You will be asked to
create an account, this account will be the first administrator account (you
can make other people adminstrators later).  

Do not forget your password.  There is no password reset feature (yet).`)
	return err
}

func showNotSetup(w io.Writer) {
	fmt.Fprintln(w, `
Greetings, User.  Welcome to ClayMUD.

This instance of ClayMUD has not been set up and is not ready for public
consumption.  If you are the administrator of this MUD, please connect from the
machine where the MUD runs to start setup.`)
}

// showCreate leads the user through the process of creating a user.
func showCreate(st *db.Store, ws util.WriteScanner, ip net.Addr) (*User, error) {
	_, err := io.WriteString(ws, `
Please enter a username.  Note that this is only for use in logging into the MUD
and will not be visible to non-admins.

`)
	if err != nil {
		return nil, err
	}

	for {
		u, pw, err := queryNewUser(st, ws)
		if err != nil {
			return nil, err
		}
		user, err := createDBUser(st, u, pw, ip)
		if err == ErrExists {
			_, err := io.WriteString(ws, "That username already exists, please choose another.\n")
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

// createDBUser creates the user in the DB if it does not exist.  If it does exist,
// createDBUser will return ErrExists.
func createDBUser(st *db.Store, username, pw string, ip net.Addr) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	if err != nil {
		return nil, err
	}
	setup, err := st.IsSetup()
	if err != nil {
		return nil, err
	}
	user := &User{
		Username: username,
		IP:       ip,
		bits:     big.NewInt(0),
	}
	if !setup {
		// the first person in gets to be admin
		user.SetFlag(UFlagAdmin)
	}
	doc := db.User{
		Username:  username,
		LastIP:    ip.String(),
		LastLogin: time.Now(),
		Flags:     user.bits,
	}
	if err := st.CreateUser(doc, hash); err != nil {
		if _, ok := err.(db.ErrExists); ok {
			return nil, ErrExists
		}
		return nil, err
	}
	log.Printf("created user %q", username)
	return user, nil
}

// queryCreds asks the user for their username and password.
func queryCreds(ws util.WriteScanner) (user, pwd string, err error) {
	user, err = util.Query(ws, "Username: ")
	if err != nil {
		return "", "", err
	}
	pwd, err = util.Query(ws, "Password: ")
	if err != nil {
		return "", "", err
	}
	return user, pwd, nil
}

// queryNewUser asks the user to create a new username and password.
func queryNewUser(st *db.Store, ws util.WriteScanner) (user, pwd string, err error) {
	user, err = util.QueryVerify(ws, "Username: ",
		func(user string) (string, error) {
			exists, err := st.UserExists(user)
			if err != nil {
				return "", fmt.Errorf("error checking for existence of username: %s", err)
			}
			if !exists {
				return "", nil
			}
			return "That username already exists, please choose another.", nil
		})
	if err != nil {
		return "", "", err
	}
	pwd, err = util.QueryVerify(ws, "Password: ",
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

// checkPass verifies that the given user exists and that the password matches.
func checkPass(st *db.Store, username, pass string, ip net.Addr) (*User, error) {
	passb := []byte(pass)
	c, err := st.FindCreds(username)
	if _, ok := err.(db.ErrNotFound); ok {
		// User does not exist. Fake out the time we would otherwise take to run
		// the hash.  Ignore the error, we really only care about sucking up
		// some CPU cycles here.
		_ = bcrypt.CompareHashAndPassword(fakehash, passb)
		return nil, ErrAuth
	}
	start := time.Now()
	err = bcrypt.CompareHashAndPassword(c.PwdHash, passb)
	log.Printf("user password hashed in %v", time.Since(start))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, ErrAuth
	}
	if err != nil {
		return nil, err
	}

	cost, err := bcrypt.Cost(c.PwdHash)
	if err != nil {
		return nil, err
	}

	// Handle bcrypt cost change, rehash with new cost.
	if cost != bcryptCost {
		hash, err := bcrypt.GenerateFromPassword(passb, bcryptCost)
		if err != nil {
			return nil, err
		}
		c.PwdHash = hash
		if err := st.SaveCreds(c); err != nil {
			return nil, err
		}
	}
	// Login successful, update info.
	u, err := st.FindUser(username)
	if err != nil {
		return nil, err
	}
	u.LastIP = ip.String()
	u.LastLogin = time.Now()
	if err := st.SaveUser(u); err != nil {
		return nil, err
	}
	user := &User{
		Username: username,
		IP:       ip,
		Players:  u.Players,
		bits:     u.Flags,
	}
	if user.bits == nil {
		user.bits = big.NewInt(0)
	}
	return user, nil
}
