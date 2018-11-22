// Package main is the entry point for ClayMUD.  This defines the command line
// args and listens for incoming connections.
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"

	"github.com/natefinch/claymud/auth"
	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/db"
	"github.com/natefinch/claymud/game/gender"
	"github.com/natefinch/claymud/game/social"
	"github.com/natefinch/claymud/world"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8888, "specifies the port the server listens on")
}

func main() {
	if err := Main(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

// Main is the main entrypoint to the server
func Main() error {
	flag.Parse()

	// config must be first!
	cfg, err := config.Initialize()
	if err != nil {
		return err
	}
	dir := config.DataDir()
	if err := gender.Initialize(dir); err != nil {
		return err
	}
	if err := social.Initialize(dir); err != nil {
		return err
	}
	if err := auth.Initialize(dir); err != nil {
		return err
	}

	// db must be before world!
	maybeFatal(db.Initialize(dir))

	shutdown := make(chan struct{})
	wg := &sync.WaitGroup{}
	defer func() {
		close(shutdown)
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			log.Print("Timed out waiting for all goroutines to clean up.  Killing process.")
		}
	}()
	lock := &sync.RWMutex{}

	// World needs to be last.
	if err := world.Initialize(dir, lock.RLocker(), shutdown, wg); err != nil {
		return err
	}
	if err := world.SetStart(util.Id(cfg.StartRoom)); err != nil {
		return err
	}
	global := game.SpawnWorker(lock, shutdown, wg)

	host := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	log.Printf("Running ClayMUD on %v", host)

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
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
		go auth.Login(conn, conn.RemoteAddr(), global)
	}
}

func maybeFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
