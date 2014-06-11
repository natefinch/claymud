package gender

type Gender int

const (
	None Gender = iota
	Male
	Female
)

type vals struct {
	Male   string
	Female string
	None   string
}

type configuration struct {
	Xself vals
	Xe    vals
	Xim   vals
	Xis   vals
}

var names configuration

// Xself returns the reflexive pronoun for the gender (e.g. himself)
func (s Gender) Xself() string {
	switch s {
	case Male:
		return names.Xself.Male
	case Female:
		return names.Xself.Female
	default:
		return names.Xself.None
	}
}

// Xe returns the pronoun for the gender (e.g. he)
func (s Gender) Xe() string {
	switch s {
	case Male:
		return names.Xe.Male
	case Female:
		return names.Xe.Female
	default:
		return names.Xe.None
	}
}

// Xim returns the subject pronoun for the gender (e.g. him)
func (s Gender) Xim() string {
	switch s {
	case Male:
		return names.Xim.Male
	case Female:
		return names.Xim.Female
	default:
		return names.Xim.None
	}
}

// Xis returns the possesive pronoun for the gender (e.g. his)
func (s Gender) Xis() string {
	switch s {
	case Male:
		return names.Xis.Male
	case Female:
		return names.Xis.Female
	default:
		return names.Xis.None
	}
}
