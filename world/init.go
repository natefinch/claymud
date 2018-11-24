// Package world holds most of the MUD code, including locations, players, etc
package world

import (
	"sync"
)

// ChatModeMode determines whether ChatMode is allowed to be on, required to be on, or not allowed to be on.
type ChatModeMode int

// Allow, disallow, or require chatmode.
const (
	ChatModeAllow ChatModeMode = iota
	ChatModeDeny
	ChatModeRequire
)

// ChatMode defines if and how users enter chatmode.
type ChatMode struct {
	Default bool   // Whether players default to being in chat mode
	Prefix  string // prefix required for commands in chatmode
	Mode    ChatModeMode
}

// Config determines the configuration of the world.
type Config struct {
	StartRoom int       // ID of room players start in
	Commands  CmdConfig // command names
	ChatMode  ChatMode
}

// Init spawns the zones and their attendant workers, creates all areas
// and locations.
func Init(cfg Config, datadir string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) error {
	if err := loadLocTempl(datadir); err != nil {
		return err
	}

	chatMode = cfg.ChatMode

	// ensure that require or deny have the corresponding on or off default
	switch cfg.ChatMode.Mode {
	case ChatModeRequire:
		chatMode.Default = true
	case ChatModeDeny:
		chatMode.Default = false
	default:
		// whatever the config set is fine.
	}
	initCommands(cfg.Commands)
	return loadWorld(datadir, zoneLock, shutdown, wg)
}
