// Package main is the entry point for NateMud.  This defines the command line args
// and listens for incoming connections
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"

	"github.com/natefinch/natemud/auth"
	"github.com/natefinch/natemud/config"
	"github.com/natefinch/natemud/db"
	"github.com/natefinch/natemud/game/emote"
	"github.com/natefinch/natemud/game/gender"
	"github.com/natefinch/natemud/world"
)

var port int

func init() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	flag.IntVar(&port, "port", 8888, "specifies the port the server listens on")
}

func main() {
	flag.Parse()

	// config must be first!
	maybeFatal(config.Initialize())

	maybeFatal(gender.Initialize())
	maybeFatal(emote.Initialize())
	maybeFatal(auth.Initialize())

	// db must be before world!
	maybeFatal(db.Initialize())

	// World needs to be last.
	maybeFatal(world.Initialize())

	host := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	log.Printf("Running NateMUD on %v", host)

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
