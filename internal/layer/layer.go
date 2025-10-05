package layer

import (
	"fmt"
	"math/rand/v2"
	"san-nn/internal/nnmath"
)

type Layer struct {
	Weights [][]float64
	Output  []float64
}

func (l *Layer) GetOutput(input []float64) error {
	if len(l.Weights) == 0 {
		return fmt.Errorf("layer has no weights")
	}
	if len(input) != len(l.Weights[0]) {
		return fmt.Errorf("got input size %d but expected %d", len(input), len(l.Weights[0]))
	}

	for i, row := range l.Weights {
		for j, weight := range row {
			l.Output[i] += weight * input[j]
		}
		l.Output[i] = nnmath.Sigmoid(l.Output[i])
	}
	return nil
}

func NewLayer(inputSize, outputSize int) Layer {
	l := Layer{
		Weights: make([][]float64, outputSize),
		Output:  make([]float64, outputSize),
	}
	for i := range l.Weights {
		l.Weights[i] = make([]float64, inputSize)
	}
	return l
}

func (l *Layer) InitWeightsRand() {
	for i, row := range l.Weights {
		for j := range row {
			l.Weights[i][j] = rand.Float64() + 1.0
		}
	}
}
