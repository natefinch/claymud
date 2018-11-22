package world

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/natefinch/claymud/config"
	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var (
	locMap = map[util.Id]*Location{}
	start  *Location
	zones  = map[util.Id]*Zone{}
)

// SetStart sets the starting room of the mud.
func SetStart(room util.Id) error {
	loc, exists := locMap[room]
	if !exists {
		return fmt.Errorf("starting room %v does not exist", room)
	}
	start = loc
	return nil
}

// The start room of the MUD, where players appear
// TODO: multiple / configurable start rooms
func Start() *Location {
	return start
}

func genWorld(datadir string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) error {
	log.Printf("loading zones from %v", filepath.Join(datadir, "zones"))
	files, err := filepath.Glob(filepath.Join(datadir, "zones", "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read zone files: %v", err)
	}
	for _, file := range files {
		zone, err := decodeZone(file)
		if err != nil {
			return err
		}
		if z, exists := zones[zone.ID]; exists {
			return fmt.Errorf("file %q contains %s which duplicates zone %s", file, zone, z)
		}
		zone.Add(
			&Area{
				ID:      <-ids,
				Name:    zone.Name,
				LocByID: map[util.Id]*Location{},
			})
		zones[zone.ID] = zone
		zone.Worker = game.SpawnWorker(zoneLock, shutdown, wg)
	}
	log.Printf("loaded %v zones", len(files))

	log.Printf("loading rooms from %v", filepath.Join(datadir, "rooms"))
	files, err = filepath.Glob(filepath.Join(datadir, "rooms", "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read room files: %v", err)
	}
	log.Printf("found %d room files", len(files))
	var jsonRooms []jsonRoom
	count := 0
	for _, file := range files {
		log.Printf("decoding rooms in file %v", file)
		jrs, err := decodeRooms(locMap, file)
		if err != nil {
			return err
		}
		count += len(jrs)
		if len(jrs) > 0 {
			log.Printf("found %d rooms for zone %v", len(jrs), jrs[0].Zone)
		}
		jsonRooms = append(jsonRooms, jrs...)
	}
	log.Printf("found %v rooms", count)

	// ok, now that we've loaded all the room definitions, we have to go back
	// and hook up all the exits. We have to do this afterward because an exit
	// might refer to a location that hasn't been loaded yet.
	for _, r := range jsonRooms {
		loc, exists := locMap[util.Id(r.ID)]
		if !exists {
			return fmt.Errorf("should be impossible, jsonRoom refers to unknown location %v", r.ID)
		}
		loc.Exits = make(Exits, len(r.Exits))
		for i, e := range r.Exits {
			if e.Destination == -1 {
				continue
			}
			dir, exists := config.FindDirection(e.Direction)
			if !exists {
				return fmt.Errorf("should be impossible, direction %q for exit in room %v doesn't exist", e.Direction, r.ID)
			}
			target, exists := locMap[util.Id(e.Destination)]
			if !exists {
				return fmt.Errorf("direction %q for exit in room %v references non-existant room %v", e.Direction, r.ID, e.Destination)
			}
			loc.Exits[i] = Exit{
				Direction:   dir,
				Desc:        e.Description,
				Destination: target,
			}
		}
	}
	return nil
}

func dir(d string) config.Direction {
	dir, ok := config.FindDirection(d)
	if !ok {
		panic(fmt.Errorf("Can't find direction %s", d))
	}
	return dir
}

func decodeZone(file string) (*Zone, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't open zone file: %v", err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var zone Zone
	if err := d.Decode(&zone); err != nil {
		return nil, fmt.Errorf("unable to decode zone file %q: %v", file, err)
	}
	return &zone, nil
}

func decodeRooms(rooms map[util.Id]*Location, file string) ([]jsonRoom, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't open room file: %v", err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var decoded struct {
		Rooms []jsonRoom `json:"rooms"`
	}
	if err := d.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("unable to decode room file %q: %v", file, err)
	}
	for _, r := range decoded.Rooms {
		if rm, exists := rooms[util.Id(r.ID)]; exists {
			return nil, fmt.Errorf("room %v (%s) already exists as %q", r.ID, r.Name, rm.Name)
		}
		loc, err := r.toLoc()
		if err != nil {
			return nil, err
		}
		rooms[loc.ID] = loc
	}
	return decoded.Rooms, nil
}

// jsonRoom is a representation of a room in a MUD.
type jsonRoom struct {
	ID          int             `json:"ID"`
	Zone        int             `json:"Zone"`
	Name        string          `json:"Name"`
	Description string          `json:"Description"`
	Bits        []string        `json:"Bits"`
	Sector      string          `json:"Sector"`
	Exits       []jsonExit      `json:"Exits"`
	Extras      []jsonExtraDesc `json:"ExtraDescs"`
}

func (j jsonRoom) toLoc() (*Location, error) {
	z, exists := zones[util.Id(j.Zone)]
	if !exists {
		return nil, fmt.Errorf("room %v's zone %v does not exist", j.ID, j.Zone)
	}
	for _, e := range j.Exits {
		if _, exists := config.FindDirection(e.Direction); !exists {
			return nil, fmt.Errorf("direction %q for exit in room %v does not exist", e.Direction, j.ID)
		}
	}
	loc := &Location{
		ID:           util.Id(j.ID),
		Name:         j.Name,
		Desc:         j.Description,
		Descriptions: map[string]string{},
		Players:      map[string]*Player{},
	}
	for _, e := range j.Extras {
		for _, k := range e.Keywords {
			loc.Descriptions[k] = e.Description
		}
	}
	z.Areas[0].Add(loc)
	return loc, nil
}

// jsonExit represents a way you may move out of a room.
type jsonExit struct {
	Direction   string   `json:"Direction"`
	Description string   `json:"Description"`
	Keywords    []string `json:"Keywords"`
	DoorFlags   []string `json:"DoorFlags"`
	KeyNumber   int      `json:"KeyID"`
	Destination int      `json:"Destination"`
}

// jsonExtraDesc represents other things you can look at in the room.
type jsonExtraDesc struct {
	Keywords    []string `json:"Keywords"`
	Description string   `json:"Description"`
}
