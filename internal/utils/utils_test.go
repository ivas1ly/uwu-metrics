package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandFloat(t *testing.T) {
	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{
			name: "min 10, max 100",
			min:  10,
			max:  100,
		},
		{
			name: "min 25, max 1000",
			min:  25,
			max:  1000,
		},
		{
			name: "min 1000, max -1000",
			min:  1000,
			max:  -1000,
		},
		{
			name: "min -1345, max 7856",
			min:  -1345,
			max:  7856,
		},
	}

	var prev float64
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			randFloat := RandFloat(test.min, test.max)
			fmt.Println(randFloat)
			assert.NotEqual(t, prev, randFloat)
			prev = randFloat
			if test.max < test.min {
				assert.Equal(t, 0.0, randFloat)
				return
			}
			assert.GreaterOrEqual(t, randFloat, test.min)
			assert.LessOrEqual(t, randFloat, test.max)
		})
	}
}
