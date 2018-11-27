package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// Dice represents a randomized value based on a number of dice, e.g. 3d6+5.
type Dice struct {
	Count, Size, Modifier int
}

// Roll rolls the dice and returns the result.
func (d Dice) Roll() int {
	result := 0
	for i := 0; i < d.Count; i++ {
		result += rand.Intn(d.Size) + 1
	}
	result += d.Modifier
	return result
}

// MakeDice converts an #d#+# formatted string to a Dice struct.
func MakeDice(s string) (Dice, error) {
	vals := strings.Split(s, "d")
	if len(vals) != 2 {
		return Dice{}, fmt.Errorf("expected dice to be #d#+-# but was %v", s)
	}
	count, err := strconv.Atoi(vals[0])
	if err != nil {
		return Dice{}, fmt.Errorf("count of dice is not a number: %v", vals[0])
	}
	sep := strings.IndexAny(vals[1], "+-")
	if sep == -1 {
		size, err := strconv.Atoi(vals[1])
		if err != nil {
			return Dice{}, fmt.Errorf("size of dice is not a number: %v", vals[1])
		}
		return Dice{Count: count, Size: size}, nil
	}
	s = vals[1][:sep]
	size, err := strconv.Atoi(s)
	if err != nil {
		return Dice{}, fmt.Errorf("size of dice is not a number: %v", s)
	}
	s = vals[1][sep+1:]
	mod, err := strconv.Atoi(s)
	if err != nil {
		return Dice{}, fmt.Errorf("modifier on dice is not a number: %v", s)
	}
	return Dice{Count: count, Size: size, Modifier: mod}, nil
}
