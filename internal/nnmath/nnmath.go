package nnmath

import (
	"fmt"
	"math"
)

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func Cost(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0.0, fmt.Errorf("got input with length %d, but expected %d", len(x), len(y))
	}

	sum := 0.0
	for i := range x {
		diff := x[i] - y[i]
		sum += (diff * diff) / 2
	}

	return sum, nil
}
