package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/natefinch/claymud/auth"
	"github.com/natefinch/claymud/db"
	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/game/social"
	"github.com/natefinch/claymud/server/config"
	"github.com/natefinch/claymud/util"
	"github.com/natefinch/claymud/world"
)

// Main is the main entrypoint to the server
func Main() error {
	var port int

	flag.IntVar(&port, "port", 8888, "specifies the port the server listens on")
	flag.Parse()

	// config must be first!
	cfg, err := config.Init()
	if err != nil {
		return err
	}
	dir := cfg.DataDir
	game.InitGenders(cfg.Gender)
	game.InitDirs(cfg.Direction)
	if err := social.Initialize(dir); err != nil {
		return err
	}
	auth.Init(cfg.MainTitle, cfg.BcryptCost)
	// db must be before world!
	if err := db.Initialize(dir); err != nil {
		return err
	}

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
	wc := world.Config{
		Commands:  cfg.Commands,
		StartRoom: cfg.StartRoom,
	}
	wc.ChatMode.Default = cfg.ChatMode.Default
	wc.ChatMode.Prefix = cfg.ChatMode.Prefix
	switch cfg.ChatMode.Enabled {
	case "allow":
		wc.ChatMode.Mode = world.ChatModeAllow
	case "deny":
		wc.ChatMode.Mode = world.ChatModeDeny
	case "require":
		wc.ChatMode.Mode = world.ChatModeRequire
	default:
		// we already checked this, but belt and suspenders is ok
		return fmt.Errorf("Expected allow, deny, or require for ChatMode.Enabled, got %q", cfg.ChatMode.Enabled)
	}

	lock := &sync.RWMutex{}

	// World needs to be last.
	if err := world.Init(wc, dir, lock.RLocker(), shutdown, wg); err != nil {
		return err
	}
	if err := world.SetStart(util.Id(cfg.StartRoom)); err != nil {
		return err
	}
	global := game.SpawnWorker(lock, shutdown, wg)

	host := net.JoinHostPort("", strconv.Itoa(port))
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
