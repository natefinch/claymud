// Package world holds most of the MUD code, including locations, players, etc
package world

import (
	"github.com/natefinch/natemud/util"
	"log"
	"strings"
)

// TODO: make all this less terrible

var (
	adds    = make(chan *Player)
	deletes = make(chan *Player)
	finds   = make(chan string)
	findRes = make(chan *Player)
)

func AddPlayer(p *Player) {
	adds <- p
}

func RemovePlayer(p *Player) {
	deletes <- p
}

func FindPlayer(name string) (p *Player) {
	finds <- name
	p = <-findRes
	return
}

func Initialize() {
	go playerListLoop()
	genWorld()
}

// This method runs as a goroutine that communicates via
// channels with the rest of the mud.  This is our global
// list of players, and by using channels to communicate,
// we synchronize access to it
func playerListLoop() {
	log.Print("Initializing player list loop")

	playerIds := make(map[util.Id]*Player)

	// note that all username lookups are case insensitive
	playerNames := make(map[string]*Player)
	for {
		select {
		case p := <-adds:
			playerIds[p.Id()] = p
			playerNames[strings.ToLower(p.Name())] = p
		case p := <-deletes:
			delete(playerIds, p.Id())
			delete(playerNames, strings.ToLower(p.Name()))
		case name := <-finds:
			findRes <- playerNames[strings.ToLower(name)]
		}
	}
}
