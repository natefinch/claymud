package world

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

var (
	allZones = map[util.ID]*Zone{}
	allMobs  = map[util.ID]*Mob{}
)

func loadWorld(datadir string, zoneLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) error {
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
		if z, exists := allZones[zone.ID]; exists {
			return fmt.Errorf("file %q contains %s which duplicates zone %s", file, zone, z)
		}
		zone.Add(
			&Area{
				Name:    zone.Name,
				LocByID: map[util.ID]*Location{},
			})
		allZones[zone.ID] = zone
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
		jrs, err := decodeRooms(locMap, file)
		if err != nil {
			return err
		}
		count += len(jrs)
		jsonRooms = append(jsonRooms, jrs...)
	}
	log.Printf("loaded %v rooms", count)

	// ok, now that we've loaded all the room definitions, we have to go back
	// and hook up all the exits. We have to do this afterward because an exit
	// might refer to a location that hasn't been loaded yet.
	for _, r := range jsonRooms {
		loc, exists := locMap[util.ID(r.ID)]
		if !exists {
			return fmt.Errorf("should be impossible, jsonRoom refers to unknown location %v", r.ID)
		}
		loc.Exits = make(Exits, len(r.Exits))
		for i, e := range r.Exits {
			if e.Destination == -1 {
				continue
			}
			dir, exists := game.FindDirection(e.Direction)
			if !exists {
				return fmt.Errorf("should be impossible, direction %q for exit in room %v doesn't exist", e.Direction, r.ID)
			}
			target, exists := locMap[util.ID(e.Destination)]
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

	log.Printf("loading mobs from %v", filepath.Join(datadir, "mobs"))
	files, err = filepath.Glob(filepath.Join(datadir, "mobs", "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read mob files: %v", err)
	}
	log.Printf("found %d mob files", len(files))
	count = 0
	for _, file := range files {
		c, err := decodeMobs(file)
		if err != nil {
			return err
		}
		count += c
	}
	log.Printf("loaded %v mobs", count)

	return nil
}

func dir(d string) game.Direction {
	dir, ok := game.FindDirection(d)
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

func decodeRooms(rooms map[util.ID]*Location, file string) ([]jsonRoom, error) {
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
		if rm, exists := rooms[util.ID(r.ID)]; exists {
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
	Actions     map[string]jsonAction
}

type jsonAction struct {
	Filename string
	IsGlobal bool
}

func (j jsonRoom) toLoc() (*Location, error) {
	z, exists := allZones[util.ID(j.Zone)]
	if !exists {
		return nil, fmt.Errorf("room %v's zone %v does not exist", j.ID, j.Zone)
	}
	for _, e := range j.Exits {
		if _, exists := game.FindDirection(e.Direction); !exists {
			return nil, fmt.Errorf("direction %q for exit in room %v does not exist", e.Direction, j.ID)
		}
	}
	loc := &Location{
		ID:           util.ID(j.ID),
		Name:         j.Name,
		Desc:         j.Description,
		Descriptions: map[string]string{},
		Players:      map[string]*Player{},
		Actions:      map[string]Action{},
	}
	for k, v := range j.Actions {
		loc.Actions[k] = Action(v)
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

type jsonMob struct {
	Number          int
	Aliases         []string
	ShortDesc       string
	LongDesc        string
	DetailedDesc    string
	Actions         []string
	Affections      []string
	Alignment       int
	Level           int
	THAC0           int
	AC              int
	HP              string // xdy+z
	Damage          string // xdy+z
	Gold            int
	XP              int
	LoadPosition    string
	DefaultPosition string
	Gender          string
}

func (m jsonMob) ToMob() (*Mob, error) {
	hp, err := game.MakeDice(m.HP)
	if err != nil {
		return nil, err
	}
	dmg, err := game.MakeDice(m.Damage)
	if err != nil {
		return nil, err
	}
	gen := strings.ToLower(m.Gender)
	if gen == "neutral" {
		gen = "none"
	}
	var gender *game.Gender
	for i := range game.Genders {
		if gen == strings.ToLower(game.Genders[i].Name) {
			gender = &game.Genders[i]
			break

		}
	}
	if gender == nil {
		return nil, fmt.Errorf("unknown gender %v", m.Gender)
	}
	return &Mob{
		ID:           util.ID(m.Number),
		Aliases:      m.Aliases,
		Name:         m.ShortDesc,
		LongDesc:     m.LongDesc,
		DetailedDesc: m.DetailedDesc,
		Alignment:    m.Alignment,
		Level:        m.Level,
		THAC0:        m.THAC0,
		AC:           m.AC,
		HP:           hp,
		Damage:       dmg,
		Gold:         m.Gold,
		XP:           m.XP,
		Gender:       *gender,
	}, nil
}

func decodeMobs(file string) (int, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("can't open room file: %v", err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var decoded struct {
		Mobs []jsonMob `json:"mobs"`
	}
	if err := d.Decode(&decoded); err != nil {
		return 0, fmt.Errorf("unable to decode room file %q: %v", file, err)
	}
	for _, m := range decoded.Mobs {
		if mb, exists := allMobs[util.ID(m.Number)]; exists {
			return 0, fmt.Errorf("mob %v (%s) already exists as %q", m.Number, m.ShortDesc, mb.Name)
		}
		mb, err := m.ToMob()
		if err != nil {
			return 0, err
		}
		allMobs[mb.ID] = mb
	}
	return len(decoded.Mobs), nil
}
