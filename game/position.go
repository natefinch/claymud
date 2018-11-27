package game

// Position represents the physical state of a mob or player.
type Position int

// All the positions
const (
	PositionDead Position = iota
	PositionMortallyWounded
	PositionIncapacitated
	PositionStunned
	PositionSleeping
	PositionResting
	PositionSitting
	PositionFighting
	PositionStanding
)
