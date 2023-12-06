package utils

import (
	"math/rand"
)

func RandFloat(min, max float64) float64 {
	if max < min {
		return 0
	}
	//nolint:gosec // need a float64 random number
	return min + rand.Float64()*(max-min)
}
