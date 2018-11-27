package world

import (
	"github.com/natefinch/claymud/game"
	"github.com/natefinch/claymud/util"
)

type Mob struct {
	ID           util.ID
	Aliases      []string
	Name         string
	LongDesc     string
	DetailedDesc string
	//Actions         *big.Int
	//Affections      *big.Int
	Alignment int
	Level     int
	THAC0     int
	AC        int
	HP        game.Dice // xdy+z
	Damage    game.Dice // xdy+z
	Gold      int
	XP        int
	//LoadPosition    game.Position
	//DefaultPosition game.Position
	Gender game.Gender
}
