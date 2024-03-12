package randfloat

import (
	"math/rand"
)

// RandFloat returns a random number in the specified range between the minimum and maximum.
func RandFloat(min, max float64) float64 {
	if max < min {
		return 0
	}
	//nolint:gosec // need a float64 random number
	return min + rand.Float64()*(max-min)
}
