package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
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

// set by ldflags when you "mage build"
var (
	commitHash = "<not set>"
	timestamp  = "<not set>"
	gitTag     = "<not set>"
)

// Main is the main entrypoint to the server
func Main() error {
	var port int
	var version bool
	flag.IntVar(&port, "port", 8888, "specifies the port the server listens on")
	flag.BoolVar(&version, "version", false, "show version info")
	flag.Parse()

	if version {
		fmt.Println("ClayMUD", gitTag)
		fmt.Println("Build Date:", timestamp)
		fmt.Println("Commit:", commitHash)
		fmt.Println("built with:", runtime.Version())
		return nil
	}

	// config must be first!
	cfg, err := config.Init()
	if err != nil {
		return err
	}
	log.Println("ClayMUD", gitTag)
	log.Println("Build Date:", timestamp)
	log.Println("Commit:", commitHash)
	log.Println("built with:", runtime.Version())

	dir := cfg.DataDir
	game.InitGenders(cfg.Gender)
	game.InitDirs(cfg.Direction)
	if err := social.Initialize(dir); err != nil {
		return err
	}
	auth.Init(cfg.MainTitle, cfg.BcryptCost)

	// db must be before world!
	st, err := db.Init(dir)
	if err != nil {
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
	if err := world.SetStart(util.ID(cfg.StartRoom)); err != nil {
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

		go func() {
			log.Printf("New connection from %v", conn.RemoteAddr())
			user, err := auth.Login(st, conn, conn.RemoteAddr())
			if err != nil {
				log.Printf("error logging in user: ")
			}
			if err := world.SpawnPlayer(st, user, global); err != nil {
				log.Printf("error during spawn player for user %s: %s", user.Username, err)
			}
		}()
	}
}
