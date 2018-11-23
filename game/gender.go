package game

import "log"

// Genders is the list of globally available genders.
var Genders []Gender

// Gender defines a person's pronouns.
type Gender struct {
	Name  string
	Xself string
	Xe    string
	Xim   string
	Xis   string
}

// InitGenders sets up the globally available genders.
func InitGenders(g []Gender) {
	if len(g) == 0 {
		log.Println("WARNING: no genders defined")
		Genders = []Gender{
			{
				Name:  "none",
				Xself: "itself",
				Xe:    "it",
				Xim:   "it",
				Xis:   "its",
			},
		}
	}
	Genders = g
}
