// Package main is the entry point for ClayMUD.  This defines the command line
// args and listens for incoming connections.
package main

import (
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/natefinch/claymud/auth"
	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/db"
	"github.com/natefinch/claymud/game/emote"
	"github.com/natefinch/claymud/game/gender"
	"github.com/natefinch/claymud/world"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8888, "specifies the port the server listens on")
}

func main() {
	Main()
}

func Main() {
	flag.Parse()

	// config must be first!
	maybeFatal(config.Initialize())

	maybeFatal(gender.Initialize(config.DataDir()))
	maybeFatal(emote.Initialize())
	maybeFatal(auth.Initialize())

	// db must be before world!
	maybeFatal(db.Initialize())

	// World needs to be last.
	maybeFatal(world.Initialize())

	host := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	log.Printf("Running ClayMUD on %v", host)

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Printf("Error resolving host %v: %v", host, err)
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Failed listening for connections: %s", err)
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		conn.SetKeepAlive(false)
		conn.SetLinger(0)

		log.Printf("New connection from %v", conn.RemoteAddr())
		go auth.Login(conn, conn.RemoteAddr())
	}
}

func maybeFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
