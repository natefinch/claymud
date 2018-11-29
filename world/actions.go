package world

import (
	"io/ioutil"
	"path/filepath"

	"github.com/hippogryph/skyhook"
	"go.starlark.net/starlark"
)

var allActions = map[string]*starlark.Program{}

var actionDir string

// An action is a script that can be run.
type Action struct {
	Filename string
	IsGlobal bool
}

func InitActions(dir string) error {
	actionDir = dir
	// files, err := ioutil.ReadDir(dir)
	// if err != nil {
	// 	return fmt.Errorf("error reading script directory: %v", err)
	// }
	// for _, f := range files {
	// 	if f.IsDir() {
	// 		continue
	// 	}
	// 	start := time.Now()
	// 	predec := func(s string) bool {
	// 		return s == "echo"
	// 	}
	// 	filename := filepath.Join(dir, f.Name())
	// 	log.Printf("loading action script %v", filename)
	// 	_, p, err := starlark.SourceProgram(filename, nil, predec)
	// 	if err != nil {
	// 		return fmt.Errorf("error parsing %q: %v", filename, err)
	// 	}
	// 	log.Println("time to parse a skylark file:", time.Since(start))
	// 	allActions[f.Name()] = p
	// }
	return nil
}

func runLocAction(name string, actor *Player, loc *Location) error {
	b, err := ioutil.ReadFile(filepath.Join(actionDir, name))
	if err != nil {
		return err
	}
	dict := map[string]interface{}{
		"echo": func(msg string) {
			echo(loc, msg)
		},
		"around":   around,
		"actor":    actor,
		"location": loc,
	}
	_, err = skyhook.Eval(b, dict)
	return err
}

// say something to all the people in the room
func echo(loc *Location, msg string) {
	for _, p := range loc.Players {
		p.WriteString(msg + "\n")
	}
}

func around(player *Player, msg string) {
	for _, p := range player.loc.Players {
		if p.ID != player.ID {
			p.WriteString(msg + "\n")
		}
	}
}
