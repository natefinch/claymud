package config

type Sex int

const (
	SEX_NONE Sex = iota
	SEX_MALE
	SEX_FEMALE
)

// Xself returns the reflexive pronoun for the sex (e.g. himself)
func (sex Sex) Xself() string {
	// TODO: make these configurable
	switch sex {
	case SEX_MALE:
		return "himself"
	case SEX_FEMALE:
		return "herself"
	}
	return "itself"
}

// TODO: he/him/etc
